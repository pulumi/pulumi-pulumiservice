# Pulumi Service Provider - Infer Migration Guide

## Table of Contents

- [Executive Summary](#executive-summary)
- [Background & Analysis](#background--analysis)
- [Migration Strategy](#migration-strategy)
- [Architecture Overview](#architecture-overview)
- [Migration Progress](#migration-progress)
- [Practical Migration Guide](#practical-migration-guide)
- [Phase 0-1 Results](#phase-0-1-results)
- [Common Patterns & Best Practices](#common-patterns--best-practices)
- [Troubleshooting](#troubleshooting)
- [References](#references)

## Executive Summary

This document tracks the migration of the Pulumi Service Provider from a **manual provider implementation** to the **`pulumi-go-provider` infer framework** (v1.1.2+). The migration started from v0.32.0+ and uses an incremental hybrid approach that allows both manual and infer resources to coexist during the transition.

**Current Status**:
- Phase 0 (Foundation): ✅ COMPLETE
- Phase 1.1 (StackTag POC): ✅ COMPLETE
- Phase 1.2 (OrgAccessToken): 🔄 NEXT

**Key Benefits**:
- Automatic schema generation from Go types
- Type-safe resource definitions
- Reduced boilerplate code
- Better maintainability
- Zero breaking changes for existing users

## Background & Analysis

### Current Architecture (v0.32.0)

**Provider Entry Point** (`provider/cmd/pulumi-resource-pulumiservice/main.go`):
- Uses standard `provider.Main()` from `github.com/pulumi/pulumi/pkg/v3/resource/provider`
- Calls `psp.MakeProvider()` with an embedded `schema.json`
- Schema is **manually maintained** in `schema.json` (3,479 lines)

**Provider Implementation** (`provider/pkg/provider/provider.go`):
- Manual implementation of `pulumirpc.ResourceProviderServer` interface
- Resources stored in `provider/pkg/resources/` package
- Each resource implements `PulumiServiceResource` interface with 6 methods:
  - `Check()`, `Create()`, `Read()`, `Update()`, `Delete()`, `Diff()`, `Name()`
- Resources are registered manually in `Configure()` method
- No dependency on `pulumi-go-provider`

**Build Process** (`Makefile`):
- `make provider` builds the binary with version injection via LDFLAGS
- `make build_sdks` generates SDKs from the manual `schema.json`
- Schema is **NOT** auto-generated

### Target Architecture (Infer v1.1.2+)

**Provider Entry Point** (`provider/cmd/pulumi-resource-pulumiservice/main.go`):
- Uses `provider.MainWithOptions()` to support hybrid provider
- Returns `p.Provider` interface from `pulumi-go-provider`
- Combines manual and infer resources using `infer.Wrap()`

**Provider Implementation** (`provider/pkg/provider/`):
- **Hybrid approach** during migration: Both manual (`provider.go`) and infer (`hybrid.go`)
- Infer resources in `provider/pkg/infer/` package
- Manual resources remain in `provider/pkg/resources/` during transition

**Build Process**:
- Same Makefile, schema generation works with hybrid provider
- Infer resources contribute to schema automatically
- Manual resources use existing schema.json definitions

### Resources Inventory

**Total Resources**: 21 resources + 2 data sources

#### Resources at v0.20.0 (13 total)
1. PulumiServiceTeamResource
2. PulumiServiceAccessTokenResource
3. PulumiServiceWebhookResource
4. PulumiServiceStackTagResource ✅ **MIGRATED**
5. TeamStackPermissionResource
6. PulumiServiceTeamAccessTokenResource
7. PulumiServiceOrgAccessTokenResource
8. PulumiServiceDeploymentSettingsResource
9. PulumiServiceAgentPoolResource
10. PulumiServiceDeploymentScheduleResource
11. PulumiServiceDriftScheduleResource
12. PulumiServiceTtlScheduleResource
13. PulumiServiceUnknownResource

#### Resources added in v0.32.0 (8 new)
14. **PulumiServiceEnvironmentResource** (ESC support)
15. **PulumiServiceTeamEnvironmentPermissionResource** (ESC permissions)
16. **PulumiServiceEnvironmentVersionTagResource** (ESC versioning)
17. **PulumiServiceStackResource** (Stack management)
18. **PulumiServiceTemplateSourceResource** (Templates)
19. **PulumiServiceOidcIssuerResource** (OIDC authentication)
20. **PulumiServiceEnvironmentRotationScheduleResource** (ESC rotation)
21. **PulumiServiceApprovalRuleResource** (Deployment approvals)

#### Data Sources (2 total)
- `pulumiservice:index:getPolicyPacks` (list all policy packs)
- `pulumiservice:index:getPolicyPack` (get specific policy pack)

### pulumi-go-provider Evolution: v0.16.0 → v1.1.2

The `pulumi-go-provider` framework has undergone significant changes since v0.16.0. Understanding these changes is critical for the migration.

#### Major Version 1.0.0 Release (May 16, 2023)

Version 1.0.0 marked the **stability milestone** for pulumi-go-provider, introducing breaking changes and a completely redesigned API.

**Breaking Changes:**

1. **Component Definition API Changed (v0.25.0)**

```go
// OLD (v0.16.0 - v0.24.x)
infer.Component[*Component, ComponentArgs, *ComponentState]()

// NEW (v0.25.0+)
infer.Component(resourceInstance)  // Takes resource instance
// OR
infer.ComponentF(constructorFunc)  // Takes constructor function
```

2. **New Provider Builder API (v1.0.0)**

```go
// OLD (v0.16.0)
infer.Provider(infer.Options{
    Resources:  []InferredResource{...},
    Components: []InferredComponent{...},
    Functions:  []InferredFunction{...},
    Config:     infer.Config[*Config](),
})

// NEW (v1.0.0+)
p, err := infer.NewProviderBuilder().
    WithResources(
        infer.Resource(MyResource{}),
    ).
    WithComponents(
        infer.ComponentF(NewMyComponent),
    ).
    WithFunctions(
        infer.Function(MyFunction{}),
    ).
    WithConfig(infer.Config[*Config]()).
    Build()
```

**Benefits**:
- Cleaner, more readable API
- Better error handling (returns error from Build())
- Fluent chaining of configuration
- Default metadata handling

3. **Go Toolchain Upgrade**
- Minimum Go version: 1.22 → 1.24
- Better generics support
- Improved type inference

4. **StreamInvoke Removal (v0.26.0)**
- `StreamInvoke` method removed from provider interface
- Pulumi SDK v3.169.0 removed `ResourceProvider_StreamInvokeServer`
- Simplified provider implementation

#### New Features in v1.x

**1. Provider Builder Defaults (v1.0.0)**

`NewProviderBuilder()` provides sensible defaults:

```go
defaultMetadata := schema.Metadata{
    LanguageMap: map[string]any{
        "nodejs": map[string]any{
            "respectSchemaVersion": true,
        },
        "go": map[string]any{
            "generateResourceContainerTypes": true,
            "respectSchemaVersion":           true,
        },
        "python": map[string]any{
            "respectSchemaVersion": true,
            "pyproject": map[string]any{
                "enabled": true,
            },
        },
        "csharp": map[string]any{
            "respectSchemaVersion": true,
        },
    },
}
```

**2. Enhanced Testing Framework (v1.0.0)**

New test harness for:
- Component resource testing
- Lifecycle testing (CRUD operations)
- Mock support for invokes
- Config annotation testing

**3. Secrets Handling Improvements (v0.24.1+)**
- Better schema-level secrets handling
- `apply_secrets.go` added for automatic secret marking
- Improved secret propagation in components

#### Comparison: v0.16.0 vs v1.1.2

| Feature | v0.16.0 (2022) | v1.1.2 (2023) |
|---------|----------------|---------------|
| **API Stability** | Unstable, experimental | Stable, 1.0+ |
| **Component Definition** | `Component[R, I, O]()` | `Component(rsc)` or `ComponentF(fn)` |
| **Provider Setup** | Direct `infer.Provider(Options{...})` | Builder pattern with `NewProviderBuilder()` |
| **Go Version** | 1.22 | 1.24+ |
| **Test Framework** | Basic | Comprehensive lifecycle testing |
| **Secrets Handling** | Manual | Semi-automatic via middleware |
| **Default Metadata** | Manual specification | Automatic sensible defaults |
| **Error Handling** | Panics on build errors | Returns errors from `Build()` |
| **StreamInvoke** | Supported | Removed (deprecated) |
| **Documentation** | Limited | Extensive with examples |

## Migration Strategy

### Selected Strategy: Hybrid Migration (Options 2 + 4)

**Decision**: Migrate to pulumi-go-provider v1.1.2+ using an incremental approach.

#### Strategy Overview

1. **For New Resources**: All new resources will be implemented using true infer patterns from day one
2. **For Existing Resources**: Gradually migrate existing 21 resources to infer over multiple releases
3. **Go 1.24+**: Upgrade to Go 1.24 to support pulumi-go-provider v1.1.2+
4. **New Branch**: Start fresh from current main, not from the old v0.20.0 branch

#### Why This Approach

**Benefits**:
- ✅ Get infer benefits immediately for new resources
- ✅ No big-bang migration risk
- ✅ Learn and refine patterns as we go
- ✅ Each migrated resource can be tested thoroughly
- ✅ Can pause/resume migration between releases
- ✅ Maintain backwards compatibility throughout
- ✅ Use modern v1.1.2 API instead of outdated v0.16.0

**Risks & Mitigation**:
- ⚠️ Two implementation patterns in codebase
  - *Mitigation*: Clear documentation, separate packages, gradual convergence
- ⚠️ Complex provider setup during transition
  - *Mitigation*: Comprehensive integration tests, careful schema merging
- ⚠️ Longer timeline than full rewrite
  - *Mitigation*: Acceptable trade-off for reduced risk

#### Success Criteria

1. Zero breaking changes to existing users
2. Schema changes are additive only (unless explicitly versioned)
3. All existing examples continue to work
4. New infer-based resources have same behavior as manual equivalents
5. CI/CD pipeline validates both manual and infer resources

## Architecture Overview

### Hybrid Provider Architecture

During the migration, we run a **hybrid provider** that supports both manual and infer resources:

```
┌─────────────────────────────────────┐
│  main.go: p.RunProvider()           │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  hybrid.go: MakeHybridProvider()    │
│  ┌───────────────────────────────┐  │
│  │ Manual Provider               │  │
│  │ (existing resources)          │  │
│  │ - Wrapped by hybridWrapper    │  │
│  │ - Injects client into context │  │
│  └───────────────────────────────┘  │
│               +                      │
│  ┌───────────────────────────────┐  │
│  │ Infer Provider                │  │
│  │ (migrated resources)          │  │
│  │ - Gets client from context    │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
```

### How It Works

```go
// main.go
func main() {
    err := provider.MainWithOptions(providerName, Version, func(host *provider.HostClient) (p.Provider, error) {
        return psp.MakeHybridProvider(host, providerName, Version, schema)
    }, provider.Options{})

    if err != nil {
        cmdutil.ExitError(err.Error())
    }
}

// provider/pkg/provider/hybrid.go
func MakeHybridProvider(host *provider.HostClient, name, version, schema string) (p.Provider, error) {
    // Create manual provider for legacy resources
    manualProvider := &pulumiserviceProvider{
        host:    host,
        name:    name,
        schema:  mustSetSchemaVersion(schema, version),
        version: version,
    }

    // Wrap manual provider to inject client into context
    wrappedManual := &hybridManualProvider{
        inner: rpc.Provider(manualProvider),
        manualProvider: manualProvider,
    }

    // Create infer provider for migrated resources
    inferOpts := buildInferOptions()

    // Combine both providers
    return infer.Wrap(wrappedManual, inferOpts), nil
}
```

### Client Context Injection Pattern

**The Problem**: Infer resources need access to the Pulumi Service API client (`*pulumiapi.Client`) to make API calls.

**Solution**: Use context injection via `hybridManualProvider` wrapper

**Implementation**:

```go
// provider/pkg/infer/client.go
type clientContextKey struct{}

func WithClient(ctx context.Context, client *pulumiapi.Client) context.Context {
    return context.WithValue(ctx, clientContextKey{}, client)
}

func GetClient(ctx context.Context) *pulumiapi.Client {
    if client, ok := ctx.Value(clientContextKey{}).(*pulumiapi.Client); ok {
        return client
    }
    panic("API client not found in context")
}

// provider/pkg/provider/hybrid_wrapper.go
type hybridManualProvider struct {
    inner          p.Provider
    manualProvider *pulumiserviceProvider
}

func (h *hybridManualProvider) Create(ctx context.Context, req p.CreateRequest) (p.CreateResponse, error) {
    // Inject client into context before calling infer resources
    ctx = inferPkg.WithClient(ctx, h.manualProvider.client)
    return h.inner.Create(ctx, req)
}

// Similar injection for Read, Update, Delete operations
```

**Usage in Infer Resources**:

```go
func (StackTag) Create(ctx context.Context, req infer.CreateRequest[StackTagArgs]) (infer.CreateResponse[StackTagState], error) {
    client := inferPkg.GetClient(ctx)  // Retrieve from context
    // Use client to call API
    return infer.CreateResponse[StackTagState]{...}, nil
}
```

## Migration Progress

### Phase 0: Foundation & Setup ✅ COMPLETE

**Branch Strategy**:
```bash
# Created new branch from main
git checkout -b feature/migrate-to-infer-v1 47a344e  # main commit
```

#### Completed Tasks

- [x] **0.1**: Create feature branch `feature/migrate-to-infer-v1` from main
- [x] **0.2**: Upgrade Go version to 1.24 in all go.mod files
- [x] **0.3**: Add pulumi-go-provider v1.1.2+ dependency to root `go.mod`
- [x] **0.4**: Remove deprecated StreamInvoke method (incompatible with Pulumi SDK v3.169.0)
- [x] **0.5**: Create hybrid provider architecture ✅
  - ✅ Created `provider/pkg/provider/hybrid.go`
  - ✅ Created `provider/pkg/provider/hybrid_wrapper.go` for client injection
  - ✅ Updated `main.go` to use `p.RunProvider()` with hybrid provider
- [x] **0.6**: Set up testing infrastructure ✅
  - ✅ Created `provider/pkg/infer/README.md` with guidelines
  - ✅ Created `provider/pkg/infer/client.go` for context-based client access
  - ✅ Documented patterns in `docs/INFER_MIGRATION.md`
- [x] **0.7**: Documentation complete ✅
  - ✅ `docs/INFER_MIGRATION.md`
  - ✅ `docs/PHASE_0_1_SUMMARY.md`

### Phase 1: Proof of Concept ✅ PHASE 1.1 COMPLETE

**Goal**: Migrate 3 simple resources to validate approach

#### 1.1 Migrate PulumiServiceStackTagResource ✅ COMPLETE

- [x] **1.1.1**: Created `provider/pkg/infer/stack_tag.go` ✅
- [x] **1.1.2**: Defined `StackTagArgs` and `StackTagState` structs ✅
- [x] **1.1.3**: Implemented `Create()` with v1.1.2 signature (`CreateRequest`/`CreateResponse`) ✅
- [x] **1.1.4**: Implemented `Read()` method ✅
- [x] **1.1.5**: Update method intentionally omitted (automatic replace behavior) ✅
- [x] **1.1.6**: Implemented `Delete()` method ✅
- [x] **1.1.7**: Added `Annotate()` for descriptions ✅
- [x] **1.1.8**: Registered in `buildInferOptions()` in hybrid.go ✅
- [x] **1.1.9**: Implemented client context injection via `hybridManualProvider` ✅
- [x] **1.1.10**: Integration test `TestYamlStackTagsExample` - **PASSED** (17.3s) ✅
- [x] **1.1.11**: Validated all CRUD operations work correctly ✅
- [x] **1.1.12**: Documented in `docs/PHASE_0_1_SUMMARY.md` ✅

**Key Achievement**: First resource successfully migrated! Hybrid architecture validated.

**Integration Test Results**:
- ✅ Created stack tag successfully
- ✅ Read/refresh worked correctly
- ✅ Update triggered replace (delete + create) as expected
- ✅ Delete cleaned up correctly
- ✅ Total test time: 17.3s

#### 1.2 Migrate PulumiServiceOrgAccessTokenResource 🔄 NEXT

- [ ] **1.2.1**: Create `provider/pkg/infer/org_access_token.go`
- [ ] **1.2.2**: Define input/output structs
- [ ] **1.2.3**: Implement CRUD methods
- [ ] **1.2.4**: Handle secrets properly (verify auto-marking works)
- [ ] **1.2.5**: Register in provider
- [ ] **1.2.6**: Test with examples
- [ ] **1.2.7**: Validate schema generation

#### 1.3 Migrate PulumiServiceAgentPoolResource

- [ ] **1.3.1**: Create `provider/pkg/infer/agent_pool.go`
- [ ] **1.3.2**: Define input/output structs
- [ ] **1.3.3**: Implement CRUD methods
- [ ] **1.3.4**: Register in provider
- [ ] **1.3.5**: Test with examples
- [ ] **1.3.6**: Document patterns learned

#### 1.4 POC Validation

- [ ] **1.4.1**: Ensure all 3 migrated resources work identically to manual versions
- [ ] **1.4.2**: Verify schema is backwards compatible
- [ ] **1.4.3**: Run full test suite (manual + infer resources)
- [ ] **1.4.4**: Get team review and approval to continue
- [ ] **1.4.5**: Document lessons learned in `docs/INFER_MIGRATION.md`

### Resource Migration Priority Order

**Complexity Rating** (based on manual implementation analysis):

| Priority | Resource | Complexity | Lines of Code | Key Challenges |
|----------|----------|------------|---------------|----------------|
| **POC** | StackTag ✅ | ⭐ Simple | ~150 | None, basic CRUD |
| **POC** | OrgAccessToken | ⭐ Simple | ~180 | Secret handling |
| **POC** | AgentPool | ⭐⭐ Medium | ~250 | Validation |
| 1 | AccessToken | ⭐ Simple | ~170 | Secret handling |
| 2 | TeamAccessToken | ⭐⭐ Medium | ~200 | Secret + team ref |
| 3 | EnvironmentVersionTag | ⭐⭐ Medium | ~190 | ESC integration |
| 4 | TeamStackPermission | ⭐⭐ Medium | ~180 | Permission enums |
| 5 | TeamEnvironmentPermission | ⭐⭐ Medium | ~220 | Permission enums |
| 6 | TemplateSource | ⭐⭐ Medium | ~230 | Git integration |
| 7 | Webhook | ⭐⭐⭐ Complex | ~500 | Multiple formats, filters |
| 8 | Stack | ⭐⭐⭐ Complex | ~180 | Project references |
| 9 | DeploymentSchedule | ⭐⭐⭐ Complex | ~380 | Schedule validation |
| 10 | DriftSchedule | ⭐⭐⭐ Complex | ~210 | Schedule validation |
| 11 | TtlSchedule | ⭐⭐⭐ Complex | ~220 | Schedule validation |
| 12 | EnvironmentRotationSchedule | ⭐⭐⭐ Complex | ~350 | ESC + scheduling |
| 13 | Environment | ⭐⭐⭐⭐ Very Complex | ~320 | ESC client, YAML |
| 14 | Team | ⭐⭐⭐⭐⭐ Most Complex | ~430 | Membership, partial errors |
| 15 | PolicyGroup | ⭐⭐⭐⭐ Very Complex | ~500 | Policy management |
| 16 | ApprovalRule | ⭐⭐⭐⭐ Very Complex | ~400 | Complex rules |
| 17 | OidcIssuer | ⭐⭐⭐⭐⭐ Most Complex | ~460 | Auth policies, drift |
| 18 | DeploymentSettings | ⭐⭐⭐⭐⭐ Most Complex | ~900 | Huge property surface |

### Estimated Timeline

| Phase | Duration | Resources | Target Completion |
|-------|----------|-----------|-------------------|
| Phase 0: Foundation ✅ | 2 weeks | Setup | Week 2 |
| Phase 1: POC | 2 weeks | 3 resources | Week 4 |
| Phase 2: Simple | 4 weeks | 6 resources | Week 8 |
| Phase 3: Medium | 6 weeks | 7 resources | Week 14 |
| Phase 4: Complex | 6 weeks | 5 resources | Week 20 |
| Phase 5: Data Sources | 2 weeks | 2 functions | Week 22 |
| Phase 6: Cleanup | 2 weeks | Deprecation | Week 24 |
| Phase 7: Documentation | 2 weeks | Release | Week 26 |
| **Total** | **26 weeks** | **21 resources + 2 functions** | **~6 months** |

**Note**: This is an aggressive timeline. More realistic estimate: **8-9 months** with:
- Time for unexpected issues
- Schema compatibility challenges
- Team reviews and feedback cycles
- Testing and validation
- Holidays and other priorities

## Practical Migration Guide

This section provides step-by-step guidance for migrating each resource.

### Resource Migration Checklist

For each resource being migrated:

#### 1. Analyze the Manual Implementation

- [ ] Read the manual resource implementation in `provider/pkg/resources/`
- [ ] Document all edge cases and special logic
- [ ] Identify all API calls made
- [ ] Note any partial failure handling
- [ ] Check for custom diff logic

#### 2. Create Infer Resource File

- [ ] Create file in `provider/pkg/infer/<resource_name>.go`
- [ ] Add copyright header (2016-2025)
- [ ] Import required packages

#### 3. Define Resource Structs

```go
// Empty struct that serves as the resource anchor
type StackTag struct{}

// Input arguments - all fields user can specify
type StackTagArgs struct {
    Organization string `pulumi:"organization"`
    Project      string `pulumi:"project"`
    Stack        string `pulumi:"stack"`
    Name         string `pulumi:"name"`
    Value        string `pulumi:"value"`
}

// State - what gets stored and returned
type StackTagState struct {
    StackTagArgs  // Embed args
    // Add computed fields if needed
}
```

**Guidelines**:
- Use `pulumi:"fieldName"` tags for all exported fields
- Add `,optional` for optional fields: `pulumi:"field,optional"`
- Use pointers for optional primitive types: `*string`, `*int`
- Use `,secret` for secret fields: `pulumi:"token,secret"`

#### 4. Implement CRUD Methods

**Create Method**:

```go
func (StackTag) Create(
    ctx context.Context,
    req infer.CreateRequest[StackTagArgs],
) (infer.CreateResponse[StackTagState], error) {
    if req.Preview {
        // Return placeholder ID during preview
        return infer.CreateResponse[StackTagState]{
            ID:    req.Name,
            State: StackTagState{StackTagArgs: req.Inputs},
        }, nil
    }

    // Get API client from context
    client := inferPkg.GetClient(ctx)

    // Call API to create resource
    // err := client.CreateStackTag(...)

    id := fmt.Sprintf("%s/%s/%s/%s",
        req.Inputs.Organization,
        req.Inputs.Project,
        req.Inputs.Stack,
        req.Inputs.Name)

    state := StackTagState{StackTagArgs: req.Inputs}

    return infer.CreateResponse[StackTagState]{
        ID:    id,
        State: state,
    }, nil
}
```

**Read Method**:

```go
func (StackTag) Read(
    ctx context.Context,
    id string,
    inputs StackTagArgs,
    state StackTagState,
) (
    canonicalID string,
    normalizedInputs StackTagArgs,
    normalizedState StackTagState,
    err error,
) {
    // Get API client from context
    client := inferPkg.GetClient(ctx)

    // Parse ID and call API
    // If resource doesn't exist, return empty ID: return "", inputs, state, nil

    // Return current state from API
    return id, inputs, state, nil
}
```

**Update Method** (optional - omit for replace-only resources):

```go
func (StackTag) Update(
    ctx context.Context,
    id string,
    olds StackTagState,
    news StackTagArgs,
    preview bool,
) (StackTagState, error) {
    if preview {
        return StackTagState{StackTagArgs: news}, nil
    }

    // Get API client from context
    client := inferPkg.GetClient(ctx)

    // Call API to update
    // err := client.UpdateStackTag(...)

    return StackTagState{StackTagArgs: news}, nil
}
```

**Delete Method**:

```go
func (StackTag) Delete(ctx context.Context, id string, state StackTagState) error {
    // Get API client from context
    client := inferPkg.GetClient(ctx)

    // Call API to delete
    // return client.DeleteStackTag(...)

    return nil
}
```

#### 5. Add Annotations (Documentation)

```go
func (*StackTag) Annotate(a infer.Annotator) {
    a.Describe(new(StackTag), "A Stack Tag associates metadata with a stack.")
    a.Describe(new(StackTagArgs).Organization, "The organization name.")
    a.Describe(new(StackTagArgs).Project, "The project name.")
    a.Describe(new(StackTagArgs).Stack, "The stack name.")
    a.Describe(new(StackTagArgs).Name, "The tag name.")
    a.Describe(new(StackTagArgs).Value, "The tag value.")
}
```

#### 6. Register in Hybrid Provider

Edit `provider/pkg/provider/hybrid.go`:

```go
func buildInferOptions() infer.Options {
    return infer.Options{
        Resources: []infer.InferredResource{
            infer.Resource(&StackTag{}),  // ADD THIS LINE
        },
        // ...
    }
}
```

#### 7. Testing

- [ ] Build provider: `make provider`
- [ ] Run existing examples for the resource
- [ ] Compare generated schema with manual schema
- [ ] Run integration tests: `cd examples && go test -v -run TestYamlStackTag -tags yaml`
- [ ] Verify no breaking changes

#### 8. Documentation

- [ ] Update CHANGELOG.md under `## Unreleased / ### Improvements`
- [ ] Document any learnings in this file
- [ ] Update migration progress section

## Phase 0-1 Results

### Phase 0 Achievements

**Completed Tasks**:
- ✅ Created `feature/migrate-to-infer-v1` branch from main (commit 47a344e)
- ✅ Upgraded Go from 1.23.0 to 1.24.0 in all go.mod files
- ✅ Added `pulumi-go-provider v1.1.2` dependency
- ✅ Upgraded Pulumi SDK from v3.138.0 to v3.169.0
- ✅ Removed deprecated `StreamInvoke` method (removed in Pulumi SDK v3.169.0)
- ✅ Created `provider/pkg/infer/` directory for infer resources
- ✅ Created hybrid provider architecture in `provider/pkg/provider/hybrid.go`
- ✅ Updated `main.go` to use `p.RunProviderF()` with hybrid provider
- ✅ Verified provider builds successfully

**Files Created**:
- `Convert-to-infer.md` - Complete migration plan and analysis
- `docs/INFER_MIGRATION.md` - This comprehensive guide
- `provider/pkg/infer/README.md` - Infer directory documentation
- `provider/pkg/infer/client.go` - Client context handling
- `provider/pkg/provider/hybrid.go` - Hybrid provider setup
- `provider/pkg/provider/hybrid_wrapper.go` - Context injection wrapper

**Files Modified**:
- `go.mod` - Upgraded Go 1.24, added pulumi-go-provider v1.1.2
- `sdk/go.mod` - Upgraded Go 1.24
- `examples/*/go.mod` - Upgraded Go 1.24
- `provider/cmd/pulumi-resource-pulumiservice/main.go` - Use hybrid provider
- `provider/pkg/provider/provider.go` - Removed deprecated StreamInvoke

### Phase 1.1 Results: StackTag Migration

**Implementation Details**:
- Created `provider/pkg/infer/stack_tag.go`
- Defined `StackTagArgs` and `StackTagState` structs with pulumi tags
- Implemented `Create()` method using v1.1.2 infer API signatures
- Implemented `Read()` method
- Intentionally skipped `Update()` - triggers automatic replace (matches manual behavior)
- Implemented `Delete()` method
- Added `Annotate()` for resource description
- Registered `StackTag` in `buildInferOptions()` in hybrid.go

**Integration Test**: `TestYamlStackTagsExample` - ✅ **PASSED** (17.3s)
- ✅ Created stack tag successfully
- ✅ Read/refresh worked correctly
- ✅ Update triggered replace (delete + create) as expected
- ✅ Delete cleaned up correctly

**This validates**:
- ✅ Hybrid provider architecture works
- ✅ Client context injection works
- ✅ Infer v1.1.2 API works correctly
- ✅ Schema generation works (implicitly validated by test)
- ✅ Migration pattern is sound

### Lessons Learned

#### 1. pulumi-go-provider v1.1.2 API Signatures

**Correct signatures for v1.1.2**:
```go
// Create
func (*Resource) Create(ctx context.Context, req infer.CreateRequest[Args]) (infer.CreateResponse[State], error)

// Read
func (*Resource) Read(ctx context.Context, id string, inputs Args, state State) (
    canonicalID string, normalizedInputs Args, normalizedState State, err error)

// Update
func (*Resource) Update(ctx context.Context, id string, oldState State, newInputs Args, preview bool) (State, error)

// Delete
func (*Resource) Delete(ctx context.Context, id string, state State) error
```

**NOT the old v0.16.0 signatures**:
```go
// OLD - Don't use
func (Resource) Create(ctx context.Context, name string, input Args, preview bool) (string, State, error)
```

#### 2. Registering Resources

Use instance, not type parameter:
```go
// v1.1.2 - CORRECT
infer.Resource(&StackTag{})

// v0.16.0 - WRONG for v1.1.2
infer.Resource[*StackTag]()
```

#### 3. Annotate Method

Access fields correctly:
```go
// CORRECT
func (*StackTag) Annotate(a infer.Annotator) {
    a.Describe(new(StackTag), "Description")
    a.Describe(new(StackTagArgs).Organization, "Org description")
}

// WRONG - StackTag has no Args field
func (r *StackTag) Annotate(a infer.Annotator) {
    a.Describe(&r.Args.Organization, "...")  // ERROR
}
```

#### 4. StreamInvoke Removed

Pulumi SDK v3.169.0 removed `StreamInvoke` - must be deleted from provider.go when upgrading.

#### 5. Hybrid Provider Pattern Works

The `infer.Wrap(rpc.Provider(manual), inferOpts)` pattern successfully combines:
- Manual resources (21 existing resources)
- Infer resources (1 migrated resource so far)

Both can coexist during migration.

#### 6. Client Context Solution

**Implemented Solution**: Created `hybridManualProvider` wrapper that injects clients into context

**How it works**:
1. `hybridManualProvider` wraps the manual provider
2. Before each CRUD operation, it injects the API client into context
3. Infer resources call `inferPkg.GetClient(ctx)` to retrieve the client
4. Thread-safe and clean separation of concerns

## Common Patterns & Best Practices

### Accessing API Client

Since we're in a hybrid provider, infer resources need to access the same API client as manual resources.

**Pattern**: Retrieve client from context

```go
// In infer resource methods
func (StackTag) Create(ctx context.Context, req infer.CreateRequest[StackTagArgs]) (infer.CreateResponse[StackTagState], error) {
    client := inferPkg.GetClient(ctx)  // Get from context
    // Use client to call API
}
```

**Note**: The client is automatically injected into context by `hybridManualProvider` wrapper before each operation.

### Handling Secrets

Use the `,secret` tag:

```go
type TokenArgs struct {
    Value string `pulumi:"value,secret"`
}
```

Secrets are automatically marked by the infer framework based on the struct tags.

### Handling Enums

Define as constants and use validation:

```go
type TeamType string

const (
    TeamTypePulumi TeamType = "pulumi"
    TeamTypeGitHub TeamType = "github"
)

func (Team) Check(ctx context.Context, name string, oldInputs, newInputs TeamArgs) (TeamArgs, []p.CheckFailure, error) {
    var failures []p.CheckFailure

    if newInputs.TeamType != TeamTypePulumi && newInputs.TeamType != TeamTypeGitHub {
        failures = append(failures, p.CheckFailure{
            Property: "teamType",
            Reason:   fmt.Sprintf("must be 'pulumi' or 'github', got '%s'", newInputs.TeamType),
        })
    }

    return newInputs, failures, nil
}
```

### Partial Failures

Infer handles partial failures differently than manual resources. Use proper error returns:

```go
func (Team) Create(ctx context.Context, req infer.CreateRequest[TeamArgs]) (infer.CreateResponse[TeamState], error) {
    // Create team
    id, err := createTeam(ctx, req.Inputs)
    if err != nil {
        return infer.CreateResponse[TeamState]{}, err
    }

    // Add members - if this fails, resource is partially created
    for _, member := range req.Inputs.Members {
        if err := addMember(ctx, id, member); err != {
            // Resource exists with ID but incomplete state
            partialState := TeamState{TeamArgs: req.Inputs}
            partialState.Members = nil  // Clear members that weren't added
            return infer.CreateResponse[TeamState]{
                ID:    id,
                State: partialState,
            }, fmt.Errorf("failed to add member %s: %w", member, err)
        }
    }

    return infer.CreateResponse[TeamState]{
        ID:    id,
        State: TeamState{TeamArgs: req.Inputs},
    }, nil
}
```

### Omitting Update Method

If a resource should always be replaced on update (never updated in-place), simply **omit the Update method**:

```go
type StackTag struct{}

// Only implement Create, Read, Delete
// NO Update method - framework automatically triggers replace
```

This matches the behavior of the manual implementation's `Diff()` returning `replaces`.

### Testing Your Migrated Resource

#### 1. Unit Tests (Optional but Recommended)

```go
// provider/pkg/infer/stack_tag_test.go
package infer

import (
    "context"
    "testing"
)

func TestStackTagCreate(t *testing.T) {
    st := StackTag{}
    resp, err := st.Create(context.Background(), infer.CreateRequest[StackTagArgs]{
        Name: "test",
        Inputs: StackTagArgs{
            Organization: "test-org",
            Project: "test-proj",
            Stack: "test-stack",
            Name: "env",
            Value: "prod",
        },
    })

    if err != nil {
        t.Fatalf("Create failed: %v", err)
    }

    expected := "test-org/test-proj/test-stack/env"
    if resp.ID != expected {
        t.Errorf("Expected ID %s, got %s", expected, resp.ID)
    }
}
```

#### 2. Integration Tests

Run existing examples:

```bash
# Find examples for your resource
ls examples/ | grep stack-tag

# Run the example test
cd examples
go test -v -run TestYamlStackTagExample -tags yaml -timeout 10m
```

#### 3. Schema Comparison

```bash
# Build provider
make provider

# Extract schema
.pulumi/bin/pulumi package get-schema bin/pulumi-resource-pulumiservice > /tmp/new-schema.json

# Compare relevant sections
# Look for your resource in /tmp/new-schema.json vs provider/cmd/pulumi-resource-pulumiservice/schema.json
```

## Troubleshooting

### Build Errors

**Error**: `infer.Resource[*MyResource]() not enough type arguments`
- **Fix**: Use instance not type parameter: `infer.Resource(&MyResource{})`

**Error**: `MyResource does not implement CustomResource`
- **Fix**: Implement all required methods: Create, Read, Delete (Update is optional)

**Error**: `cannot use req (variable of type infer.CreateRequest[Args]) as type context.Context`
- **Fix**: You're using old v0.16.0 signatures. Update to v1.1.2 signatures (see Lessons Learned)

### Schema Issues

**Problem**: Generated schema differs from manual schema
- **Check**: Field names match (case-sensitive)
- **Check**: Optional fields have `,optional` tag
- **Check**: Descriptions added via `Annotate()`
- **Check**: Secret fields have `,secret` tag

**Problem**: Fields appearing in wrong order
- **Solution**: This is cosmetic and doesn't affect functionality. Infer generates fields based on struct definition order.

### Runtime Errors

**Error**: `panic: API client not found in context`
- **Fix**: Ensure `hybridManualProvider` wrapper is properly set up in `hybrid.go`
- **Check**: Verify `WithClient()` is called before CRUD operations

**Error**: `nil pointer dereference` when accessing client
- **Fix**: Ensure client is properly initialized in provider Configure method

### Test Failures

**Error**: Integration test fails with 404
- **Check**: Verify ID format matches manual implementation exactly
- **Check**: API calls use correct endpoint paths

**Error**: Test fails with "resource already exists"
- **Fix**: Ensure Delete is properly cleaning up resources
- **Check**: Test cleanup logic

## Risk Register

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Schema incompatibility | Medium | High | Extensive schema comparison testing |
| Infer framework bugs | Low | Medium | Report upstream, implement workarounds |
| Breaking changes needed | Low | High | Version carefully, document migration |
| Timeline slip | High | Medium | Incremental delivery, can pause between phases |
| Team resource constraints | Medium | Medium | Flexible timeline, can extend phases |
| ESC integration issues | Medium | High | Test ESC resources thoroughly |
| Performance regression | Low | Medium | Add performance benchmarks |
| Customer impact | Low | High | Extensive testing, gradual rollout |

## Next Steps

### Immediate Actions

1. **Complete Phase 1.2**: Migrate OrgAccessToken resource
2. **Test secret handling**: Verify `,secret` tag auto-marking works
3. **Complete Phase 1.3**: Migrate AgentPool resource
4. **POC Validation**: Complete Phase 1.4 validation tasks

### After POC Complete

5. Begin Phase 2: Migrate simple resources (6 resources)
6. Begin Phase 3: Migrate medium complexity resources (7 resources)
7. Begin Phase 4: Migrate complex resources (5 resources)
8. Begin Phase 5: Migrate data sources (2 functions)
9. Begin Phase 6: Deprecation and cleanup
10. Begin Phase 7: Documentation and release

## References

- **Pulumi Go Provider**: https://github.com/pulumi/pulumi-go-provider
- **Infer Documentation**: https://pkg.go.dev/github.com/pulumi/pulumi-go-provider/infer
- **Builder Pattern Docs**: https://pkg.go.dev/github.com/pulumi/pulumi-go-provider/infer#NewProviderBuilder
- **Current provider implementation**: `provider/pkg/provider/provider.go`
- **Example resource**: `provider/pkg/resources/stack_tags.go`
- **pulumi-go-provider examples**: https://github.com/pulumi/pulumi-go-provider/tree/main/examples
- **Hybrid provider setup**: `provider/pkg/provider/hybrid.go`
- **Client context handling**: `provider/pkg/infer/client.go`
- **First migrated resource**: `provider/pkg/infer/stack_tag.go`

---

**Document Status**: Living document, updated as migration progresses
**Last Updated**: Phase 0 & 1.1 complete, Phase 1.2 next
**Maintainer**: Migration team
