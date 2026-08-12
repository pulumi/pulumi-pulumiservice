# Issue #938: api.Role refresh drift not detected

## Plan

- [x] Verify checkout state; create `fix-api-refresh-drift` from `origin/main`
- [x] Root cause: `rest.Resource.Read` returns stored inputs verbatim; engine refresh diff is input-driven, so drift never surfaces
- [x] Failing test first: `TestReadRefreshProjectsRemoteDriftIntoInputs` (red)
- [x] Fix: project freshly-read state onto existing input keys in `Read` (green)
- [x] Chain test with real spec + `api:Role`: `TestRefreshChainSurfacesRoleDetailsDrift` (drift + quiet cases)
- [x] Full provider unit suite, gofmt, go vet: clean
- [x] CHANGELOG entry under Unreleased / Bug Fixes

## Review

Root cause: two layers hid drift for every `pulumiservice:api:*` resource.
1. `Read` preserved `req.Inputs` verbatim on refresh. The engine (pulumi >= 3.x,
   non-legacy path) computes the refresh diff via provider `Diff` between the
   Read-returned inputs and the old inputs, so it always saw "no changes".
2. `Diff` compares only inputs, so `pulumi up` could not see output drift either.

Fix: `projectStateOntoInputs` in `provider/pkg/rest/resource.go`. On refresh,
keys already present in inputs take the freshly-read server value (with Check
normalization: enum case fold, Unordered sort, secret marking preserved).
Write-only fields (absent from GET) keep prior values. Response-only fields
stay out of inputs, which preserves the "+issuerId" no-noise guarantee.

Matches the pattern every legacy hand-written resource already uses
(OrganizationRole, Environment). Import path (empty inputs) unchanged.

Blast radius guard: every example integration test runs
`assertrefresh.HasNoChanges` right after `up`, so CI flags any api resource
whose GET response does not round-trip its inputs.

Not done: live end-to-end repro (the `.env` review-stack token returns 401).
CI example tests (up → refresh no-changes) served as the end-to-end pass.

Shipped as PR #1031. CI surfaced and the PR also fixes: phantom drift from
server-added map entries (system stack tags; fixed with key-scoped recursion)
and a pre-existing silent no-op in api.teams.Team updates (UpdateTeamRequest
newDescription/newDisplayName never sent; fixed with flat new<Field> aliases,
also covering PolicyGroup/AuditLogExportConfiguration/SAMLOrganization).
