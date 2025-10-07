# Permissions Query Example

This example demonstrates how to query organization permissions using Pulumi Service Provider data sources. It showcases two core invoke functions that allow you to discover and audit access to your Pulumi Cloud organization.

## Features Demonstrated

### 1. List All Teams (`getTeams`)
Query all teams in an organization, including:
- Team metadata (kind, name, displayName, description)
- Team members with their roles
- Stack permissions assigned to each team
- Environment permissions assigned to each team

### 2. List Accessible Stacks (`getStacks`)
Query all stacks accessible by the authenticated user, including:
- Stack identifiers (orgName, projectName, stackName)
- Last update timestamp
- Resource count

## Additional Available Functions

The provider also includes two additional data sources for more specific queries:

- `getTeamsForUser`: Find which teams a specific user belongs to
- `getStackPermissions`: Get detailed team and user permissions for a specific stack

See the provider documentation for usage examples of these functions.

## Usage

1. Set your Pulumi access token:
   ```bash
   export PULUMI_ACCESS_TOKEN=pul-xxx...
   ```

2. Configure the organization name:
   ```bash
   pulumi config set organizationName your-org-name
   ```

3. Run the example:
   ```bash
   pulumi up
   ```

## Outputs

The example exports two outputs:

- `organizationTeams`: Complete list of all teams with members and permissions
- `accessibleStacks`: All stacks accessible by the authenticated user

## Converting to Other Languages

This example is written in YAML for simplicity. To convert it to another programming language (TypeScript, Python, Go, C#, or Java), use the [`pulumi convert`](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/) command:

```bash
# Convert to TypeScript
pulumi convert --language typescript --out ../ts-permissions-query

# Convert to Python
pulumi convert --language python --out ../py-permissions-query

# Convert to Go
pulumi convert --language go --out ../go-permissions-query
```

## Use Cases

### Audit Organization Access
Use `getTeams` to get a complete view of all teams, their members, and what resources they can access.

### User Access Review
Use `getTeamsForUser` to quickly see which teams a user belongs to, helping with access audits and compliance.

### Stack Discovery
Use `getStacks` to discover all stacks you have access to across your organization.

### Permission Analysis
Use `getStackPermissions` to see exactly who has access to a specific stack and at what permission level.

## Permission Levels

Stack permissions are represented as integers:
- `0`: None
- `101`: Read
- `102`: Write
- `103`: Admin
- `104`: Creator (converted to Admin)

## Notes

- All invoke functions respect the authenticated user's permissions
- The `getStacks` function returns only stacks the user has access to
- The `getTeamsForUser` function matches against both username and GitHub login
- Empty configuration values will skip optional queries (using `ignoreChanges`)
