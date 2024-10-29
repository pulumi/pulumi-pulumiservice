### Improvements

- Make SourceContext optional in DeploymentSettings [#427](https://github.com/pulumi/pulumi-pulumiservice/pulls/427)

  In some advanced use cases, for example if your source code is baked into a custom image, or you are obtaining 
  the source code from a different source, you may not want to specify a `SourceContext` in your `DeploymentSettings`.

### Bug Fixes
- Fixing TeamEnvironmentPermission, project field was not working [#429](https://github.com/pulumi/pulumi-pulumiservice/issues/429)

### Miscellaneous
