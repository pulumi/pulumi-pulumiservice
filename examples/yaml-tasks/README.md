# Neo Task Creation Example (YAML)

This example demonstrates how to create a Neo task using the Pulumi Service Provider with YAML.

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- A Pulumi Cloud account
- Access to an organization where you can create tasks

## Running the Example

1. Configure your organization name:

   ```bash
   pulumi config set organizationName your-org-name
   ```

2. Run `pulumi up` to create the task:

   ```bash
   pulumi up
   ```

3. The task ID will be exported as an output.

## Converting to Other Languages

This YAML example can be converted to other programming languages using the `pulumi convert` command. For more information, see the [pulumi convert documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).

Examples:

```bash
# Convert to TypeScript
pulumi convert --from yaml --to typescript --out ../ts-tasks

# Convert to Python
pulumi convert --from yaml --to python --out ../py-tasks

# Convert to Go
pulumi convert --from yaml --to go --out ../go-tasks
```

## About Neo Tasks

Tasks in the Neo agent system are immutable once created. They track user instructions and entity changes for AI-powered workflows. The `createTask` function creates a new task and returns its ID and metadata.
