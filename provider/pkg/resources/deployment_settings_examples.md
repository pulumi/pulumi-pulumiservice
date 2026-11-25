{{% examples %}}
## Example Usage

{{% example %}}
### Deploy on every commit

This runs `pulumi up` on each push to `main` that changes a file under `infra/`, and `pulumi preview` on pull requests. The organization must have the Pulumi GitHub app installed on the repository.

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const settings = new pulumiservice.DeploymentSettings("settings", {
    organization: "my-org",
    project: "my-project",
    stack: "dev",
    sourceContext: {
        git: {
            branch: "refs/heads/main",
        },
    },
    github: {
        repository: "my-org/my-repo",
        deployCommits: true,
        previewPullRequests: true,
        paths: ["infra/**"],
    },
});
```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

settings = pulumiservice.DeploymentSettings("settings",
    organization="my-org",
    project="my-project",
    stack="dev",
    source_context={
        "git": {
            "branch": "refs/heads/main",
        },
    },
    github={
        "repository": "my-org/my-repo",
        "deploy_commits": True,
        "preview_pull_requests": True,
        "paths": ["infra/**"],
    })
```

```go
package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		_, err := pulumiservice.NewDeploymentSettings(ctx, "settings", &pulumiservice.DeploymentSettingsArgs{
			Organization: pulumi.String("my-org"),
			Project:      pulumi.String("my-project"),
			Stack:        pulumi.String("dev"),
			SourceContext: &pulumiservice.DeploymentSettingsSourceContextArgs{
				Git: &pulumiservice.DeploymentSettingsGitSourceArgs{
					Branch: pulumi.String("refs/heads/main"),
				},
			},
			Github: &pulumiservice.DeploymentSettingsGithubArgs{
				Repository:          pulumi.String("my-org/my-repo"),
				DeployCommits:       pulumi.Bool(true),
				PreviewPullRequests: pulumi.Bool(true),
				Paths: pulumi.StringArray{
					pulumi.String("infra/**"),
				},
			},
		})
		if err != nil {
			return err
		}
		return nil
	})
}
```

```csharp
using System.Collections.Generic;
using System.Linq;
using Pulumi;
using PulumiService = Pulumi.PulumiService;

return await Deployment.RunAsync(() => 
{
    var settings = new PulumiService.DeploymentSettings("settings", new()
    {
        Organization = "my-org",
        Project = "my-project",
        Stack = "dev",
        SourceContext = new PulumiService.Inputs.DeploymentSettingsSourceContextArgs
        {
            Git = new PulumiService.Inputs.DeploymentSettingsGitSourceArgs
            {
                Branch = "refs/heads/main",
            },
        },
        Github = new PulumiService.Inputs.DeploymentSettingsGithubArgs
        {
            Repository = "my-org/my-repo",
            DeployCommits = true,
            PreviewPullRequests = true,
            Paths = new[]
            {
                "infra/**",
            },
        },
    });

});
```

```java
package generated_program;

import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.DeploymentSettings;
import com.pulumi.pulumiservice.DeploymentSettingsArgs;
import com.pulumi.pulumiservice.inputs.DeploymentSettingsSourceContextArgs;
import com.pulumi.pulumiservice.inputs.DeploymentSettingsGitSourceArgs;
import com.pulumi.pulumiservice.inputs.DeploymentSettingsGithubArgs;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Map;
import java.io.File;
import java.nio.file.Files;
import java.nio.file.Paths;

public class App {
    public static void main(String[] args) {
        Pulumi.run(App::stack);
    }

    public static void stack(Context ctx) {
        var settings = new DeploymentSettings("settings", DeploymentSettingsArgs.builder()
            .organization("my-org")
            .project("my-project")
            .stack("dev")
            .sourceContext(DeploymentSettingsSourceContextArgs.builder()
                .git(DeploymentSettingsGitSourceArgs.builder()
                    .branch("refs/heads/main")
                    .build())
                .build())
            .github(DeploymentSettingsGithubArgs.builder()
                .repository("my-org/my-repo")
                .deployCommits(true)
                .previewPullRequests(true)
                .paths("infra/**")
                .build())
            .build());

    }
}
```

```yaml
resources:
  settings:
    type: pulumiservice:DeploymentSettings
    properties:
      organization: my-org
      project: my-project
      stack: dev
      sourceContext:
        git:
          branch: refs/heads/main
      github:
        repository: my-org/my-repo
        deployCommits: true
        previewPullRequests: true
        paths:
          - infra/**
```
{{% /example %}}

{{% example %}}
### Deploy when a git tag is pushed

This runs `pulumi up` when a tag matching `v*` is pushed, except release candidates such as `v1.2.0-rc1`. Pulumi Cloud rejects `deployCommits` and `deployTags` together, and `deployCommits` defaults to `true`, so set `deployCommits` to `false` when you enable `deployTags`. The organization must have the Pulumi GitHub app installed on the repository.

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const settings = new pulumiservice.DeploymentSettings("settings", {
    organization: "my-org",
    project: "my-project",
    stack: "production",
    sourceContext: {
        git: {
            branch: "refs/heads/main",
        },
    },
    github: {
        repository: "my-org/my-repo",
        deployCommits: false,
        deployTags: true,
        tagFilters: [
            "v*",
            "!*-rc*",
        ],
    },
});
```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

settings = pulumiservice.DeploymentSettings("settings",
    organization="my-org",
    project="my-project",
    stack="production",
    source_context={
        "git": {
            "branch": "refs/heads/main",
        },
    },
    github={
        "repository": "my-org/my-repo",
        "deploy_commits": False,
        "deploy_tags": True,
        "tag_filters": [
            "v*",
            "!*-rc*",
        ],
    })
```

```go
package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		_, err := pulumiservice.NewDeploymentSettings(ctx, "settings", &pulumiservice.DeploymentSettingsArgs{
			Organization: pulumi.String("my-org"),
			Project:      pulumi.String("my-project"),
			Stack:        pulumi.String("production"),
			SourceContext: &pulumiservice.DeploymentSettingsSourceContextArgs{
				Git: &pulumiservice.DeploymentSettingsGitSourceArgs{
					Branch: pulumi.String("refs/heads/main"),
				},
			},
			Github: &pulumiservice.DeploymentSettingsGithubArgs{
				Repository:    pulumi.String("my-org/my-repo"),
				DeployCommits: pulumi.Bool(false),
				DeployTags:    pulumi.Bool(true),
				TagFilters: pulumi.StringArray{
					pulumi.String("v*"),
					pulumi.String("!*-rc*"),
				},
			},
		})
		if err != nil {
			return err
		}
		return nil
	})
}
```

```csharp
using System.Collections.Generic;
using System.Linq;
using Pulumi;
using PulumiService = Pulumi.PulumiService;

return await Deployment.RunAsync(() => 
{
    var settings = new PulumiService.DeploymentSettings("settings", new()
    {
        Organization = "my-org",
        Project = "my-project",
        Stack = "production",
        SourceContext = new PulumiService.Inputs.DeploymentSettingsSourceContextArgs
        {
            Git = new PulumiService.Inputs.DeploymentSettingsGitSourceArgs
            {
                Branch = "refs/heads/main",
            },
        },
        Github = new PulumiService.Inputs.DeploymentSettingsGithubArgs
        {
            Repository = "my-org/my-repo",
            DeployCommits = false,
            DeployTags = true,
            TagFilters = new[]
            {
                "v*",
                "!*-rc*",
            },
        },
    });

});
```

```java
package generated_program;

import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.DeploymentSettings;
import com.pulumi.pulumiservice.DeploymentSettingsArgs;
import com.pulumi.pulumiservice.inputs.DeploymentSettingsSourceContextArgs;
import com.pulumi.pulumiservice.inputs.DeploymentSettingsGitSourceArgs;
import com.pulumi.pulumiservice.inputs.DeploymentSettingsGithubArgs;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Map;
import java.io.File;
import java.nio.file.Files;
import java.nio.file.Paths;

public class App {
    public static void main(String[] args) {
        Pulumi.run(App::stack);
    }

    public static void stack(Context ctx) {
        var settings = new DeploymentSettings("settings", DeploymentSettingsArgs.builder()
            .organization("my-org")
            .project("my-project")
            .stack("production")
            .sourceContext(DeploymentSettingsSourceContextArgs.builder()
                .git(DeploymentSettingsGitSourceArgs.builder()
                    .branch("refs/heads/main")
                    .build())
                .build())
            .github(DeploymentSettingsGithubArgs.builder()
                .repository("my-org/my-repo")
                .deployCommits(false)
                .deployTags(true)
                .tagFilters(                
                    "v*",
                    "!*-rc*")
                .build())
            .build());

    }
}
```

```yaml
resources:
  settings:
    type: pulumiservice:DeploymentSettings
    properties:
      organization: my-org
      project: my-project
      stack: production
      sourceContext:
        git:
          branch: refs/heads/main
      github:
        repository: my-org/my-repo
        # Pulumi Cloud rejects deployCommits and deployTags together, and
        # deployCommits defaults to true.
        deployCommits: false
        deployTags: true
        tagFilters:
          - v*
          - "!*-rc*"
```
{{% /example %}}

{{% example %}}
### Run deployments as a service principal

This makes the stack's deployments run as a service principal, with the roles and entity access it holds. Pulumi Cloud requires the service principal to have write access to the stack, so the settings depend on that grant.

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const deployer = new pulumiservice.ServicePrincipal("deployer", {
    organizationName: "my-org",
    name: "deployer",
});
const deployerStackAccess = new pulumiservice.OrganizationMemberEntityAccess("deployerStackAccess", {
    organizationName: "my-org",
    username: deployer.login,
    entityType: "stack",
    entityName: "my-project/production",
    permission: "stack-write",
});
const settings = new pulumiservice.DeploymentSettings("settings", {
    organization: "my-org",
    project: "my-project",
    stack: "production",
    sourceContext: {
        git: {
            branch: "refs/heads/main",
        },
    },
    operationContext: {
        servicePrincipal: {
            login: deployer.login,
        },
    },
}, {
    dependsOn: [deployerStackAccess],
});
```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

deployer = pulumiservice.ServicePrincipal("deployer",
    organization_name="my-org",
    name="deployer")
deployer_stack_access = pulumiservice.OrganizationMemberEntityAccess("deployerStackAccess",
    organization_name="my-org",
    username=deployer.login,
    entity_type="stack",
    entity_name="my-project/production",
    permission="stack-write")
settings = pulumiservice.DeploymentSettings("settings",
    organization="my-org",
    project="my-project",
    stack="production",
    source_context={
        "git": {
            "branch": "refs/heads/main",
        },
    },
    operation_context={
        "service_principal": {
            "login": deployer.login,
        },
    },
    opts = pulumi.ResourceOptions(depends_on=[deployer_stack_access]))
```

```go
package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		deployer, err := pulumiservice.NewServicePrincipal(ctx, "deployer", &pulumiservice.ServicePrincipalArgs{
			OrganizationName: pulumi.String("my-org"),
			Name:             pulumi.String("deployer"),
		})
		if err != nil {
			return err
		}
		deployerStackAccess, err := pulumiservice.NewOrganizationMemberEntityAccess(ctx, "deployerStackAccess", &pulumiservice.OrganizationMemberEntityAccessArgs{
			OrganizationName: pulumi.String("my-org"),
			Username:         deployer.Login,
			EntityType:       pulumi.String("stack"),
			EntityName:       pulumi.String("my-project/production"),
			Permission:       pulumi.String("stack-write"),
		})
		if err != nil {
			return err
		}
		_, err = pulumiservice.NewDeploymentSettings(ctx, "settings", &pulumiservice.DeploymentSettingsArgs{
			Organization: pulumi.String("my-org"),
			Project:      pulumi.String("my-project"),
			Stack:        pulumi.String("production"),
			SourceContext: &pulumiservice.DeploymentSettingsSourceContextArgs{
				Git: &pulumiservice.DeploymentSettingsGitSourceArgs{
					Branch: pulumi.String("refs/heads/main"),
				},
			},
			OperationContext: &pulumiservice.DeploymentSettingsOperationContextArgs{
				ServicePrincipal: &pulumiservice.OperationContextServicePrincipalArgs{
					Login: deployer.Login,
				},
			},
		}, pulumi.DependsOn([]pulumi.Resource{
			deployerStackAccess,
		}))
		if err != nil {
			return err
		}
		return nil
	})
}
```

```csharp
using System.Collections.Generic;
using System.Linq;
using Pulumi;
using PulumiService = Pulumi.PulumiService;

return await Deployment.RunAsync(() => 
{
    var deployer = new PulumiService.ServicePrincipal("deployer", new()
    {
        OrganizationName = "my-org",
        Name = "deployer",
    });

    var deployerStackAccess = new PulumiService.OrganizationMemberEntityAccess("deployerStackAccess", new()
    {
        OrganizationName = "my-org",
        Username = deployer.Login,
        EntityType = "stack",
        EntityName = "my-project/production",
        Permission = "stack-write",
    });

    var settings = new PulumiService.DeploymentSettings("settings", new()
    {
        Organization = "my-org",
        Project = "my-project",
        Stack = "production",
        SourceContext = new PulumiService.Inputs.DeploymentSettingsSourceContextArgs
        {
            Git = new PulumiService.Inputs.DeploymentSettingsGitSourceArgs
            {
                Branch = "refs/heads/main",
            },
        },
        OperationContext = new PulumiService.Inputs.DeploymentSettingsOperationContextArgs
        {
            ServicePrincipal = new PulumiService.Inputs.OperationContextServicePrincipalArgs
            {
                Login = deployer.Login,
            },
        },
    }, new CustomResourceOptions
    {
        DependsOn =
        {
            deployerStackAccess,
        },
    });

});
```

```java
package generated_program;

import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.ServicePrincipal;
import com.pulumi.pulumiservice.ServicePrincipalArgs;
import com.pulumi.pulumiservice.OrganizationMemberEntityAccess;
import com.pulumi.pulumiservice.OrganizationMemberEntityAccessArgs;
import com.pulumi.pulumiservice.DeploymentSettings;
import com.pulumi.pulumiservice.DeploymentSettingsArgs;
import com.pulumi.pulumiservice.inputs.DeploymentSettingsSourceContextArgs;
import com.pulumi.pulumiservice.inputs.DeploymentSettingsGitSourceArgs;
import com.pulumi.pulumiservice.inputs.DeploymentSettingsOperationContextArgs;
import com.pulumi.pulumiservice.inputs.OperationContextServicePrincipalArgs;
import com.pulumi.resources.CustomResourceOptions;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Map;
import java.io.File;
import java.nio.file.Files;
import java.nio.file.Paths;

public class App {
    public static void main(String[] args) {
        Pulumi.run(App::stack);
    }

    public static void stack(Context ctx) {
        var deployer = new ServicePrincipal("deployer", ServicePrincipalArgs.builder()
            .organizationName("my-org")
            .name("deployer")
            .build());

        var deployerStackAccess = new OrganizationMemberEntityAccess("deployerStackAccess", OrganizationMemberEntityAccessArgs.builder()
            .organizationName("my-org")
            .username(deployer.login())
            .entityType("stack")
            .entityName("my-project/production")
            .permission("stack-write")
            .build());

        var settings = new DeploymentSettings("settings", DeploymentSettingsArgs.builder()
            .organization("my-org")
            .project("my-project")
            .stack("production")
            .sourceContext(DeploymentSettingsSourceContextArgs.builder()
                .git(DeploymentSettingsGitSourceArgs.builder()
                    .branch("refs/heads/main")
                    .build())
                .build())
            .operationContext(DeploymentSettingsOperationContextArgs.builder()
                .servicePrincipal(OperationContextServicePrincipalArgs.builder()
                    .login(deployer.login())
                    .build())
                .build())
            .build(), CustomResourceOptions.builder()
                .dependsOn(deployerStackAccess)
                .build());

    }
}
```

```yaml
resources:
  deployer:
    type: pulumiservice:ServicePrincipal
    properties:
      organizationName: my-org
      name: deployer
  deployerStackAccess:
    type: pulumiservice:OrganizationMemberEntityAccess
    properties:
      organizationName: my-org
      username: ${deployer.login}
      entityType: stack
      entityName: my-project/production
      permission: stack-write
  settings:
    type: pulumiservice:DeploymentSettings
    properties:
      organization: my-org
      project: my-project
      stack: production
      sourceContext:
        git:
          branch: refs/heads/main
      operationContext:
        servicePrincipal:
          login: ${deployer.login}
    options:
      dependsOn:
        - ${deployerStackAccess}
```
{{% /example %}}
{{% /examples %}}
