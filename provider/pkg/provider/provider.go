// Copyright 2016-2026, Pulumi Corporation.
//
// Package provider hosts the Pulumi Service Provider v2 on top of
// github.com/pulumi/pulumi-go-provider. It builds a Provider literal
// whose function fields close over a metadata-driven runtime
// dispatcher: every CRUD/Invoke RPC looks up the resource (or
// function) in the metadata derived from resource-map.yaml, splits
// inputs into path/query/body, and forwards to Pulumi Cloud over
// HTTP. There is no per-resource Go code — the metadata schema is the
// extension surface.
//
// Compared to v2's earlier raw-gRPC server, this layout:
//   - Lets us reuse pulumi-go-provider's request marshaling, cancel,
//     and Parameterize plumbing for free.
//   - Defers schema generation to the first GetSchema call (lazy),
//     so a freshly-built binary can serve CRUD before anyone asks
//     for the schema.
//   - Surfaces unmapped operationIds as an error from GetSchema,
//     per iwahbe's design note that the runtime should fail loudly
//     when the map is incomplete.

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	pgo "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/gen"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/runtime"
)

// New builds a v2 provider Provider literal from the embedded OpenAPI
// spec and resource-map bytes. Metadata is parsed eagerly so that
// CRUD lookups don't pay a per-call generation cost; the Pulumi
// schema is generated lazily on first GetSchema (it's bigger and
// only needed when the engine asks).
//
// Iwahbe's design note: GetSchema must error if the map is incomplete
// or incorrect. We keep the strict check there (it's the gate the
// engine actually trips when the user runs `pulumi up` against a
// freshly-refreshed spec) and the test-time variant lives in
// embedded_test.go's TestEmbedded_CoverageGate.
func New(specBytes, mapBytes []byte) (pgo.Provider, error) {
	if len(specBytes) == 0 {
		return pgo.Provider{}, errors.New("empty OpenAPI spec bytes")
	}
	if len(mapBytes) == 0 {
		return pgo.Provider{}, errors.New("empty resource-map bytes")
	}

	metadataBytes, err := gen.EmitMetadataFromBytes(specBytes, mapBytes)
	if err != nil {
		return pgo.Provider{}, fmt.Errorf("emitting runtime metadata: %w", err)
	}
	var md runtime.CloudAPIMetadata
	if err := json.Unmarshal(metadataBytes, &md); err != nil {
		return pgo.Provider{}, fmt.Errorf("decoding runtime metadata: %w", err)
	}

	st := &state{
		spec:     specBytes,
		mapBytes: mapBytes,
		metadata: &md,
	}

	return pgo.Provider{
		GetSchema:    st.getSchema,
		CheckConfig:  st.checkConfig,
		DiffConfig:   st.diffConfig,
		Configure:    st.configure,
		Check:        st.check,
		Diff:         st.diff,
		Create:       st.create,
		Read:         st.read,
		Update:       st.update,
		Delete:       st.delete,
		Invoke:       st.invoke,
		Parameterize: st.parameterize,
		Cancel:       st.cancel,
	}, nil
}

// ─── Provider config (CheckConfig / DiffConfig / Cancel) ─────────────────

// checkConfig accepts any config bag the engine hands us. The two
// configuration variables the provider supports — accessToken and
// apiUrl — are validated implicitly by Configure (the API client
// fails to authenticate if the token is wrong, the user sees a clear
// error). Returning the inputs verbatim with no failures matches v1
// semantics. The framework's default would return Unimplemented;
// explicit is safer.
func (s *state) checkConfig(_ context.Context, req pgo.CheckRequest) (pgo.CheckResponse, error) {
	return pgo.CheckResponse{Inputs: req.Inputs}, nil
}

// diffConfig is invariant under any change: we treat config as
// non-replacing (changing the token, swapping the API URL, etc. do
// not destroy resources — they just affect how subsequent CRUD
// calls authenticate). v1 had the same behavior.
func (s *state) diffConfig(_ context.Context, _ pgo.DiffRequest) (pgo.DiffResponse, error) {
	return pgo.DiffResponse{HasChanges: false}, nil
}

// cancel is a no-op. We don't hold long-running work; CRUD calls
// are individual HTTP requests against Pulumi Cloud, and the engine
// already cancels the context when the user aborts a run.
func (s *state) cancel(_ context.Context) error {
	return nil
}

// ─── State ──────────────────────────────────────────────────────────────

// state holds the per-process resources the function-field callbacks
// close over: the embedded inputs (so GetSchema can regenerate
// lazily), the parsed metadata (used by every CRUD call), and the
// HTTP client + dispatcher built at Configure time.
type state struct {
	spec     []byte
	mapBytes []byte
	metadata *runtime.CloudAPIMetadata

	// Schema is computed once on first GetSchema and reused for every
	// subsequent call. The error is sticky: if generation fails (e.g.
	// the map is inconsistent), every GetSchema returns the same
	// failure rather than re-running the (expensive) emission.
	schemaOnce sync.Once
	schemaJSON []byte
	schemaErr  error

	// Configure populates these. They're read concurrently across
	// CRUD callbacks; we don't take a lock because the engine
	// guarantees a happens-before edge between Configure and the
	// first CRUD call, and they're never mutated afterwards.
	cfg        config
	client     *runtime.Client
	dispatcher *runtime.Dispatcher
}

// config carries the provider-level settings the user supplied via
// stack config or constructor args. Mirrors the schema's `config`
// block.
type config struct {
	AccessToken string
	APIURL      string
}

// ─── Schema ─────────────────────────────────────────────────────────────

// getSchema runs gen.EmitSchemaFromBytes on the embedded inputs,
// then verifies that every operationId in the spec is claimed by the
// map. An unmapped operation is a "schema is incomplete" condition
// and we surface it as a structured error so the engine doesn't try
// to advertise an SDK that lies about coverage.
func (s *state) getSchema(_ context.Context, _ pgo.GetSchemaRequest) (pgo.GetSchemaResponse, error) {
	s.schemaOnce.Do(func() {
		report, err := gen.CoverageReportFromBytes(s.spec, s.mapBytes)
		if err != nil {
			s.schemaErr = fmt.Errorf("running coverage gate: %w", err)
			return
		}
		if report.UnmappedCount > 0 {
			s.schemaErr = fmt.Errorf(
				"resource map is incomplete: %d operationId(s) in the OpenAPI"+
					" spec are neither mapped to a resource/function/method"+
					" nor explicitly excluded; first few: %s",
				report.UnmappedCount, firstUnmappedNames(report, 5))
			return
		}
		raw, err := gen.EmitSchemaFromBytes(s.spec, s.mapBytes)
		if err != nil {
			s.schemaErr = fmt.Errorf("emitting Pulumi schema: %w", err)
			return
		}
		s.schemaJSON = raw
	})
	if s.schemaErr != nil {
		return pgo.GetSchemaResponse{}, s.schemaErr
	}
	return pgo.GetSchemaResponse{Schema: string(s.schemaJSON)}, nil
}

func firstUnmappedNames(r *gen.Report, n int) string {
	names := make([]string, 0, n)
	for i, op := range r.Unmapped {
		if i >= n {
			break
		}
		names = append(names, op.OperationID)
	}
	return strings.Join(names, ", ")
}

// ─── Configure ──────────────────────────────────────────────────────────

// configure stashes the access token and API URL, then constructs
// the HTTP client + dispatcher every CRUD callback uses. Called once
// per provider lifecycle.
//
// The schema declares PULUMI_ACCESS_TOKEN / PULUMI_BACKEND_URL as
// env-var defaults, but the engine doesn't always materialize those
// into the Configure RPC args (in particular not for in-process
// integration harnesses), so we fall back to reading the env
// directly. This matches v1 behavior — programs and tests can set
// the env var instead of stack config and everything still works.
func (s *state) configure(_ context.Context, req pgo.ConfigureRequest) error {
	args := req.Args.AsMap()
	if v, ok := args["accessToken"]; ok && v.IsString() {
		s.cfg.AccessToken = v.AsString()
	}
	if v, ok := args["apiUrl"]; ok && v.IsString() {
		s.cfg.APIURL = v.AsString()
	}
	if s.cfg.AccessToken == "" {
		s.cfg.AccessToken = os.Getenv("PULUMI_ACCESS_TOKEN")
	}
	if s.cfg.APIURL == "" {
		s.cfg.APIURL = os.Getenv("PULUMI_BACKEND_URL")
	}
	s.client = runtime.NewClient(s.cfg.APIURL, s.cfg.AccessToken)
	s.dispatcher = runtime.NewDispatcher(s.client, s.metadata)
	return nil
}

// ─── CRUD ───────────────────────────────────────────────────────────────

// check evaluates declarative checks (requireOneOf / requireTogether
// / requireIfSet / requireIf) and canonicalizes any sortOnRead
// arrays so that subsequent diffs against the server's sorted
// response don't flag spurious ordering changes.
func (s *state) check(_ context.Context, req pgo.CheckRequest) (pgo.CheckResponse, error) {
	token := tokenFromURN(string(req.Urn))
	res, found := s.metadata.Resources[token]
	if !found {
		return pgo.CheckResponse{Inputs: req.Inputs}, nil
	}
	if len(res.Checks) == 0 && !runtime.HasSortOnRead(res.Properties) {
		return pgo.CheckResponse{Inputs: req.Inputs}, nil
	}
	news := resource.ToResourcePropertyMap(req.Inputs)
	news = runtime.CanonicalizeSortedInputs(res.Properties, news)
	failures := runtime.EvaluateChecks(&res, news)

	resp := pgo.CheckResponse{Inputs: resource.FromResourcePropertyMap(news)}
	for _, f := range failures {
		resp.Failures = append(resp.Failures, pgo.CheckFailure{
			Property: f.Property,
			Reason:   f.Reason,
		})
	}
	return resp, nil
}

// diff compares old vs new inputs using ForceNew from metadata. Any
// ForceNew property that changed becomes a replace; anything else
// that differs is an in-place update.
//
// Pulumi engines from v3.74.0+ send the prior inputs separately
// (OldInputs); compare against those rather than State (which is the
// prior outputs). The two diverge whenever the API echoes server-
// computed defaults — comparing inputs to inputs avoids spurious
// "the server added members:[]" diffs on every refresh. Older
// engines without OldInputs fall through to State as a best effort.
func (s *state) diff(_ context.Context, req pgo.DiffRequest) (pgo.DiffResponse, error) {
	token := tokenFromURN(string(req.Urn))
	res, ok := s.metadata.Resources[token]
	if !ok {
		return pgo.DiffResponse{}, nil
	}
	priorInputs := req.OldInputs
	if priorInputs.Len() == 0 {
		priorInputs = req.State
	}
	olds := resource.ToResourcePropertyMap(priorInputs)
	news := resource.ToResourcePropertyMap(req.Inputs)

	forceNew := map[string]bool{}
	for _, f := range res.ForceNew {
		forceNew[f] = true
	}
	detailed := map[string]pgo.PropertyDiff{}
	hasChanges := false
	seen := map[string]bool{}
	consider := func(key resource.PropertyKey) {
		name := string(key)
		if seen[name] {
			return
		}
		seen[name] = true
		if meta, has := res.Properties[name]; has && meta.Output {
			return
		}
		oldV, oldOk := olds[key]
		newV, newOk := news[key]
		if !oldOk && !newOk {
			return
		}
		if oldOk && newOk && oldV.DeepEquals(newV) {
			return
		}
		hasChanges = true
		kind := pgo.Update
		switch {
		case !oldOk && newOk:
			kind = pgo.Add
		case oldOk && !newOk:
			kind = pgo.Delete
		}
		if forceNew[name] {
			switch kind {
			case pgo.Add:
				kind = pgo.AddReplace
			case pgo.Delete:
				kind = pgo.DeleteReplace
			default:
				kind = pgo.UpdateReplace
			}
		}
		detailed[name] = pgo.PropertyDiff{Kind: kind}
	}
	for k := range olds {
		consider(k)
	}
	for k := range news {
		consider(k)
	}
	return pgo.DiffResponse{
		HasChanges:   hasChanges,
		DetailedDiff: detailed,
	}, nil
}

// create dispatches through runtime.Dispatcher.
func (s *state) create(ctx context.Context, req pgo.CreateRequest) (pgo.CreateResponse, error) {
	if s.dispatcher == nil {
		return pgo.CreateResponse{}, errors.New("provider not configured; Configure must precede Create")
	}
	news := resource.ToResourcePropertyMap(req.Properties)
	id, outputs, err := s.dispatcher.Create(ctx, tokenFromURN(string(req.Urn)), news)
	if err != nil {
		return pgo.CreateResponse{}, err
	}
	return pgo.CreateResponse{
		ID:         id,
		Properties: resource.FromResourcePropertyMap(outputs),
	}, nil
}

// read refreshes the resource. A nil outputs map signals "resource no
// longer exists" — the framework handles that via an empty ID in the
// response.
func (s *state) read(ctx context.Context, req pgo.ReadRequest) (pgo.ReadResponse, error) {
	if s.dispatcher == nil {
		return pgo.ReadResponse{}, errors.New("provider not configured; Configure must precede Read")
	}
	olds := resource.ToResourcePropertyMap(req.Properties)
	id, outputs, err := s.dispatcher.Read(ctx, tokenFromURN(string(req.Urn)), req.ID, olds)
	if err != nil {
		return pgo.ReadResponse{}, err
	}
	if outputs == nil {
		// Resource no longer exists.
		return pgo.ReadResponse{}, nil
	}
	// Inputs in the response are the user-program-shaped inputs the
	// engine compares against the program's current inputs on the
	// next diff. Use req.Inputs (prior inputs from the engine) when
	// they're present (refresh path); fall back to outputs only for
	// import (where the engine has no prior inputs to hand us).
	inputs := req.Inputs
	if inputs.Len() == 0 {
		inputs = resource.FromResourcePropertyMap(outputs)
	}
	return pgo.ReadResponse{
		ID:         id,
		Properties: resource.FromResourcePropertyMap(outputs),
		Inputs:     inputs,
	}, nil
}

// update applies an in-place change.
func (s *state) update(ctx context.Context, req pgo.UpdateRequest) (pgo.UpdateResponse, error) {
	if s.dispatcher == nil {
		return pgo.UpdateResponse{}, errors.New("provider not configured; Configure must precede Update")
	}
	olds := resource.ToResourcePropertyMap(req.State)
	news := resource.ToResourcePropertyMap(req.Inputs)
	outputs, err := s.dispatcher.Update(ctx, tokenFromURN(string(req.Urn)), req.ID, olds, news)
	if err != nil {
		return pgo.UpdateResponse{}, err
	}
	return pgo.UpdateResponse{Properties: resource.FromResourcePropertyMap(outputs)}, nil
}

// delete removes the resource.
func (s *state) delete(ctx context.Context, req pgo.DeleteRequest) error {
	if s.dispatcher == nil {
		return errors.New("provider not configured; Configure must precede Delete")
	}
	state := resource.ToResourcePropertyMap(req.Properties)
	return s.dispatcher.Delete(ctx, tokenFromURN(string(req.Urn)), req.ID, state)
}

// invoke runs a data-source function (e.g. listAccounts).
func (s *state) invoke(ctx context.Context, req pgo.InvokeRequest) (pgo.InvokeResponse, error) {
	if s.dispatcher == nil {
		return pgo.InvokeResponse{}, errors.New("provider not configured; Configure must precede Invoke")
	}
	args := resource.ToResourcePropertyMap(req.Args)
	out, err := s.dispatcher.Invoke(ctx, string(req.Token), args)
	if err != nil {
		return pgo.InvokeResponse{}, err
	}
	return pgo.InvokeResponse{Return: resource.FromResourcePropertyMap(out)}, nil
}

// parameterize is the hook iwahbe flagged as a future capability:
// a self-hosted customer could swap in a different OpenAPI spec /
// resource map at runtime. v2.0.0-alpha.1 ships the structural
// wiring (so callers don't get an "unimplemented at the framework
// layer" error) but doesn't yet rebuild metadata from the
// parameters; the Args/Value carry the bytes the next 2.x will
// use. Returning a typed error rather than a panic so a misbehaving
// consumer gets a clean failure.
func (s *state) parameterize(_ context.Context, _ pgo.ParameterizeRequest) (pgo.ParameterizeResponse, error) {
	return pgo.ParameterizeResponse{}, errors.New(
		"parameterize is not yet implemented in pulumiservice v2.0.0-alpha.1;" +
			" the wiring is in place for a follow-up to swap the OpenAPI spec /" +
			" resource map at runtime")
}

// ─── Helpers ────────────────────────────────────────────────────────────

// tokenFromURN extracts the resource-type token (e.g.
// "pulumiservice:orgs/agents:AgentPool") from a URN.
func tokenFromURN(urn string) string {
	// URN format: urn:pulumi:<stack>::<project>::<type>::<name>
	// The type token is the second-to-last "::"-delimited segment.
	parts := splitURN(urn)
	if len(parts) < 2 {
		return urn
	}
	return parts[len(parts)-2]
}

func splitURN(urn string) []string {
	var out []string
	cur := 0
	for i := 0; i < len(urn)-1; i++ {
		if urn[i] == ':' && urn[i+1] == ':' {
			out = append(out, urn[cur:i])
			cur = i + 2
			i++
		}
	}
	out = append(out, urn[cur:])
	return out
}

// Compile-time sanity check: property package is imported (some tests
// reference it). Without this, `goimports` may strip the import.
var _ = property.Map{}
