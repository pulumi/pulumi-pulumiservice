//go:build api
// +build api

// Pre-test sweeper for the api test org. The TestApi suite creates real
// resources on the configured Pulumi backend; if a test panics or the
// process is killed before t.Cleanup runs, those resources leak. They
// then collide with the next run (`already exists` 409s on
// non-idempotent endpoints, or shared singletons on the org).
//
// sweepOnce runs once per `go test` invocation, before TestApi starts,
// and deletes leftover resources that match the test-naming conventions
// used by this suite and are older than sweepMinAge.
//
// Set PULUMITEST_SKIP_SWEEP=1 to bypass.

package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

var sweepOnce sync.Once

// sweepApiTestState removes leftover test resources from the configured
// Pulumi org. Best-effort: failures are logged, never fatal.
func sweepApiTestState(t *testing.T) {
	sweepOnce.Do(func() {
		if os.Getenv("PULUMITEST_SKIP_SWEEP") != "" {
			t.Logf("sweep: PULUMITEST_SKIP_SWEEP set; skipping")
			return
		}
		token := os.Getenv("PULUMI_ACCESS_TOKEN")
		apiURL := os.Getenv("PULUMI_BACKEND_URL")
		if token == "" || apiURL == "" {
			t.Logf("sweep: no PULUMI_ACCESS_TOKEN or PULUMI_BACKEND_URL; skipping")
			return
		}
		s := &sweeper{
			t:      t,
			token:  token,
			apiURL: strings.TrimRight(apiURL, "/"),
			org:    ServiceProviderTestOrg,
			http:   &http.Client{Timeout: 30 * time.Second},
			now:    time.Now(),
		}
		// Webhooks, teams, policy groups and template sources aren't swept:
		// their list responses carry no creation time, so a live resource
		// from a concurrent run can't be told apart from a leaked one.
		s.sweepOidcIssuers()
		s.sweepRoles()
		s.sweepEnvironments()
		s.sweepAgentPools()
		s.sweepOrgTokens()
	})
}

type sweeper struct {
	t      *testing.T
	token  string
	apiURL string
	org    string
	http   *http.Client
	now    time.Time
}

// sweepMinAge is how old a resource must be before the sweeper deletes it.
// The acceptance tests in test.yml share this org and these name prefixes,
// and they run at the same time as this suite on other PRs, so a younger
// resource may still be in use. It exceeds the longest test timeout (2h).
const sweepMinAge = 3 * time.Hour

// sweepTime is a creation time as the list endpoints return it, which varies
// by endpoint: RFC 3339, "2006-01-02 15:04:05.000" in UTC, or epoch
// milliseconds. It is zero when the field is missing or unparseable.
type sweepTime struct{ time.Time }

func (st *sweepTime) UnmarshalJSON(b []byte) error {
	st.Time = time.Time{}
	var ms int64
	if err := json.Unmarshal(b, &ms); err == nil && ms > 0 {
		st.Time = time.UnixMilli(ms).UTC()
		return nil
	}
	var s string
	if json.Unmarshal(b, &s) != nil {
		return nil
	}
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05.999999999"} {
		if t, err := time.Parse(layout, s); err == nil {
			st.Time = t
			return nil
		}
	}
	return nil
}

// shouldSweep returns true if a resource looks like a leftover from the
// test suite: its name follows a test-suite naming convention and it was
// created at least sweepMinAge before now. A resource with no known
// creation time is never swept.
//
// Conservative: only sweeps obvious test names so we don't nuke
// human-created resources by accident.
func shouldSweep(name string, created sweepTime, now time.Time) bool {
	if created.IsZero() || now.Sub(created.Time) < sweepMinAge {
		return false
	}
	for _, pre := range []string{
		"api-", "v1-", "test-", "org-webhook-", "stack-webhook-",
		"env-webhook-", "brand-new-api-", "brand-new-v1-", "example-policy-group",
		"github_issuer", "pulumi_issuer",
	} {
		if strings.HasPrefix(name, pre) {
			return true
		}
	}
	return false
}

func (s *sweeper) get(path string, out any) error {
	req, _ := http.NewRequestWithContext(context.Background(), "GET", s.apiURL+path, nil)
	req.Header.Set("Authorization", "token "+s.token)
	req.Header.Set("Accept", "application/vnd.pulumi+9")
	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%s: HTTP %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (s *sweeper) del(path string) error {
	req, _ := http.NewRequestWithContext(context.Background(), "DELETE", s.apiURL+path, nil)
	req.Header.Set("Authorization", "token "+s.token)
	req.Header.Set("Accept", "application/vnd.pulumi+9")
	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 && resp.StatusCode != 404 {
		return fmt.Errorf("%s: HTTP %d", path, resp.StatusCode)
	}
	return nil
}

func (s *sweeper) sweepOidcIssuers() {
	var resp struct {
		OidcIssuers []struct {
			ID      string    `json:"id"`
			Name    string    `json:"name"`
			Created sweepTime `json:"created"`
		} `json:"oidcIssuers"`
	}
	if err := s.get("/api/orgs/"+s.org+"/oidc/issuers", &resp); err != nil {
		s.t.Logf("sweep oidc list: %v", err)
		return
	}
	for _, iss := range resp.OidcIssuers {
		if !shouldSweep(iss.Name, iss.Created, s.now) {
			continue
		}
		if err := s.del("/api/orgs/" + s.org + "/oidc/issuers/" + iss.ID); err != nil {
			s.t.Logf("sweep oidc %s: %v", iss.Name, err)
			continue
		}
		s.t.Logf("sweep: deleted oidc issuer %s", iss.Name)
	}
}

func (s *sweeper) sweepRoles() {
	var resp struct {
		Roles []struct {
			ID      string    `json:"id"`
			Name    string    `json:"name"`
			Created sweepTime `json:"created"`
		} `json:"roles"`
	}
	if err := s.get("/api/orgs/"+s.org+"/roles?uxPurpose=role", &resp); err != nil {
		s.t.Logf("sweep roles list: %v", err)
		return
	}
	for _, r := range resp.Roles {
		if !shouldSweep(r.Name, r.Created, s.now) {
			continue
		}
		if err := s.del("/api/orgs/" + s.org + "/roles/" + r.ID); err != nil {
			s.t.Logf("sweep role %s: %v", r.Name, err)
			continue
		}
		s.t.Logf("sweep: deleted role %s", r.Name)
	}
}

func (s *sweeper) sweepAgentPools() {
	var resp struct {
		AgentPools []struct {
			ID      string    `json:"id"`
			Name    string    `json:"name"`
			Created sweepTime `json:"created"`
		} `json:"agentPools"`
	}
	if err := s.get("/api/orgs/"+s.org+"/agent-pools", &resp); err != nil {
		s.t.Logf("sweep agent-pools list: %v", err)
		return
	}
	for _, p := range resp.AgentPools {
		if !shouldSweep(p.Name, p.Created, s.now) {
			continue
		}
		if err := s.del("/api/orgs/" + s.org + "/agent-pools/" + p.ID); err != nil {
			s.t.Logf("sweep agent-pool %s: %v", p.Name, err)
			continue
		}
		s.t.Logf("sweep: deleted agent-pool %s", p.Name)
	}
}

func (s *sweeper) sweepOrgTokens() {
	var resp struct {
		Tokens []struct {
			ID      string    `json:"id"`
			Name    string    `json:"name"`
			Created sweepTime `json:"created"`
		} `json:"tokens"`
	}
	if err := s.get("/api/orgs/"+s.org+"/tokens", &resp); err != nil {
		s.t.Logf("sweep org-tokens list: %v", err)
		return
	}
	for _, tk := range resp.Tokens {
		if !shouldSweep(tk.Name, tk.Created, s.now) {
			continue
		}
		if err := s.del("/api/orgs/" + s.org + "/tokens/" + tk.ID); err != nil {
			s.t.Logf("sweep org-token %s: %v", tk.Name, err)
			continue
		}
		s.t.Logf("sweep: deleted org-token %s", tk.Name)
	}
}

func (s *sweeper) sweepEnvironments() {
	// ESC environments are listed under a different path. Same name-based
	// filter applies. Best-effort — projects with no envs return 404.
	var resp struct {
		Environments []struct {
			Project string    `json:"project"`
			Name    string    `json:"name"`
			Created sweepTime `json:"created"`
		} `json:"environments"`
	}
	if err := s.get("/api/esc/environments/"+s.org, &resp); err != nil {
		s.t.Logf("sweep envs list: %v", err)
		return
	}
	for _, e := range resp.Environments {
		if !shouldSweep(e.Name, e.Created, s.now) && !shouldSweep(e.Project, e.Created, s.now) {
			continue
		}
		path := "/api/esc/environments/" + s.org + "/" + e.Project + "/" + e.Name
		if err := s.del(path); err != nil {
			s.t.Logf("sweep env %s/%s: %v", e.Project, e.Name, err)
			continue
		}
		s.t.Logf("sweep: deleted env %s/%s", e.Project, e.Name)
	}
}
