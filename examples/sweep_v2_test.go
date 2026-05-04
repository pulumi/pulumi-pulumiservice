//go:build v2 || all
// +build v2 all

// Pre-test sweeper for the v2 test org. The TestV2 suite creates real
// resources on the configured Pulumi backend; if a test panics or the
// process is killed before t.Cleanup runs, those resources leak. They
// then collide with the next run (`already exists` 409s on
// non-idempotent endpoints, or shared singletons on the org).
//
// sweepOnce runs once per `go test` invocation, before TestV2 starts,
// and deletes any leftover resources matching the test-naming
// conventions used by this suite.
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

// sweepV2TestState removes leftover test resources from the configured
// Pulumi org. Best-effort: failures are logged, never fatal.
func sweepV2TestState(t *testing.T) {
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
		}
		s.sweepOidcIssuers()
		s.sweepOrgWebhooks()
		s.sweepPolicyGroups()
		s.sweepRoles()
		s.sweepTeams()
		s.sweepEnvironments()
		s.sweepAgentPools()
		s.sweepOrgTokens()
		s.sweepTemplateCollections()
	})
}

type sweeper struct {
	t      *testing.T
	token  string
	apiURL string
	org    string
	http   *http.Client
}

// shouldSweep returns true if a resource name looks like a leftover
// from the test suite. Test-suite naming conventions:
//
//   - suffix "-<5digits>" (random-suffix examples)
//   - prefixes: "v2-", "test-", "org-webhook-", "pulumiIssuer", etc.
//
// Conservative: only sweeps obvious test names so we don't nuke
// human-created resources by accident.
func (s *sweeper) shouldSweep(name string) bool {
	for _, pre := range []string{
		"v2-", "test-", "org-webhook-", "stack-webhook-",
		"env-webhook-", "brand-new-v2-", "example-policy-group",
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
	req.Header.Set("Accept", "application/vnd.pulumi+8")
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
	req.Header.Set("Accept", "application/vnd.pulumi+8")
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
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"oidcIssuers"`
	}
	if err := s.get("/api/orgs/"+s.org+"/oidc/issuers", &resp); err != nil {
		s.t.Logf("sweep oidc list: %v", err)
		return
	}
	for _, iss := range resp.OidcIssuers {
		if !s.shouldSweep(iss.Name) {
			continue
		}
		if err := s.del("/api/orgs/" + s.org + "/oidc/issuers/" + iss.ID); err != nil {
			s.t.Logf("sweep oidc %s: %v", iss.Name, err)
			continue
		}
		s.t.Logf("sweep: deleted oidc issuer %s", iss.Name)
	}
}

func (s *sweeper) sweepOrgWebhooks() {
	var resp []struct {
		Name string `json:"name"`
	}
	if err := s.get("/api/orgs/"+s.org+"/hooks", &resp); err != nil {
		s.t.Logf("sweep webhooks list: %v", err)
		return
	}
	for _, h := range resp {
		if !s.shouldSweep(h.Name) {
			continue
		}
		if err := s.del("/api/orgs/" + s.org + "/hooks/" + h.Name); err != nil {
			s.t.Logf("sweep webhook %s: %v", h.Name, err)
			continue
		}
		s.t.Logf("sweep: deleted webhook %s", h.Name)
	}
}

func (s *sweeper) sweepPolicyGroups() {
	var resp struct {
		PolicyGroups []struct {
			Name string `json:"name"`
		} `json:"policyGroups"`
	}
	if err := s.get("/api/orgs/"+s.org+"/policygroups", &resp); err != nil {
		s.t.Logf("sweep policygroups list: %v", err)
		return
	}
	for _, g := range resp.PolicyGroups {
		if !s.shouldSweep(g.Name) {
			continue
		}
		if err := s.del("/api/orgs/" + s.org + "/policygroups/" + g.Name); err != nil {
			s.t.Logf("sweep policygroup %s: %v", g.Name, err)
			continue
		}
		s.t.Logf("sweep: deleted policygroup %s", g.Name)
	}
}

func (s *sweeper) sweepRoles() {
	var resp struct {
		Roles []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"roles"`
	}
	if err := s.get("/api/orgs/"+s.org+"/roles", &resp); err != nil {
		s.t.Logf("sweep roles list: %v", err)
		return
	}
	for _, r := range resp.Roles {
		if !s.shouldSweep(r.Name) {
			continue
		}
		if err := s.del("/api/orgs/" + s.org + "/roles/" + r.ID); err != nil {
			s.t.Logf("sweep role %s: %v", r.Name, err)
			continue
		}
		s.t.Logf("sweep: deleted role %s", r.Name)
	}
}

func (s *sweeper) sweepTeams() {
	var resp struct {
		Teams []struct {
			Name string `json:"name"`
		} `json:"teams"`
	}
	if err := s.get("/api/orgs/"+s.org+"/teams", &resp); err != nil {
		s.t.Logf("sweep teams list: %v", err)
		return
	}
	for _, t := range resp.Teams {
		if !s.shouldSweep(t.Name) {
			continue
		}
		if err := s.del("/api/orgs/" + s.org + "/teams/" + t.Name); err != nil {
			s.t.Logf("sweep team %s: %v", t.Name, err)
			continue
		}
		s.t.Logf("sweep: deleted team %s", t.Name)
	}
}

func (s *sweeper) sweepAgentPools() {
	var resp struct {
		AgentPools []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"agentPools"`
	}
	if err := s.get("/api/orgs/"+s.org+"/agent-pools", &resp); err != nil {
		s.t.Logf("sweep agent-pools list: %v", err)
		return
	}
	for _, p := range resp.AgentPools {
		if !s.shouldSweep(p.Name) {
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
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"tokens"`
	}
	if err := s.get("/api/orgs/"+s.org+"/tokens", &resp); err != nil {
		s.t.Logf("sweep org-tokens list: %v", err)
		return
	}
	for _, tk := range resp.Tokens {
		if !s.shouldSweep(tk.Name) {
			continue
		}
		if err := s.del("/api/orgs/" + s.org + "/tokens/" + tk.ID); err != nil {
			s.t.Logf("sweep org-token %s: %v", tk.Name, err)
			continue
		}
		s.t.Logf("sweep: deleted org-token %s", tk.Name)
	}
}

func (s *sweeper) sweepTemplateCollections() {
	var resp struct {
		Sources []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"sources"`
	}
	if err := s.get("/api/orgs/"+s.org+"/templates/sources", &resp); err != nil {
		s.t.Logf("sweep template-collections list: %v", err)
		return
	}
	for _, c := range resp.Sources {
		if !s.shouldSweep(c.Name) {
			continue
		}
		if err := s.del("/api/orgs/" + s.org + "/templates/sources/" + c.ID); err != nil {
			s.t.Logf("sweep template-collection %s: %v", c.Name, err)
			continue
		}
		s.t.Logf("sweep: deleted template-collection %s", c.Name)
	}
}

func (s *sweeper) sweepEnvironments() {
	// ESC environments are listed under a different path. Same name-based
	// filter applies. Best-effort — projects with no envs return 404.
	var resp struct {
		Environments []struct {
			Project string `json:"project"`
			Name    string `json:"name"`
		} `json:"environments"`
	}
	if err := s.get("/api/esc/environments/"+s.org, &resp); err != nil {
		s.t.Logf("sweep envs list: %v", err)
		return
	}
	for _, e := range resp.Environments {
		if !s.shouldSweep(e.Name) && !s.shouldSweep(e.Project) {
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
