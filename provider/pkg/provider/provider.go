// Copyright 2016-2026, Pulumi Corporation.
//
// Package provider is the gRPC resource-provider server for the
// OpenAPI-driven Pulumi Service Provider v2. It embeds the generated
// schema.json and metadata.json at build time, and routes every CRUD RPC
// through runtime.Dispatcher. No per-resource Go code — if a resource
// can't be expressed in the metadata, the right answer is to extend the
// metadata schema, not to add an escape hatch here.

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	pbempty "google.golang.org/protobuf/types/known/emptypb"
	structpb "google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/runtime"
)

// Server implements pulumirpc.ResourceProviderServer for the v2 provider.
// Embed UnimplementedResourceProviderServer so unimplemented RPCs (Construct,
// Invoke on unknown tokens, etc.) return a consistent "not implemented" error.
type Server struct {
	pulumirpc.UnimplementedResourceProviderServer

	name     string
	version  string
	schema   []byte // the embedded schema.json, returned by GetSchema verbatim
	metadata *runtime.CloudAPIMetadata

	// Populated lazily when Configure arrives.
	dispatcher *runtime.Dispatcher
	client     *runtime.Client
	config     config
}

// config holds the provider-level settings passed by the user.
type config struct {
	AccessToken string
	APIURL      string
}

// New constructs a v2 Server. `schemaBytes` and `metadataBytes` are the
// generator's output — typically embedded via //go:embed in the binary's
// main package.
func New(name, version string, schemaBytes, metadataBytes []byte) (*Server, error) {
	var md runtime.CloudAPIMetadata
	if err := json.Unmarshal(metadataBytes, &md); err != nil {
		return nil, fmt.Errorf("loading embedded metadata.json: %w", err)
	}
	return &Server{
		name:     name,
		version:  version,
		schema:   schemaBytes,
		metadata: &md,
	}, nil
}

// ─── Configuration ──────────────────────────────────────────────────────

// CheckConfig validates provider-level config. We accept anything and
// return it unchanged — Configure does the real work.
func (s *Server) CheckConfig(ctx context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.GetNews()}, nil
}

// DiffConfig compares provider configs. The v1 provider treats everything
// as non-replacing (changing the token or URL doesn't destroy resources);
// we mirror that.
func (s *Server) DiffConfig(ctx context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return &pulumirpc.DiffResponse{Changes: pulumirpc.DiffResponse_DIFF_NONE}, nil
}

// Configure stashes provider credentials and builds the HTTP client +
// dispatcher that every subsequent CRUD call uses.
func (s *Server) Configure(ctx context.Context, req *pulumirpc.ConfigureRequest) (*pulumirpc.ConfigureResponse, error) {
	args := req.GetArgs()
	if args != nil {
		if v, ok := args.GetFields()["accessToken"]; ok {
			s.config.AccessToken = v.GetStringValue()
		}
		if v, ok := args.GetFields()["apiUrl"]; ok {
			s.config.APIURL = v.GetStringValue()
		}
	}
	s.client = runtime.NewClient(s.config.APIURL, s.config.AccessToken)
	s.dispatcher = runtime.NewDispatcher(s.client, s.metadata)
	return &pulumirpc.ConfigureResponse{
		AcceptSecrets:   true,
		AcceptResources: true,
	}, nil
}

// ─── Schema ─────────────────────────────────────────────────────────────

// GetSchema returns the embedded Pulumi schema. The engine calls this
// to drive SDK-side type validation and code generation.
func (s *Server) GetSchema(ctx context.Context, req *pulumirpc.GetSchemaRequest) (*pulumirpc.GetSchemaResponse, error) {
	return &pulumirpc.GetSchemaResponse{Schema: string(s.schema)}, nil
}

// GetPluginInfo reports the version. Engines use it for compatibility checks.
func (s *Server) GetPluginInfo(ctx context.Context, _ *pbempty.Empty) (*pulumirpc.PluginInfo, error) {
	return &pulumirpc.PluginInfo{Version: s.version}, nil
}

// Cancel is a no-op for this provider — we don't hold long-running work
// that needs graceful shutdown.
func (s *Server) Cancel(ctx context.Context, _ *pbempty.Empty) (*pbempty.Empty, error) {
	return &pbempty.Empty{}, nil
}

// ─── CRUD ───────────────────────────────────────────────────────────────

// Check validates + normalizes inputs. Declarative Checks (requireOneOf,
// requireTogether, requireIf) from the resource's metadata are evaluated
// here, and any property with `sortOnRead: true` is canonicalized into
// its sorted form so subsequent diffs against the server's (also
// canonicalized) response don't flag spurious ordering changes.
func (s *Server) Check(ctx context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	token := tokenFromURN(req.GetUrn())
	res, found := s.metadata.Resources[token]
	if !found {
		return &pulumirpc.CheckResponse{Inputs: req.GetNews()}, nil
	}
	if len(res.Checks) == 0 && !runtime.HasSortOnRead(res.Properties) {
		return &pulumirpc.CheckResponse{Inputs: req.GetNews()}, nil
	}
	news, err := propertiesFromStruct(req.GetNews())
	if err != nil {
		return nil, err
	}
	news = runtime.CanonicalizeSortedInputs(res.Properties, news)
	failures := runtime.EvaluateChecks(&res, news)
	return makeCheckResponse(news, failures)
}

// Diff reports whether a change is needed and which properties require
// replacement. Any property in ForceNew that changed triggers replacement;
// anything else is an in-place update.
func (s *Server) Diff(ctx context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return s.genericDiff(tokenFromURN(req.GetUrn()), req)
}

// Create dispatches through runtime.Dispatcher.
func (s *Server) Create(ctx context.Context, req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	if s.dispatcher == nil {
		return nil, fmt.Errorf("provider not configured; Configure must precede Create")
	}
	news, err := propertiesFromStruct(req.GetProperties())
	if err != nil {
		return nil, err
	}
	id, outputs, err := s.dispatcher.Create(ctx, tokenFromURN(req.GetUrn()), news)
	if err != nil {
		return nil, err
	}
	return makeCreateResponse(id, outputs)
}

// Read refreshes the resource's state. Empty outputs (nil map) signal
// "resource no longer exists" — Pulumi will recreate on the next up.
func (s *Server) Read(ctx context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	if s.dispatcher == nil {
		return nil, fmt.Errorf("provider not configured; Configure must precede Read")
	}
	olds, err := propertiesFromStruct(req.GetProperties())
	if err != nil {
		return nil, err
	}
	id, outputs, err := s.dispatcher.Read(ctx, tokenFromURN(req.GetUrn()), req.GetId(), olds)
	if err != nil {
		return nil, err
	}
	if outputs == nil {
		// Resource no longer exists — Pulumi signals this via an empty ID.
		return &pulumirpc.ReadResponse{}, nil
	}
	return makeReadResponse(id, outputs, olds)
}

// Update applies an in-place change.
func (s *Server) Update(ctx context.Context, req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	if s.dispatcher == nil {
		return nil, fmt.Errorf("provider not configured; Configure must precede Update")
	}
	olds, err := propertiesFromStruct(req.GetOlds())
	if err != nil {
		return nil, err
	}
	news, err := propertiesFromStruct(req.GetNews())
	if err != nil {
		return nil, err
	}
	outputs, err := s.dispatcher.Update(ctx, tokenFromURN(req.GetUrn()), req.GetId(), olds, news)
	if err != nil {
		return nil, err
	}
	return makeUpdateResponse(outputs)
}

// Delete removes the resource from the backing Pulumi Cloud API.
func (s *Server) Delete(ctx context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	if s.dispatcher == nil {
		return nil, fmt.Errorf("provider not configured; Configure must precede Delete")
	}
	state, err := propertiesFromStruct(req.GetProperties())
	if err != nil {
		return nil, err
	}
	if err := s.dispatcher.Delete(ctx, tokenFromURN(req.GetUrn()), req.GetId(), state); err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

// Invoke runs a data-source function. The tok identifies which
// CloudAPIFunction in the metadata to call; args map to its path/query
// parameters. Resource method calls (Call) are not yet implemented —
// CloudAPIMethod entries are emitted into schema + metadata but the
// Call RPC is deferred to a subsequent 2.x.
func (s *Server) Invoke(ctx context.Context, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	if s.dispatcher == nil {
		return nil, fmt.Errorf("provider not configured; Configure must precede Invoke")
	}
	args, err := propertiesFromStruct(req.GetArgs())
	if err != nil {
		return nil, err
	}
	out, err := s.dispatcher.Invoke(ctx, req.GetTok(), args)
	if err != nil {
		return nil, err
	}
	ret, err := propertiesToStruct(out)
	if err != nil {
		return nil, err
	}
	return &pulumirpc.InvokeResponse{Return: ret}, nil
}

// ─── Helpers ────────────────────────────────────────────────────────────

// genericDiff compares old vs new property maps using ForceNew from metadata.
// Any property in ForceNew that changed triggers replacement; any property
// that differs at all is a diff.
func (s *Server) genericDiff(token string, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	res, ok := s.metadata.Resources[token]
	if !ok {
		// Unknown resource — no diff, no change.
		return &pulumirpc.DiffResponse{Changes: pulumirpc.DiffResponse_DIFF_NONE}, nil
	}
	olds, err := propertiesFromStruct(req.GetOlds())
	if err != nil {
		return nil, err
	}
	news, err := propertiesFromStruct(req.GetNews())
	if err != nil {
		return nil, err
	}

	forceNew := map[string]bool{}
	for _, f := range res.ForceNew {
		forceNew[f] = true
	}
	var diffs, replaces []string
	changes := pulumirpc.DiffResponse_DIFF_NONE
	// Walk the union of keys in olds and news; ignore output-only fields
	// (they're server-assigned and not meaningful for user-driven diffs).
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
		diffs = append(diffs, name)
		changes = pulumirpc.DiffResponse_DIFF_SOME
		if forceNew[name] {
			replaces = append(replaces, name)
		}
	}
	for k := range olds {
		consider(k)
	}
	for k := range news {
		consider(k)
	}
	return &pulumirpc.DiffResponse{
		Changes:  changes,
		Diffs:    diffs,
		Replaces: replaces,
	}, nil
}

// tokenFromURN extracts the resource-type token (e.g.
// "pulumiservice:orgs/agents:AgentPool") from a URN.
func tokenFromURN(urn string) string {
	// URN format: urn:pulumi:<stack>::<project>::<type>::<name>
	// The type token is the 3rd "::"-delimited segment from the right.
	parts := splitURN(urn)
	if len(parts) < 2 {
		return urn
	}
	// parts[len-2] is the type token, parts[len-1] is the name.
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

// propertiesFromStruct converts a protobuf Struct into a Pulumi PropertyMap.
// Uses the SDK's plugin helper so secret/computed markers round-trip correctly.
func propertiesFromStruct(s *structpb.Struct) (resource.PropertyMap, error) {
	if s == nil {
		return resource.PropertyMap{}, nil
	}
	return plugin.UnmarshalProperties(s, plugin.MarshalOptions{
		KeepUnknowns:  true,
		KeepSecrets:   true,
		KeepResources: true,
	})
}

// propertiesToStruct is the inverse — used when building response messages.
func propertiesToStruct(props resource.PropertyMap) (*structpb.Struct, error) {
	return plugin.MarshalProperties(props, plugin.MarshalOptions{
		KeepUnknowns:  true,
		KeepSecrets:   true,
		KeepResources: true,
	})
}

func makeCheckResponse(news resource.PropertyMap, failures []runtime.CheckFailure) (*pulumirpc.CheckResponse, error) {
	s, err := propertiesToStruct(news)
	if err != nil {
		return nil, err
	}
	resp := &pulumirpc.CheckResponse{Inputs: s}
	for _, f := range failures {
		resp.Failures = append(resp.Failures, &pulumirpc.CheckFailure{
			Property: f.Property,
			Reason:   f.Reason,
		})
	}
	return resp, nil
}

func makeCreateResponse(id string, outputs resource.PropertyMap) (*pulumirpc.CreateResponse, error) {
	s, err := propertiesToStruct(outputs)
	if err != nil {
		return nil, err
	}
	return &pulumirpc.CreateResponse{Id: id, Properties: s}, nil
}

func makeReadResponse(id string, outputs, priorInputs resource.PropertyMap) (*pulumirpc.ReadResponse, error) {
	outS, err := propertiesToStruct(outputs)
	if err != nil {
		return nil, err
	}
	// Inputs: a Read may produce changes Pulumi surfaces back to the program.
	// For MVP we pass the prior inputs through unchanged; richer import
	// support (e.g., deriving inputs from outputs) comes later.
	inputsS, err := propertiesToStruct(priorInputs)
	if err != nil {
		return nil, err
	}
	return &pulumirpc.ReadResponse{Id: id, Properties: outS, Inputs: inputsS}, nil
}

func makeUpdateResponse(outputs resource.PropertyMap) (*pulumirpc.UpdateResponse, error) {
	s, err := propertiesToStruct(outputs)
	if err != nil {
		return nil, err
	}
	return &pulumirpc.UpdateResponse{Properties: s}, nil
}

