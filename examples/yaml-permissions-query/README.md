# Permissions Query Example

This example demonstrates how to query organization permissions using Pulumi Service Provider data sources. It showcases five invoke functions that allow you to discover and audit access to your Pulumi Cloud organization.

## Features Demonstrated

### 1. List All Teams (`getTeams`)
Query all teams in an organization, including:
- Team metadata (kind, name, displayName, description)
- Team members with their roles
- Stack permissions assigned to each team
- Environment permissions assigned to each team

### 2. Find Teams for a User (`getTeamsForUser`)
Find which teams a specific user belongs to by searching their username or GitHub login:
- Returns simplified team information
- Useful for auditing user access

### 3. List Accessible Stacks (`getStacks`)
Query all stacks accessible by the authenticated user, including:
- Stack identifiers (orgName, projectName, stackName)
- Last update timestamp
- Resource count
- Optional pagination with `maxResults` parameter

### 4. Get Stack Team Permissions (`getStackTeamPermissions`)
Query which teams have access to a specific stack:
- Returns team permissions assigned at the stack level
- Permission levels: 0=None, 101=Read, 102=Write, 103=Admin
- Shows whether the requesting user is a member of each team

### 5. Get Stack Collaborators (`getStackCollaborators`)
Query individual users with direct access to a specific stack:
- Returns users with direct collaborator permissions (not team-derived)
- Permission levels: 0=None, 101=Read, 102=Write, 103=Admin
- Useful for auditing individual user grants

## Understanding Stack Permissions

The provider separates stack access into two distinct data sources:

- **`getStackTeamPermissions`**: Returns teams that have stack-level permissions configured. Users in these teams inherit stack access through their team membership.

- **`getStackCollaborators`**: Returns individual users who have been granted direct stack access (not through team membership). These are explicit user-level grants.

This 1-to-1 mapping with the Pulumi Cloud API provides clarity about the source of stack access and enables precise access auditing.

## Usage

### Basic Usage (Teams and Stacks Only)

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

### Testing User Lookup

To test `getTeamsForUser`, provide a username:

```bash
pulumi config set testUserName john-doe
pulumi up
```

### Testing Stack Permissions

To test `getStackTeamPermissions` and `getStackCollaborators`, provide stack details:

```bash
pulumi config set testStackOrg my-org
pulumi config set testStackProject my-project
pulumi config set testStackName dev
pulumi up
```

## Outputs

The example exports five outputs:

- `organizationTeams`: Complete list of all teams with members and permissions
- `accessibleStacks`: All stacks accessible by the authenticated user
- `userTeamMemberships`: Teams the specified user belongs to (if `testUserName` is set)
- `stackTeamPermissions`: Teams with access to the specified stack (if stack config is set)
- `stackCollaboratorPermissions`: Individual users with direct access to the specified stack (if stack config is set)

## Use Cases

1. **Organization Audit**: List all teams and their members to understand your org structure
2. **User Access Review**: Check which teams and resources a specific user can access
3. **Stack Discovery**: Find all stacks you have access to across the organization
4. **Permission Analysis**: Understand exactly who has access to sensitive stacks (both via teams and direct grants)
5. **Compliance Reporting**: Generate access reports for security and compliance requirements

## Converting to Other Languages

This example is written in YAML for simplicity. To convert it to another programming language (TypeScript, Python, Go, C#, or Java), use the [`pulumi convert`](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/) command:

```bash
# Convert to TypeScript
pulumi convert --language typescript --out ../ts-permissions-query

# Convert to Python
pulumi convert --language python --out ../py-permissions-query

# Convert to Go
pulumi convert --language go --out ../go-permissions-query

# Convert to C#
pulumi convert --language csharp --out ../cs-permissions-query

# Convert to Java
pulumi convert --language java --out ../java-permissions-query
```

## Permission Levels

All permission fields use integer values:
- `0`: None (no access)
- `101`: Read (can view but not modify)
- `102`: Write (can modify and deploy)
- `103`: Admin (full control including permission management)

## Notes

- The `getTeams` and `getStacks` functions always run and return results
- The `getTeamsForUser`, `getStackTeamPermissions`, and `getStackCollaborators` functions use `ignoreChanges` to prevent errors when optional config values are not provided
- All functions use the authenticated user's access token to determine visibility
