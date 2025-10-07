# Service Resource Example

This example demonstrates how to create and manage a Service resource in Pulumi Cloud.

## Overview

Services in Pulumi Cloud allow you to group and organize related stacks and environments. This example shows three different approaches for specifying the organization:

### 1. Automatic Default (Provider Infers)
The simplest approach - omit `organizationName` and the provider will automatically use the first organization from your Pulumi Cloud account where you are a member, admin, or billing manager.

```yaml
resources:
  myService:
    type: pulumiservice:index:Service
    properties:
      # organizationName omitted - uses default from your account
      ownerType: user
      ownerName: ${ownerName}
      name: my-service
```

### 2. Explicit API Call (`getCurrentUser`)
Use the `getCurrentUser()` function to explicitly retrieve and reference your default organization from the Pulumi Cloud API:

```yaml
variables:
  currentUser:
    fn::invoke:
      function: pulumiservice:index:getCurrentUser
      return: defaultOrganization

resources:
  myService:
    type: pulumiservice:index:Service
    properties:
      organizationName: ${currentUser}  # Explicitly reference
      ownerType: user
      ownerName: ${ownerName}
      name: my-service
```

### 3. Stack Context (`pulumi.getOrganization()`)
In TypeScript, Python, and other SDK languages, you can use the stack's deployment context:

**TypeScript:**
```typescript
const service = new pulumiservice.Service("myService", {
    organizationName: pulumi.getOrganization(),  // Uses org from stack context
    ownerType: "user",
    ownerName: ownerName,
    name: "my-service",
});
```

**Python:**
```python
service = pulumi_service.Service("my-service",
    organization_name=pulumi.get_organization(),  # Uses org from stack context
    owner_type="user",
    owner_name=owner_name,
    name="my-service"
)
```

## When to Use Each Approach

- **Automatic**: Best for quick prototyping or when you only have one organization
- **getCurrentUser()**: Best when you want to be explicit and reference your default org from multiple places
- **pulumi.getOrganization()**: Best when deploying to different orgs (the stack context determines the org)

## Difference Between API Default and Stack Context

- **getCurrentUser() / Automatic**: Returns the first org from your **Pulumi Cloud account** (e.g., `personal-org`)
- **pulumi.getOrganization()**: Returns the org from your **current stack deployment** (e.g., if deploying `work-org/infra/prod`, returns `work-org`)

If you're deploying `work-org/project/stack` but your default account org is `personal-org`:
- `getCurrentUser()` → `personal-org`
- `pulumi.getOrganization()` → `work-org`

## Running the Example

The example requires minimal configuration since most values are derived automatically:

```bash
# Only set the owner username (should be your Pulumi Cloud username)
pulumi config set ownerName <your-pulumi-username>

# The rest are handled automatically:
# - organizationName: Derived from getCurrentUser() or omitted for automatic default
# - projectName/stackName: Used only in the item name, can use pulumi.getProject()/getStack() in other languages

pulumi up
```

**Note**: In a real-world scenario using TypeScript or Python, you'd typically use:
```typescript
ownerName: pulumi.getOrganization()  // Gets current user/org from stack context
projectName: pulumi.getProject()     // Gets project from stack
stackName: pulumi.getStack()         // Gets stack name
```

So you wouldn't need to configure these at all - they come from your deployment context automatically.

## Converting to Other Languages

This YAML example can be converted to other Pulumi languages using:

```bash
pulumi convert --language typescript --out ts-service
pulumi convert --language python --out py-service
pulumi convert --language go --out go-service
```

See the [Pulumi convert documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/) for more details.
