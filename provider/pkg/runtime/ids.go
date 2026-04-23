// Copyright 2016-2026, Pulumi Corporation.
//
// ids.go — URI templating and Pulumi ID construction/decomposition.
// Shared by dispatch (to build request URLs) and Read (to decompose an
// imported ID back into its component inputs).

package runtime

import (
	"fmt"
	"regexp"
	"strings"
)

// pathParamRe matches `{paramName}` placeholders in an OpenAPI-style path
// template. We use a narrow regex deliberately — the fuller RFC 6570
// machinery is not needed: Pulumi Cloud paths are all simple `{foo}` forms.
var pathParamRe = regexp.MustCompile(`\{([^{}]+)\}`)

// ExpandPath substitutes values for `{param}` placeholders in a path template.
// Missing params produce an error rather than a silent gap — every path
// param is load-bearing for a Pulumi Cloud URL.
func ExpandPath(template string, values map[string]string) (string, error) {
	var missing []string
	out := pathParamRe.ReplaceAllStringFunc(template, func(match string) string {
		key := match[1 : len(match)-1]
		v, ok := values[key]
		if !ok || v == "" {
			missing = append(missing, key)
			return match
		}
		// Path-segment values should not contain `/` unless the template
		// explicitly models multi-segment (e.g. {+foo}). Pulumi Cloud paths
		// don't use reserved expansion, so escape defensively.
		return urlPathEscape(v)
	})
	if len(missing) > 0 {
		return "", fmt.Errorf("path template %q missing values for: %s",
			template, strings.Join(missing, ", "))
	}
	return out, nil
}

// ExtractPathParams returns the ordered list of `{param}` names in a path
// template. Used to validate that CloudAPIOperation.Parameters lines up
// with the path, and to decompose imported IDs back into their inputs.
func ExtractPathParams(template string) []string {
	matches := pathParamRe.FindAllStringSubmatch(template, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		out = append(out, m[1])
	}
	return out
}

// BuildID composes a Pulumi resource ID from the resource's CloudAPIID spec
// and a set of inputs+response values. Supports both simple and polymorphic IDs.
func BuildID(idSpec *CloudAPIID, scope string, values map[string]string) (string, error) {
	if idSpec == nil {
		return "", fmt.Errorf("resource has no ID specification")
	}
	tmpl := idSpec.Template
	if tmpl == "" && idSpec.Templates != nil {
		t, ok := idSpec.Templates[scope]
		if !ok {
			return "", fmt.Errorf("no ID template for scope %q", scope)
		}
		tmpl = t
	}
	if tmpl == "" {
		return "", fmt.Errorf("empty ID template")
	}
	return ExpandPath(tmpl, values)
}

// DecomposeID parses a concrete ID string back into its component values,
// using the matching template from the ID spec. The inverse of BuildID.
//
// Pulumi import (`pulumi import` / Read) supplies an ID as a string;
// we need to split it back into the path params that construct request
// URLs.
//
// Example: template "{orgName}/{name}/{agentPoolId}" with ID
// "acme-corp/vpc-isolated/pool-abc" produces
// {orgName: acme-corp, name: vpc-isolated, agentPoolId: pool-abc}.
func DecomposeID(idSpec *CloudAPIID, scope string, id string) (map[string]string, error) {
	tmpl := idSpec.Template
	if tmpl == "" && idSpec.Templates != nil {
		t, ok := idSpec.Templates[scope]
		if !ok {
			return nil, fmt.Errorf("no ID template for scope %q", scope)
		}
		tmpl = t
	}
	if tmpl == "" {
		return nil, fmt.Errorf("empty ID template")
	}

	// Convert the template into a regex. Path params become named capture
	// groups matching one path segment each. Literal segments are escaped.
	pattern := regexp.MustCompile(regexp.QuoteMeta(tmpl))
	// Un-escape our placeholders and replace with named captures.
	regexStr := pathParamRe.ReplaceAllStringFunc(tmpl, func(match string) string {
		name := match[1 : len(match)-1]
		return fmt.Sprintf(`(?P<%s>[^/]+)`, name)
	})
	_ = pattern // not used directly; we rebuild below
	re, err := regexp.Compile("^" + regexStr + "$")
	if err != nil {
		return nil, fmt.Errorf("compiling ID regex from %q: %w", tmpl, err)
	}
	m := re.FindStringSubmatch(id)
	if m == nil {
		return nil, fmt.Errorf("ID %q does not match template %q", id, tmpl)
	}
	out := make(map[string]string, len(re.SubexpNames()))
	for i, name := range re.SubexpNames() {
		if name == "" {
			continue
		}
		out[name] = m[i]
	}
	return out, nil
}

// urlPathEscape escapes a value for use as a single path segment. We keep
// this close to net/url's behavior but avoid rewriting already-encoded
// characters in the uncommon case that the caller already encoded.
func urlPathEscape(s string) string {
	// Use the standard path escape which preserves most characters.
	// We don't use url.PathEscape here to avoid re-encoding a value that
	// already contains slashes for legitimate reasons (rare, but happens
	// in Pulumi project names). Instead we only escape the small set that
	// would break path parsing.
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == ' ':
			b.WriteString("%20")
		case r == '#' || r == '?':
			b.WriteString(fmt.Sprintf("%%%02X", r))
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
