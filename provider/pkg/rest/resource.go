// Copyright 2016-2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rest

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

// pathParamRE matches OpenAPI {param} placeholders in path strings.
var pathParamRE = regexp.MustCompile(`\{([^/{}]+)\}`)

// HTTPError carries the response status and body for a non-2xx upstream
// response. Callers branch on status via errors.As.
type HTTPError struct {
	Method     string
	URL        string
	StatusCode int
	Body       []byte
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("rest: %s %s returned %d: %s", e.Method, e.URL, e.StatusCode, strings.TrimSpace(string(e.Body)))
}

// IsNotFound reports whether err is an HTTPError with status 404.
func IsNotFound(err error) bool {
	var herr *HTTPError
	return errors.As(err, &herr) && herr.StatusCode == http.StatusNotFound
}

// MissingPathParamError signals that buildURL couldn't substitute a {path}
// placeholder because the named input was absent. Surfaced as its own type so
// callers like checkAlreadyExists can distinguish "user typo / unresolvable
// reference" from "the Read URL needs a server-generated ID we don't have yet."
type MissingPathParamError struct {
	WireName, PulumiName string
}

func (e *MissingPathParamError) Error() string {
	return fmt.Sprintf("rest: path parameter %q (Pulumi name %q) missing from inputs", e.WireName, e.PulumiName)
}

// Resources builds one handler per metadata.resources entry. Operation IDs
// resolve lazily at call time, so a half-broken metadata document doesn't
// fail the running provider — broken mappings surface via BuildSchema or
// at the first CRUD call.
func Resources(spec *Spec, metadata *Metadata) map[string]*Resource {
	out := make(map[string]*Resource, len(metadata.Resources))
	for key, rm := range metadata.Resources {
		token := key
		if rm.Token != "" {
			token = rm.Token
		}
		if rm.Operations.Create != "" && rm.Operations.Create == rm.Operations.Update && !rm.RequireImport {
			fmt.Fprintf(os.Stderr,
				"rest: %s has create==update operationId %q but requireImport is unset; re-run scaffold-metadata\n",
				token, rm.Operations.Create)
		}
		out[token] = &Resource{meta: rm, spec: spec}
	}
	return out
}

// Resource is a metadata-driven CRUD handler. Operation IDs resolve at
// call time, not at construction.
type Resource struct {
	meta ResourceMeta
	spec *Spec
}

// resolveOp looks up an operation ID. Returns (nil, nil) when id is empty.
func (r *Resource) resolveOp(verb, id string) (*Operation, error) {
	if id == "" {
		return nil, nil
	}
	op, ok := r.spec.Op(id)
	if !ok {
		return nil, fmt.Errorf("rest: operations.%s = %q not found in spec", verb, id)
	}
	return op, nil
}

// Check normalizes user inputs to suppress spurious diffs: enum case-folding,
// set-like array sorting (Unordered), and autoName generation.
func (r *Resource) Check(_ context.Context, req p.CheckRequest) (p.CheckResponse, error) {
	if r.meta.Attachment != nil {
		// Attachments have no create op to normalize against; their edge inputs
		// pass through untouched (enum folding/autoname don't apply).
		return p.CheckResponse{Inputs: req.Inputs}, nil
	}
	op, _ := r.spec.Op(r.meta.Operations.Create)
	if op == nil {
		return p.CheckResponse{Inputs: req.Inputs}, nil
	}
	bodyProps := flattenedRequestProperties(r.spec, op)

	out := map[string]property.Value{}
	for k, v := range req.Inputs.AllStable {
		out[k] = normalizeValue(v, k, bodyProps, r.meta)
	}
	for name, fm := range r.meta.Fields {
		if fm.AutoName <= 0 {
			continue
		}
		if _, ok := out[name]; ok {
			continue
		}
		out[name] = property.New(generateAutoName(string(req.Urn), req.RandomSeed, fm.AutoName))
	}
	return p.CheckResponse{Inputs: property.NewMap(out)}, nil
}

// normalizeValue canonicalizes one input: enum case-fold for strings,
// sort for Unordered arrays. Other shapes pass through.
func normalizeValue(v property.Value, pulumiName string, bodyProps map[string]any, meta ResourceMeta) property.Value {
	wireName := wireSideName(pulumiName, meta.Renames)
	prop, ok := bodyProps[wireName].(map[string]any)
	if !ok {
		return v
	}
	if v.IsString() {
		if canonical, ok := matchEnumCase(v.AsString(), prop); ok {
			return property.New(canonical)
		}
	}
	if v.IsArray() && meta.Fields[pulumiName].Unordered {
		return property.New(sortArrayValue(v.AsArray()))
	}
	return v
}

// matchEnumCase returns the canonical enum value matching s case-insensitively,
// or ("", false) when the property has no enum or no match.
func matchEnumCase(s string, prop map[string]any) (string, bool) {
	rawEnum, ok := prop["enum"].([]any)
	if !ok {
		return "", false
	}
	for _, e := range rawEnum {
		es, ok := e.(string)
		if !ok {
			continue
		}
		if strings.EqualFold(s, es) {
			return es, true
		}
	}
	return "", false
}

// sortArrayValue stable-sorts an array of strings. Mixed-type or
// nested-object arrays pass through (sort semantics ill-defined).
func sortArrayValue(arr property.Array) property.Array {
	values := arr.AsSlice()
	if len(values) < 2 {
		return arr
	}
	for _, v := range values {
		if !v.IsString() {
			return arr
		}
	}
	sorted := make([]property.Value, len(values))
	copy(sorted, values)
	slices.SortFunc(sorted, func(a, b property.Value) int {
		return strings.Compare(a.AsString(), b.AsString())
	})
	return property.NewArray(sorted)
}

// generateAutoName produces a deterministic name from the URN and the
// engine-supplied random seed, capped at maxLen. Falls back to crypto/rand
// when seed is nil.
func generateAutoName(urn string, seed []byte, maxLen int) string {
	parts := strings.Split(urn, "::")
	base := parts[len(parts)-1]
	if base == "" {
		base = "resource"
	}
	const suffixLen = 7
	var suffix string
	if len(seed) > 0 {
		h := sha256.Sum256(seed)
		suffix = hex.EncodeToString(h[:4])[:suffixLen]
	} else {
		buf := make([]byte, 4)
		_, _ = rand.Read(buf)
		suffix = hex.EncodeToString(buf)[:suffixLen]
	}
	candidate := base + "-" + suffix
	if maxLen > 0 && len(candidate) > maxLen {
		budget := maxLen - suffixLen - 1
		if budget < 1 {
			return candidate[:maxLen]
		}
		candidate = base[:budget] + "-" + suffix
	}
	return candidate
}

// wireSideName inverts the renames map: Pulumi-side → wire-side.
func wireSideName(pulumiName string, renames map[string]string) string {
	if wire, ok := renames[pulumiName]; ok {
		return wire
	}
	return pulumiName
}

// flattenedRequestProperties returns the op's request body properties,
// resolving $refs and allOf. Returns nil when the op has no body.
func flattenedRequestProperties(spec *Spec, op *Operation) map[string]any {
	if op == nil || op.RequestRef == "" {
		return nil
	}
	props, _, err := flattenObjectSchema(spec, op.RequestRef)
	if err != nil {
		return nil
	}
	return props
}

// Diff classifies each changed input as Update or UpdateReplace. Path
// params and forceNew fields trigger replacement. Without an explicit
// DetailedDiff the engine never triggers replace, so the replace semantics
// must be spelled out here.
func (r *Resource) Diff(_ context.Context, req p.DiffRequest) (p.DiffResponse, error) {
	if r.meta.Attachment != nil {
		return r.diffAttachment(req), nil
	}
	if mapEqual(req.OldInputs, req.Inputs) {
		return p.DiffResponse{}, nil
	}
	replaces := r.replaceTriggeringFields()
	detailed := map[string]p.PropertyDiff{}
	for k, newV := range req.Inputs.AllStable {
		oldV, ok := req.OldInputs.GetOk(k)
		if !ok {
			detailed[k] = p.PropertyDiff{Kind: addKind(replaces[k])}
			continue
		}
		if !newV.Equals(oldV) {
			detailed[k] = p.PropertyDiff{Kind: updateKind(replaces[k])}
		}
	}
	for k := range req.OldInputs.AllStable {
		if _, ok := req.Inputs.GetOk(k); !ok {
			detailed[k] = p.PropertyDiff{Kind: deleteKind(replaces[k])}
		}
	}
	return p.DiffResponse{
		HasChanges:          true,
		DeleteBeforeReplace: r.meta.DeleteBeforeReplace,
		DetailedDiff:        detailed,
	}, nil
}

// replaceTriggeringFields returns input names whose changes force a replace:
// every op's path params plus any FieldMeta.ForceNew fields.
func (r *Resource) replaceTriggeringFields() map[string]bool {
	out := map[string]bool{}
	ops := []string{
		r.meta.Operations.Create, r.meta.Operations.Read,
		r.meta.Operations.Update, r.meta.Operations.Delete,
	}
	for _, opID := range ops {
		op, ok := r.spec.Op(opID)
		if !ok {
			continue
		}
		for _, pp := range op.Parameters {
			if pp.In == inPath {
				out[pulumiName(pp.Name, r.meta.Renames)] = true
			}
		}
	}
	for name, fm := range r.meta.Fields {
		if fm.ForceNew {
			out[name] = true
		}
	}
	return out
}

func addKind(replace bool) p.DiffKind {
	if replace {
		return p.AddReplace
	}
	return p.Add
}

func updateKind(replace bool) p.DiffKind {
	if replace {
		return p.UpdateReplace
	}
	return p.Update
}

func deleteKind(replace bool) p.DiffKind {
	if replace {
		return p.DeleteReplace
	}
	return p.Delete
}

// Create executes the create operation, then fires the read op (when
// declared) and merges its response in. Many Pulumi Cloud create endpoints
// return a sparse body, so without read-after-create downstream resources
// referencing read-only outputs would fail to converge until refresh.
func (r *Resource) Create(ctx context.Context, req p.CreateRequest) (p.CreateResponse, error) {
	if r.meta.Attachment != nil {
		return r.createAttachment(ctx, req)
	}
	if req.DryRun {
		return p.CreateResponse{Properties: req.Properties}, nil
	}
	op, err := r.resolveOp("create", r.meta.Operations.Create)
	if err != nil {
		return p.CreateResponse{}, err
	}
	if op == nil {
		return p.CreateResponse{}, fmt.Errorf("create: resource has no create operation declared")
	}
	if r.meta.RequireImport {
		if err := r.checkAlreadyExists(ctx, req.Properties); err != nil {
			return p.CreateResponse{}, err
		}
	}
	_, state, err := r.execAndDecode(ctx, op, req.Properties)
	if err != nil {
		return p.CreateResponse{}, err
	}
	if updOp, _ := r.resolveOp("update", r.meta.Operations.Update); updOp != nil &&
		updOp.RequestContentType == contentYAML &&
		op.RequestContentType != contentYAML {
		if v, ok := req.Properties.GetOk("yaml"); ok && v.IsString() && v.AsString() != "" {
			_, updState, err := r.execAndDecode(ctx, updOp, req.Properties)
			if err != nil {
				return p.CreateResponse{}, fmt.Errorf("create: post-create yaml apply: %w", err)
			}
			state = mergeMaps(updState, state)
		}
	}
	// Path-parameter values must come from inputs, not state: create endpoints
	// often return sparse bodies that don't echo path params back.
	source := mergeMaps(req.Properties, state)
	if fetched, ok, err := r.fetchState(ctx, source, state); err != nil {
		return p.CreateResponse{}, fmt.Errorf("create: read-after-create: %w", err)
	} else if ok {
		state = fetched
	}
	id, err := r.synthesizeID(state, req.Properties)
	if err != nil {
		return p.CreateResponse{}, fmt.Errorf("create: %w", err)
	}
	return p.CreateResponse{ID: id, Properties: state}, nil
}

// checkAlreadyExists is the requireImport pre-flight: a 200 from read fails
// with an "import this resource" error; 404 means proceed; other errors
// propagate. Resources without a read op opt out.
func (r *Resource) checkAlreadyExists(ctx context.Context, inputs property.Map) error {
	readOp, err := r.resolveOp("read", r.meta.Operations.Read)
	if err != nil {
		return err
	}
	if readOp == nil {
		return nil
	}
	_, _, err = r.execAndDecode(ctx, readOp, inputs)
	if err == nil {
		return fmt.Errorf("resource already exists; use the `import` resource option " +
			"to bring it under Pulumi management instead of creating a new one")
	}
	if IsNotFound(err) {
		return nil
	}
	// Read URL needs a path param we can't supply pre-create (typically a
	// server-generated ID). The probe can't run, so surface a metadata-error
	// message instead of the generic "requireImport probe" wrap.
	var mpErr *MissingPathParamError
	if errors.As(err, &mpErr) {
		token := r.meta.Token
		if token == "" {
			token = "this resource"
		}
		return fmt.Errorf("rest: metadata error: %s has requireImport set but its read URL needs "+
			"path parameter %q (Pulumi name %q), which isn't in user-supplied inputs "+
			"(likely server-generated). Remove requireImport from metadata for this resource",
			token, mpErr.WireName, mpErr.PulumiName)
	}
	return fmt.Errorf("requireImport probe: %w", err)
}

// synthesizeID returns the resource ID, using meta.IDFormat when set or
// slash-joining path-parameter values from the most authoritative non-create
// op. Values resolve from state first (covers server-generated fields), then
// inputs (covers user-supplied fields not echoed in the response).
func (r *Resource) synthesizeID(state, inputs property.Map) (string, error) {
	if r.meta.IDFormat != "" {
		return synthesizeIDFromFormat(r.meta.IDFormat, state, inputs)
	}
	var op *Operation
	ops := []string{
		r.meta.Operations.Read, r.meta.Operations.Update,
		r.meta.Operations.Delete, r.meta.Operations.Create,
	}
	for _, opID := range ops {
		if opID == "" {
			continue
		}
		candidate, ok := r.spec.Op(opID)
		if !ok {
			continue
		}
		if hasPathParams(candidate) {
			op = candidate
			break
		}
	}
	if op == nil {
		return "", fmt.Errorf("no operation with path parameters available to synthesize ID")
	}
	var parts []string
	for _, pp := range op.Parameters {
		if pp.In != inPath {
			continue
		}
		pulName := pulumiName(pp.Name, r.meta.Renames)
		v, ok := state.GetOk(pulName)
		if !ok {
			v, ok = inputs.GetOk(pulName)
		}
		if !ok {
			return "", fmt.Errorf("path parameter %q (Pulumi name %q) missing from state and inputs", pp.Name, pulName)
		}
		parts = append(parts, propertyValueToString(v))
	}
	return strings.Join(parts, "/"), nil
}

// synthesizeIDFromFormat substitutes {name} placeholders from state then inputs.
func synthesizeIDFromFormat(format string, state, inputs property.Map) (string, error) {
	var missing []string
	out := pathParamRE.ReplaceAllStringFunc(format, func(m string) string {
		name := m[1 : len(m)-1]
		v, ok := state.GetOk(name)
		if !ok {
			v, ok = inputs.GetOk(name)
		}
		if !ok {
			missing = append(missing, name)
			return m
		}
		return propertyValueToString(v)
	})
	if len(missing) > 0 {
		return "", fmt.Errorf("idFormat %q: missing values for %v", format, missing)
	}
	return out, nil
}

// parseIDIntoInputs is the inverse of synthesizeIDFromFormat: it recovers
// placeholder values from the resource ID and merges them into inputs
// without overwriting existing keys. Returns inputs unchanged when no
// IDFormat is declared or the ID doesn't match.
func (r *Resource) parseIDIntoInputs(id string, inputs property.Map) property.Map {
	if r.meta.IDFormat == "" {
		return inputs
	}
	re, names, err := compileIDFormatRegex(r.meta.IDFormat)
	if err != nil {
		return inputs
	}
	matches := re.FindStringSubmatch(id)
	if matches == nil {
		return inputs
	}
	out := map[string]property.Value{}
	for k, v := range inputs.AllStable {
		out[k] = v
	}
	for i, name := range names {
		if i+1 >= len(matches) {
			break
		}
		if _, exists := out[name]; exists {
			continue
		}
		out[name] = property.New(matches[i+1])
	}
	return property.NewMap(out)
}

// compileIDFormatRegex turns "{org}/{name}" into ^([^/]+)/([^/]+)$, returning
// the regex and the placeholder names in match-group order.
func compileIDFormatRegex(format string) (*regexp.Regexp, []string, error) {
	var names []string
	var pattern strings.Builder
	pattern.WriteByte('^')
	last := 0
	for _, m := range pathParamRE.FindAllStringSubmatchIndex(format, -1) {
		pattern.WriteString(regexp.QuoteMeta(format[last:m[0]]))
		names = append(names, format[m[2]:m[3]])
		pattern.WriteString(`([^/]+)`)
		last = m[1]
	}
	pattern.WriteString(regexp.QuoteMeta(format[last:]))
	pattern.WriteByte('$')
	re, err := regexp.Compile(pattern.String())
	if err != nil {
		return nil, nil, err
	}
	return re, names, nil
}

func hasPathParams(op *Operation) bool {
	for _, p := range op.Parameters {
		if p.In == inPath {
			return true
		}
	}
	return false
}

// Read fetches current state. Imports arrive with empty inputs; when
// IDFormat is declared, path params are recovered from the resource ID
// for URL construction.
//
// The returned Inputs differ between import and refresh: on import
// (req.Inputs empty), we surface the parsed-ID values so the user gets
// a usable program-input reconstruction; on refresh we preserve the
// caller's existing Inputs verbatim. Otherwise the parsed-ID values —
// which the user never wrote in their program — show up as a diff on
// every refresh ("+issuerId" etc).
//
// EmitOnCreate fields are preserved from prior state.
func (r *Resource) Read(ctx context.Context, req p.ReadRequest) (p.ReadResponse, error) {
	if r.meta.Attachment != nil {
		return r.readAttachment(ctx, req)
	}
	parsed := r.parseIDIntoInputs(req.ID, req.Inputs)
	source := mergeMaps(parsed, req.Properties)
	returnedInputs := req.Inputs
	if req.Inputs.Len() == 0 {
		returnedInputs = parsed
	}
	state, ok, err := r.fetchState(ctx, source, req.Properties)
	if err != nil {
		return p.ReadResponse{}, err
	}
	if !ok {
		// No read op declared: refresh is a no-op, return prior state.
		return p.ReadResponse{ID: req.ID, Inputs: returnedInputs, Properties: req.Properties}, nil
	}
	return p.ReadResponse{ID: req.ID, Properties: state, Inputs: returnedInputs}, nil
}

// fetchState runs the read op and merges EmitOnCreate fields from prior.
// Returns (prior, false, nil) when no read op is declared.
func (r *Resource) fetchState(ctx context.Context, source, prior property.Map) (property.Map, bool, error) {
	op, err := r.resolveOp("read", r.meta.Operations.Read)
	if err != nil {
		return property.Map{}, false, err
	}
	if op == nil {
		return prior, false, nil
	}
	_, state, err := r.execAndDecode(ctx, op, source)
	if err != nil {
		return property.Map{}, false, err
	}
	return r.preserveEmitOnCreate(state, prior), true, nil
}

// preserveEmitOnCreate copies EmitOnCreate fields from oldState into
// newState when missing. Token values are the canonical case.
func (r *Resource) preserveEmitOnCreate(newState, oldState property.Map) property.Map {
	out := map[string]property.Value{}
	for k, v := range newState.AllStable {
		out[k] = v
	}
	for name, fm := range r.meta.Fields {
		if !fm.EmitOnCreate {
			continue
		}
		if _, has := out[name]; has {
			continue
		}
		if v, ok := oldState.GetOk(name); ok {
			out[name] = v
		}
	}
	return property.NewMap(out)
}

// Update fires the update op (if declared), then mirrors Create's
// read-after-create: PATCH/PUT endpoints often return sparse bodies
// (sometimes just the mutated fields), so re-read when a read op is
// declared to pick up server-side fields preserved across the update.
// Without a read op, fall back to merging prior state under the update
// response so fields the update endpoint didn't echo aren't dropped.
func (r *Resource) Update(ctx context.Context, req p.UpdateRequest) (p.UpdateResponse, error) {
	if r.meta.Attachment != nil {
		// Every attachment input is replace-on-change, so Diff always replaces
		// and the engine never calls Update; reaching here is a contract break.
		return p.UpdateResponse{}, fmt.Errorf("attachment resources are replace-only and have no update path")
	}
	op, err := r.resolveOp("update", r.meta.Operations.Update)
	if err != nil {
		return p.UpdateResponse{}, err
	}
	if op == nil {
		return p.UpdateResponse{}, fmt.Errorf("update: resource has no update operation declared")
	}
	if req.DryRun {
		return p.UpdateResponse{Properties: req.Inputs}, nil
	}

	urlSrc := mergeMaps(req.State, req.OldInputs, req.Inputs)
	bodySrc := mergeMaps(req.Inputs, req.OldInputs, req.State)
	_, state, err := r.execAndDecodeSplit(ctx, op, urlSrc, bodySrc)
	if err != nil {
		return p.UpdateResponse{}, err
	}

	readURLSrc := mergeMaps(req.State, state, req.OldInputs, req.Inputs)
	if fetched, ok, err := r.fetchState(ctx, readURLSrc, req.State); err != nil {
		return p.UpdateResponse{}, fmt.Errorf("update: read-after-update: %w", err)
	} else if ok {
		state = fetched
	} else {
		state = mergeMaps(state, req.State)
	}
	return p.UpdateResponse{Properties: state}, nil
}

// Delete fires the delete op (if declared). Without one the engine drops
// state silently. Path params come from req.Properties with req.OldInputs
// as a fallback.
//
// 404 is treated as success — the resource is already gone, which is what
// Delete is trying to achieve. Centralizing this here means every api
// resource benefits without per-resource metadata or scaffolder support;
// the underlying Pulumi Cloud endpoints are uniformly idempotent on this.
func (r *Resource) Delete(ctx context.Context, req p.DeleteRequest) error {
	if r.meta.Attachment != nil {
		return r.deleteAttachment(ctx, req)
	}
	op, err := r.resolveOp("delete", r.meta.Operations.Delete)
	if err != nil {
		return err
	}
	if op == nil {
		return nil
	}
	src := mergeMaps(req.Properties, req.OldInputs)
	if _, _, err := r.execAndDecode(ctx, op, src); err != nil && !IsNotFound(err) {
		return err
	}
	return nil
}

// mergeMaps unions every key. Earlier maps take precedence on conflict.
func mergeMaps(maps ...property.Map) property.Map {
	out := map[string]property.Value{}
	// Walk in reverse so earlier maps overwrite later ones.
	for i := len(maps) - 1; i >= 0; i-- {
		for k, v := range maps[i].AllStable {
			out[k] = v
		}
	}
	return property.NewMap(out)
}

// execAndDecode performs the HTTP round-trip and returns the raw body and
// its decoded property.Map.
func (r *Resource) execAndDecode(
	ctx context.Context, op *Operation, inputs property.Map,
) ([]byte, property.Map, error) {
	return r.execAndDecodeSplit(ctx, op, inputs, inputs)
}

func (r *Resource) execAndDecodeSplit(
	ctx context.Context, op *Operation, urlSrc, bodySrc property.Map,
) ([]byte, property.Map, error) {
	url, err := r.buildURL(op, urlSrc)
	if err != nil {
		return nil, property.Map{}, err
	}

	var body io.Reader
	contentType := ""
	if needsBody(op.Method) {
		switch op.RequestContentType {
		case contentYAML:
			// Raw-string body from the "yaml" input; absent leaves the body empty.
			if v, ok := bodySrc.GetOk("yaml"); ok && v.IsString() {
				body = strings.NewReader(v.AsString())
				contentType = contentYAML
			}
		default:
			bodyJSON, err := json.Marshal(r.buildRequestBody(op, bodySrc))
			if err != nil {
				return nil, property.Map{}, fmt.Errorf("rest: marshal request body for %s: %w", op.ID, err)
			}
			body = bytes.NewReader(bodyJSON)
			contentType = contentJSON
		}
	}

	return r.roundTrip(ctx, op, url, body, contentType)
}

// roundTrip performs the HTTP request against url with an already-built body
// and decodes the response into state. Split out from execAndDecodeSplit so
// the attachment path can supply a hand-shaped body without going through the
// schema-driven buildRequestBody.
func (r *Resource) roundTrip(
	ctx context.Context, op *Operation, url string, body io.Reader, contentType string,
) ([]byte, property.Map, error) {
	transport, err := resolveTransport(ctx)
	if err != nil {
		return nil, property.Map{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, op.Method, url, body)
	if err != nil {
		return nil, property.Map{}, fmt.Errorf("rest: build HTTP request for %s: %w", op.ID, err)
	}
	if body != nil && contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}
	httpReq.Header.Set("Accept", contentJSON)

	resp, err := transport.Do(ctx, httpReq)
	// Transports (authedTransport.Do) rewrite scheme+host on the request
	// in place. Use the post-rewrite URL in user-facing error messages so
	// nobody has to debug the "transport.invalid" sentinel.
	resolvedURL := httpReq.URL.String()
	if err != nil {
		return nil, property.Map{}, fmt.Errorf("rest: %s %s: %w", op.Method, resolvedURL, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, property.Map{}, fmt.Errorf("rest: read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return respBody, property.Map{}, &HTTPError{
			Method:     op.Method,
			URL:        resolvedURL,
			StatusCode: resp.StatusCode,
			Body:       respBody,
		}
	}

	if len(respBody) == 0 || resp.StatusCode == http.StatusNoContent {
		return respBody, property.NewMap(nil), nil
	}
	// Yaml response → bind raw body to state["yaml"].
	if op.ResponseContentType == contentYAML {
		state := property.NewMap(map[string]property.Value{
			"yaml": property.New(string(respBody)),
		})
		return respBody, state, nil
	}
	// Some endpoints return non-object bodies (e.g. a bare boolean); we only
	// build state from object responses, so treat anything else as empty.
	trimmed := bytes.TrimSpace(respBody)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return respBody, property.NewMap(nil), nil
	}
	var raw map[string]any
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return respBody, property.Map{}, fmt.Errorf("rest: decode response for %s: %w", op.ID, err)
	}
	// Translate top-level response keys wire-side → Pulumi-side.
	if len(r.meta.Renames) > 0 {
		raw = renameMapKeys(raw, r.meta.Renames)
	}
	return respBody, anyMapToPropertyMap(raw), nil
}

// renameMapKeys translates wire-side keys to Pulumi-side. Nested maps are
// not touched — renames only apply at the top level of resource I/O.
func renameMapKeys(m map[string]any, renames map[string]string) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[pulumiName(k, renames)] = v
	}
	return out
}

// buildURL substitutes {path} placeholders and prepends the spec's first
// server (or a sentinel host); the Transport is expected to overwrite
// scheme+host before sending.
func (r *Resource) buildURL(op *Operation, inputs property.Map) (string, error) {
	matches := pathParamRE.FindAllStringSubmatchIndex(op.Path, -1)
	var b strings.Builder
	last := 0
	for _, m := range matches {
		b.WriteString(op.Path[last:m[0]])
		wireName := op.Path[m[2]:m[3]]
		pulName := pulumiName(wireName, r.meta.Renames)
		v, ok := inputs.GetOk(pulName)
		if !ok {
			return "", &MissingPathParamError{WireName: wireName, PulumiName: pulName}
		}
		b.WriteString(url.PathEscape(propertyValueToString(v)))
		last = m[1]
	}
	b.WriteString(op.Path[last:])
	base := "https://transport.invalid"
	if len(r.spec.Servers) > 0 {
		base = strings.TrimRight(r.spec.Servers[0], "/")
	}
	return base + b.String(), nil
}

func needsBody(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	}
	return false
}

// propertyValueToString stringifies a Value for path/URL substitution.
// Numbers go through strconv.FormatFloat with 'f' formatting so large
// integer-valued floats (timestamps, integer IDs) don't render in
// scientific notation — `1e+18` makes a useless URL path segment.
func propertyValueToString(v property.Value) string {
	switch {
	case v.IsString():
		return v.AsString()
	case v.IsNumber():
		return strconv.FormatFloat(v.AsNumber(), 'f', -1, 64)
	case v.IsBool():
		return strconv.FormatBool(v.AsBool())
	case v.IsNull():
		return ""
	default:
		return fmt.Sprintf("%v", propertyValueToAny(v))
	}
}

// propertyMapToAny converts a property.Map to a json-marshal-ready any tree.
// Secrets are unwrapped — the wire shape doesn't preserve them.
func propertyMapToAny(m property.Map) map[string]any {
	out := make(map[string]any, m.Len())
	for k, v := range m.AllStable {
		out[k] = propertyValueToAny(v)
	}
	return out
}

// buildRequestBody assembles the JSON body for a request from the
// operation's body schema (op.RequestRef): for every wire-side field the
// schema declares, pull the value from inputs using the Pulumi-side name
// (i.e. apply renames in reverse). Inputs that don't correspond to a body
// field aren't sent — that filters out path/query params naturally.
//
// Some Pulumi Cloud endpoints (e.g. CreateOrganizationWebhook) accept the
// same field in both the URL and the body and validate they match, so we
// can't blanket-strip path-param-named fields from the body. Driving the
// shape from the schema gives the API exactly what it expects regardless.
//
// Without a RequestRef (e.g. action-style POSTs with empty bodies, or
// operations the parser couldn't classify) fall back to serializing the
// pulumi inputs as-is.
func (r *Resource) buildRequestBody(op *Operation, inputs property.Map) map[string]any {
	if op.RequestRef == "" {
		return propertyMapToAny(inputs)
	}
	bodyProps, _, err := flattenObjectSchema(r.spec, op.RequestRef)
	if err != nil {
		return propertyMapToAny(inputs)
	}
	out := make(map[string]any, len(bodyProps))
	for wireKey := range bodyProps {
		pulKey := pulumiName(wireKey, r.meta.Renames)
		if v, ok := inputs.GetOk(pulKey); ok {
			out[wireKey] = propertyValueToAny(v)
		}
	}
	return out
}

func propertyValueToAny(v property.Value) any {
	switch {
	case v.IsNull():
		return nil
	case v.IsBool():
		return v.AsBool()
	case v.IsNumber():
		return v.AsNumber()
	case v.IsString():
		return v.AsString()
	case v.IsArray():
		arr := v.AsArray()
		out := make([]any, 0, arr.Len())
		for _, e := range arr.AsSlice() {
			out = append(out, propertyValueToAny(e))
		}
		return out
	case v.IsMap():
		return propertyMapToAny(v.AsMap())
	default:
		return nil
	}
}

// anyMapToPropertyMap converts a JSON-decoded map to a property.Map.
func anyMapToPropertyMap(m map[string]any) property.Map {
	out := make(map[string]property.Value, len(m))
	for k, v := range m {
		out[k] = anyToPropertyValue(v)
	}
	return property.NewMap(out)
}

func anyToPropertyValue(v any) property.Value {
	switch x := v.(type) {
	case nil:
		return property.New(property.Null)
	case bool:
		return property.New(x)
	case float64:
		return property.New(x)
	case string:
		return property.New(x)
	case []any:
		arr := make([]property.Value, len(x))
		for i, e := range x {
			arr[i] = anyToPropertyValue(e)
		}
		return property.New(property.NewArray(arr))
	case map[string]any:
		return property.New(anyMapToPropertyMap(x))
	default:
		return property.New(property.Null)
	}
}

// mapEqual is a structural comparison helper for Diff.
func mapEqual(a, b property.Map) bool {
	return property.New(a).Equals(property.New(b))
}
