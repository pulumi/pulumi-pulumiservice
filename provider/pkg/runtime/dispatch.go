// Copyright 2016-2026, Pulumi Corporation.
//
// dispatch.go — generic CRUD dispatcher. One function per Pulumi gRPC verb
// (Create/Read/Update/Delete/Check), each parameterized by CloudAPIResource
// metadata. There is no per-resource Go code — every supported resource
// is expressed in provider/resource-map.yaml.

package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

// Dispatcher owns the HTTP client and the metadata registry. Provider.go
// instantiates one Dispatcher at boot and routes every CRUD gRPC call to it.
type Dispatcher struct {
	Client   *Client
	Metadata *CloudAPIMetadata
}

// NewDispatcher constructs a Dispatcher.
func NewDispatcher(client *Client, metadata *CloudAPIMetadata) *Dispatcher {
	return &Dispatcher{Client: client, Metadata: metadata}
}

// Create executes the create operation for a resource and returns the new
// ID plus the resource's output property map.
//
// Flow:
//   1. Resolve the resource metadata by token.
//   2. Select the operation set (polymorphic-aware).
//   3. Split inputs into path/query/body per Property metadata.
//   4. Build the request URL from the create operation's path template.
//   5. POST to the server; decode the response.
//   6. Compose the Pulumi ID from the ID template and response values.
//   7. Convert the response back into a Pulumi property map.
func (d *Dispatcher) Create(ctx context.Context, token string, inputs resource.PropertyMap) (string, resource.PropertyMap, error) {
	res, ok := d.Metadata.Resources[token]
	if !ok {
		return "", nil, fmt.Errorf("no metadata for resource %q", token)
	}
	split, err := InputsToRequest(res.Properties, inputs, res.Discriminator, true)
	if err != nil {
		return "", nil, fmt.Errorf("preparing inputs: %w", err)
	}
	ops, err := SelectOperations(&res, split.Scope)
	if err != nil {
		return "", nil, err
	}
	if ops.Create == nil {
		return "", nil, fmt.Errorf("resource %q has no create operation", token)
	}

	path, err := ExpandPath(ops.Create.PathTemplate, split.Path)
	if err != nil {
		return "", nil, fmt.Errorf("expanding create path: %w", err)
	}
	req := Request{
		Method:      ops.Create.Method,
		Path:        path,
		Query:       toURLValues(split.Query),
		ContentType: ops.Create.ContentType,
	}
	if rb := rawBodyFor(ops.Create, inputs); rb != nil {
		req.RawBody = rb
	} else {
		req.Body = bodyFor(ops.Create, split.Body)
	}
	resp, err := d.Client.Call(ctx, req)
	if err != nil {
		return "", nil, err
	}

	var raw map[string]interface{}
	if err := resp.Decode(&raw); err != nil {
		return "", nil, fmt.Errorf("decoding create response: %w", err)
	}

	outputs := ResponseToOutputs(res.Properties, raw, inputs)

	// Compose ID using SDK property names. ID templates are user-facing
	// (they surface in `pulumi import`, `pulumi state`), so the template
	// is keyed by the same names a user types in code.
	idValues := sdkValues(inputs, outputs)
	id, err := BuildID(res.ID, split.Scope, idValues)
	if err != nil {
		return "", nil, fmt.Errorf("building ID: %w", err)
	}

	// PostCreate: second-phase op (e.g., ESC Environment's raw-YAML PATCH
	// after POST created the empty env). Uses the non-create input split —
	// path params from the resource URL, raw body from the RawBodyFrom
	// property if set.
	if res.PostCreate != nil {
		if err := d.runFollowOn(ctx, res.PostCreate, &res, inputs); err != nil {
			return "", nil, fmt.Errorf("post-create step: %w", err)
		}
	}
	return id, outputs, nil
}

// runFollowOn executes an additional operation after the primary CRUD call
// (today: only PostCreate). Input handling matches update semantics: path
// params from the resource's normal source values, body via rawBody or
// the JSON split. No ID construction — the resource ID was composed by
// the primary call.
func (d *Dispatcher) runFollowOn(ctx context.Context, op *CloudAPIOperation, res *CloudAPIResource, inputs resource.PropertyMap) error {
	split, err := InputsToRequest(res.Properties, inputs, res.Discriminator, false)
	if err != nil {
		return err
	}
	path, err := ExpandPath(op.PathTemplate, split.Path)
	if err != nil {
		return err
	}
	req := Request{
		Method:      op.Method,
		Path:        path,
		Query:       toURLValues(split.Query),
		ContentType: op.ContentType,
	}
	if rb := rawBodyFor(op, inputs); rb != nil {
		req.RawBody = rb
	} else {
		req.Body = bodyFor(op, split.Body)
	}
	_, err = d.Client.Call(ctx, req)
	return err
}

// sdkValues builds a stringified value map keyed by SDK property name,
// preferring values from outputs (post-Create response) but falling back
// to inputs for properties the API doesn't echo (e.g. path parameters).
func sdkValues(inputs, outputs resource.PropertyMap) map[string]string {
	out := map[string]string{}
	write := func(m resource.PropertyMap) {
		for k, v := range m {
			val := v
			if val.IsSecret() {
				val = val.SecretValue().Element
			}
			if val.IsComputed() || val.IsNull() {
				continue
			}
			switch {
			case val.IsString():
				out[string(k)] = val.StringValue()
			case val.IsNumber():
				out[string(k)] = fmt.Sprintf("%v", val.NumberValue())
			case val.IsBool():
				out[string(k)] = fmt.Sprintf("%v", val.BoolValue())
			}
		}
	}
	write(inputs)
	write(outputs) // outputs win; server-assigned IDs overwrite any placeholder inputs
	return out
}

// Read executes the read/get operation. Supports both a dedicated read
// endpoint and the ListAndFilter fallback (readVia).
func (d *Dispatcher) Read(ctx context.Context, token, id string, priorInputs resource.PropertyMap) (string, resource.PropertyMap, error) {
	res, ok := d.Metadata.Resources[token]
	if !ok {
		return "", nil, fmt.Errorf("no metadata for resource %q", token)
	}

	// Decompose the ID back into path values. For polymorphic resources we
	// try each scope's template in turn until one matches; this supports
	// `pulumi import` without the user having to name the scope.
	scope, pathValues, err := decomposeIDPolymorphic(res.ID, id)
	if err != nil {
		return "", nil, fmt.Errorf("decomposing ID %q: %w", id, err)
	}
	ops, err := SelectOperations(&res, scope)
	if err != nil {
		return "", nil, err
	}

	wirePathValues := translateSDKToWire(res.Properties, pathValues)

	readOp := ops.Read
	if readOp == nil && res.ReadVia != nil {
		if res.ReadVia.ExtractField != "" {
			return d.readViaParentField(ctx, &res, res.ReadVia, wirePathValues, id, priorInputs)
		}
		return d.readViaListFilter(ctx, &res, res.ReadVia, wirePathValues, id, priorInputs)
	}
	if readOp == nil {
		// No read endpoint and no readVia — fall back to trusting the prior
		// state. This covers resources whose public API exposes no per-item
		// GET (e.g., TeamStackPermission). Drift against out-of-band
		// mutations isn't detected, but the resource is at least usable.
		return id, priorInputs, nil
	}

	path, err := ExpandPath(readOp.PathTemplate, wirePathValues)
	if err != nil {
		return "", nil, fmt.Errorf("expanding read path: %w", err)
	}
	resp, err := d.Client.Call(ctx, Request{
		Method: readOp.Method, Path: path,
		ContentType: readOp.ContentType,
	})
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
			return "", nil, nil // resource gone — Pulumi treats this as "needs recreate"
		}
		return "", nil, err
	}
	outputs, err := decodeReadBody(readOp, resp, res.Properties, priorInputs)
	if err != nil {
		return "", nil, err
	}
	return id, outputs, nil
}

// decodeReadBody turns a Read (or Update) response into a PropertyMap.
// Honors the operation's RawBodyTo: if set, the whole response body is
// stored as a string under the named property (for endpoints that return
// non-JSON content, e.g. ESC Environment's application/x-yaml GET).
// Otherwise JSON-decodes and maps via the resource's property metadata.
//
// Secret preservation: if the RawBodyTo property is marked `secret: true`
// in the resource's metadata, the extracted string is wrapped in
// resource.MakeSecret. This is load-bearing for ESC Environment, whose
// YAML content routinely carries secrets — a miss here would leak the
// plaintext into Pulumi state.
func decodeReadBody(op *CloudAPIOperation, resp *Response, props map[string]CloudAPIProperty, priorInputs resource.PropertyMap) (resource.PropertyMap, error) {
	if op != nil && op.RawBodyTo != "" {
		out := resource.PropertyMap{}
		// Carry identity (path) inputs forward — the raw body response
		// doesn't include them.
		for name, meta := range props {
			if meta.Source != "path" {
				continue
			}
			if v, ok := priorInputs[resource.PropertyKey(name)]; ok {
				out[resource.PropertyKey(name)] = v
			}
		}
		val := resource.NewStringProperty(string(resp.Body))
		if meta, ok := props[op.RawBodyTo]; ok && meta.Secret {
			val = resource.MakeSecret(val)
		}
		out[resource.PropertyKey(op.RawBodyTo)] = val
		return out, nil
	}
	var raw map[string]interface{}
	if err := resp.Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return ResponseToOutputs(props, raw, priorInputs), nil
}

// Update runs the update operation, returning the refreshed output map.
func (d *Dispatcher) Update(ctx context.Context, token, id string, olds, news resource.PropertyMap) (resource.PropertyMap, error) {
	res, ok := d.Metadata.Resources[token]
	if !ok {
		return nil, fmt.Errorf("no metadata for resource %q", token)
	}
	split, err := InputsToRequest(res.Properties, news, res.Discriminator, false)
	if err != nil {
		return nil, err
	}
	ops, err := SelectOperations(&res, split.Scope)
	if err != nil {
		return nil, err
	}
	if ops.Update == nil {
		return nil, fmt.Errorf("resource %q has no update operation", token)
	}
	// Path values come from the decomposed ID, not the inputs — some path
	// params (e.g. server-assigned poolId) aren't in the inputs map.
	idValues, err := DecomposeID(res.ID, split.Scope, id)
	if err != nil {
		return nil, fmt.Errorf("decomposing ID: %w", err)
	}
	path, err := ExpandPath(ops.Update.PathTemplate, translateSDKToWire(res.Properties, idValues))
	if err != nil {
		return nil, err
	}
	updReq := Request{
		Method: ops.Update.Method, Path: path,
		Query:       toURLValues(split.Query),
		ContentType: ops.Update.ContentType,
	}
	if rb := rawBodyFor(ops.Update, news); rb != nil {
		updReq.RawBody = rb
	} else {
		updReq.Body = bodyFor(ops.Update, split.Body)
	}
	resp, err := d.Client.Call(ctx, updReq)
	if err != nil {
		return nil, err
	}
	outputs, err := decodeReadBody(ops.Update, resp, res.Properties, news)
	if err != nil {
		return nil, err
	}
	return outputs, nil
}

// Delete runs the delete operation.
func (d *Dispatcher) Delete(ctx context.Context, token, id string, state resource.PropertyMap) error {
	res, ok := d.Metadata.Resources[token]
	if !ok {
		return fmt.Errorf("no metadata for resource %q", token)
	}
	// Derive scope from state (discriminator is itself a property).
	scope := ""
	if res.Discriminator != "" {
		if v, ok := state[resource.PropertyKey(res.Discriminator)]; ok && v.IsString() {
			scope = v.StringValue()
		}
	}
	ops, err := SelectOperations(&res, scope)
	if err != nil {
		return err
	}
	if ops.Delete == nil {
		return fmt.Errorf("resource %q has no delete operation", token)
	}
	idValues, err := DecomposeID(res.ID, scope, id)
	if err != nil {
		return fmt.Errorf("decomposing ID: %w", err)
	}
	wireBase := translateSDKToWire(res.Properties, idValues)

	// Iterate-delete: the delete op names a property whose map keys should
	// each be fired as a separate call (e.g., Tags deleting per tagName).
	if ops.Delete.IterateOver != "" {
		return d.iterateDelete(ctx, ops.Delete, state, wireBase)
	}

	path, err := ExpandPath(ops.Delete.PathTemplate, wireBase)
	if err != nil {
		return err
	}
	if _, err := d.Client.Call(ctx, Request{
		Method: ops.Delete.Method, Path: path,
		Body: bodyFor(ops.Delete, nil),
	}); err != nil {
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
			return nil // already gone; safe to consider deleted
		}
		return err
	}
	return nil
}

// iterateDelete fires ops.Delete once per key in the state's IterateOver
// property (a map/object). Each call substitutes the key into the path
// template under IterateKeyParam. Missing entries (404) are tolerated —
// deletion is idempotent.
func (d *Dispatcher) iterateDelete(ctx context.Context, op *CloudAPIOperation, state resource.PropertyMap, wireBase map[string]string) error {
	v, ok := state[resource.PropertyKey(op.IterateOver)]
	if !ok {
		return nil // nothing tracked, nothing to delete
	}
	if v.IsSecret() {
		v = v.SecretValue().Element
	}
	if !v.IsObject() {
		return fmt.Errorf("iterateOver property %q must be an object/map", op.IterateOver)
	}
	param := op.IterateKeyParam
	if param == "" {
		return fmt.Errorf("iterateKeyParam is required when iterateOver is set (op %s)", op.OperationID)
	}
	// Walk keys in sorted order. Go's map iteration is intentionally
	// randomized; deleting sub-resources in a nondeterministic order would
	// yield unstable logs, flaky tests, and hard-to-reproduce partial-failure
	// states if the API is ever order-sensitive.
	keys := make([]string, 0, len(v.ObjectValue()))
	for k := range v.ObjectValue() {
		keys = append(keys, string(k))
	}
	sort.Strings(keys)
	for _, k := range keys {
		values := map[string]string{}
		for pk, pv := range wireBase {
			values[pk] = pv
		}
		values[param] = k
		path, err := ExpandPath(op.PathTemplate, values)
		if err != nil {
			return err
		}
		if _, err := d.Client.Call(ctx, Request{Method: op.Method, Path: path}); err != nil {
			if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
				continue // already gone
			}
			return fmt.Errorf("deleting %s=%q: %w", param, k, err)
		}
	}
	return nil
}

// bodyFor returns the request body for an operation. If the operation has a
// BodyOverride set (tombstone-style delete via update op), that wins over
// the input-derived body. Otherwise the default input-derived body is used.
func bodyFor(op *CloudAPIOperation, inputBody map[string]interface{}) interface{} {
	if op != nil && op.BodyOverride != nil {
		return op.BodyOverride
	}
	return asBody(inputBody)
}

// rawBodyFor returns the raw HTTP body bytes for an operation with
// RawBodyFrom set. Pulls the property value out of inputs as a string,
// unwrapping secrets. Returns nil if the property is absent/null.
func rawBodyFor(op *CloudAPIOperation, inputs resource.PropertyMap) []byte {
	if op == nil || op.RawBodyFrom == "" {
		return nil
	}
	v, ok := inputs[resource.PropertyKey(op.RawBodyFrom)]
	if !ok {
		return nil
	}
	if v.IsSecret() {
		v = v.SecretValue().Element
	}
	if !v.IsString() {
		return nil
	}
	return []byte(v.StringValue())
}

// readViaListFilter implements the list-and-filter fallback for resources
// whose spec has no single-resource GET.
func (d *Dispatcher) readViaListFilter(ctx context.Context, res *CloudAPIResource, rv *CloudAPIReadVia, pathValues map[string]string, id string, priorInputs resource.PropertyMap) (string, resource.PropertyMap, error) {
	// The list operation lives in Functions (all list ops are exposed as
	// Pulumi functions too). We look it up by operationId to get its
	// path template. No explicit resource-level tracking needed.
	fn, ok := lookupFunctionByOperationID(d.Metadata.Functions, rv.OperationID)
	if !ok {
		return "", nil, fmt.Errorf("readVia operationId %q not found in metadata.functions", rv.OperationID)
	}
	path, err := ExpandPath(fn.Operation.PathTemplate, pathValues)
	if err != nil {
		return "", nil, err
	}
	resp, err := d.Client.Call(ctx, Request{Method: fn.Operation.Method, Path: path})
	if err != nil {
		return "", nil, err
	}
	var listed struct {
		Items []map[string]interface{} `json:"items"`
	}
	if err := json.Unmarshal(resp.Body, &listed); err != nil {
		// Some endpoints return a bare JSON array; tolerate that shape too.
		var bare []map[string]interface{}
		if err2 := json.Unmarshal(resp.Body, &bare); err2 != nil {
			return "", nil, fmt.Errorf("decoding list response: %w", err)
		}
		listed.Items = bare
	}
	// Match id against the configured filter field.
	for _, item := range listed.Items {
		if v, ok := item[rv.FilterBy]; ok && fmt.Sprintf("%v", v) == tailOfID(id) {
			return id, ResponseToOutputs(res.Properties, item, priorInputs), nil
		}
	}
	return "", nil, nil // not found — Pulumi will treat as needs recreate
}

func lookupFunctionByOperationID(fns map[string]CloudAPIFunction, oid string) (CloudAPIFunction, bool) {
	for _, fn := range fns {
		if fn.Operation.OperationID == oid {
			return fn, true
		}
	}
	return CloudAPIFunction{}, false
}

// lookupOperationByID finds a CloudAPIOperation with the given operationId
// anywhere in the metadata — resources' CRUD slots, functions, or methods.
// Used by readVia's parent-field extraction to resolve the GET it shares
// with another resource.
func lookupOperationByID(md *CloudAPIMetadata, oid string) (*CloudAPIOperation, bool) {
	for _, r := range md.Resources {
		for _, op := range []*CloudAPIOperation{r.Create, r.Read, r.Update, r.Delete} {
			if op != nil && op.OperationID == oid {
				return op, true
			}
		}
		if r.PolymorphicScopes != nil {
			for _, scope := range r.PolymorphicScopes.Scopes {
				for _, op := range []*CloudAPIOperation{scope.Create, scope.Read, scope.Update, scope.Delete} {
					if op != nil && op.OperationID == oid {
						return op, true
					}
				}
			}
		}
	}
	for _, fn := range md.Functions {
		if fn.Operation.OperationID == oid {
			return &fn.Operation, true
		}
	}
	for _, m := range md.Methods {
		if m.Operation.OperationID == oid {
			return &m.Operation, true
		}
	}
	return nil, false
}

// readViaParentField implements the "read through a parent resource" pattern:
// call the parent's GET, pluck a named field out of the response, and — if
// KeyBy is set — look up a single entry in that map by a resource property's
// value. Used for per-key children like stack tags (stored on the parent
// stack's GET, not on their own URL).
func (d *Dispatcher) readViaParentField(ctx context.Context, res *CloudAPIResource, rv *CloudAPIReadVia, pathValues map[string]string, id string, priorInputs resource.PropertyMap) (string, resource.PropertyMap, error) {
	parentOp, ok := lookupOperationByID(d.Metadata, rv.OperationID)
	if !ok {
		return "", nil, fmt.Errorf("readVia parent operationId %q not found in metadata", rv.OperationID)
	}
	path, err := ExpandPath(parentOp.PathTemplate, pathValues)
	if err != nil {
		return "", nil, err
	}
	resp, err := d.Client.Call(ctx, Request{Method: parentOp.Method, Path: path})
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
			return "", nil, nil // parent gone → child gone
		}
		return "", nil, err
	}
	var body map[string]interface{}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return "", nil, fmt.Errorf("decoding parent read response: %w", err)
	}
	extracted, ok := body[rv.ExtractField].(map[string]interface{})
	if !ok {
		// Field absent or wrong shape — treat as if our entry is gone.
		return "", nil, nil
	}

	// Seed outputs with the inputs' identity fields (path params don't come
	// from the parent's response — they come from the decomposed ID).
	outputs := resource.PropertyMap{}
	for name, meta := range res.Properties {
		if meta.Source != "path" {
			continue
		}
		if v, ok := priorInputs[resource.PropertyKey(name)]; ok {
			outputs[resource.PropertyKey(name)] = v
		}
	}

	if rv.KeyBy == "" {
		// Whole-map mode: attach the map under ExtractField as a Pulumi
		// object property. If the ExtractField property is secret, the
		// whole map is wrapped — we can't know which entries are sensitive
		// from the response alone, so erring on the side of secrecy is
		// correct.
		pm := resource.PropertyMap{}
		for k, v := range extracted {
			if s, ok := v.(string); ok {
				pm[resource.PropertyKey(k)] = resource.NewStringProperty(s)
			} else {
				pm[resource.PropertyKey(k)] = resource.NewPropertyValue(v)
			}
		}
		val := resource.NewObjectProperty(pm)
		if meta, ok := res.Properties[rv.ExtractField]; ok && meta.Secret {
			val = resource.MakeSecret(val)
		}
		outputs[resource.PropertyKey(rv.ExtractField)] = val
		return id, outputs, nil
	}

	// Single-entry mode: look up the key named by the KeyBy property's value.
	keyV, ok := priorInputs[resource.PropertyKey(rv.KeyBy)]
	if !ok || !keyV.IsString() {
		return "", nil, fmt.Errorf("readVia keyBy property %q not present or not a string in prior inputs", rv.KeyBy)
	}
	entry, ok := extracted[keyV.StringValue()]
	if !ok {
		return "", nil, nil // entry gone out-of-band
	}
	valueProp := rv.ValueProperty
	if valueProp == "" {
		valueProp = "value"
	}
	outputs[resource.PropertyKey(rv.KeyBy)] = keyV
	var val resource.PropertyValue
	if s, ok := entry.(string); ok {
		val = resource.NewStringProperty(s)
	} else {
		val = resource.NewPropertyValue(entry)
	}
	// Wrap the extracted value as secret if the value property is marked
	// secret in the resource's metadata. Prevents extracted sensitive
	// values from landing unencrypted in Pulumi state.
	if meta, ok := res.Properties[valueProp]; ok && meta.Secret {
		val = resource.MakeSecret(val)
	}
	outputs[resource.PropertyKey(valueProp)] = val
	return id, outputs, nil
}

// decomposeIDPolymorphic tries the ID template(s) in order and returns
// the matching scope plus decomposed values. For non-polymorphic resources
// it just delegates to DecomposeID with an empty scope.
func decomposeIDPolymorphic(spec *CloudAPIID, id string) (string, map[string]string, error) {
	if spec == nil {
		return "", nil, fmt.Errorf("resource has no ID specification")
	}
	if spec.Template != "" {
		v, err := DecomposeID(spec, "", id)
		return "", v, err
	}
	for scope := range spec.Templates {
		if v, err := DecomposeID(spec, scope, id); err == nil {
			return scope, v, nil
		}
	}
	return "", nil, fmt.Errorf("ID %q does not match any template", id)
}

// tailOfID returns the last segment of a slash-delimited ID — the common
// case for "compare the server-assigned ID to our filter field."
func tailOfID(id string) string {
	for i := len(id) - 1; i >= 0; i-- {
		if id[i] == '/' {
			return id[i+1:]
		}
	}
	return id
}

// toURLValues converts a string→string map into url.Values.
func toURLValues(m map[string]string) url.Values {
	if len(m) == 0 {
		return nil
	}
	v := url.Values{}
	for k, val := range m {
		v.Set(k, val)
	}
	return v
}

// asBody returns nil if the body map is empty, so we don't POST `{}` when
// the API expects an actual empty body. (Some endpoints are strict.)
func asBody(m map[string]interface{}) interface{} {
	if len(m) == 0 {
		return nil
	}
	return m
}

// translateSDKToWire converts an SDK-keyed value map (what DecomposeID
// returns) into a wire-keyed one (what path templates expect). For each
// property with a `From:` override, emit the value under the From name;
// otherwise carry the SDK name through unchanged. Used by Read/Update/Delete
// to expand OpenAPI path templates against the values decoded from a
// Pulumi-side ID.
func translateSDKToWire(props map[string]CloudAPIProperty, sdkValues map[string]string) map[string]string {
	out := make(map[string]string, len(sdkValues))
	for k, v := range sdkValues {
		wireName := k
		if meta, ok := props[k]; ok && meta.From != "" {
			wireName = meta.From
		}
		out[wireName] = v
		// Also carry the SDK name through; path templates may legitimately
		// use either form depending on how the spec named the placeholder.
		// Duplicate entries are harmless.
		out[k] = v
	}
	return out
}

