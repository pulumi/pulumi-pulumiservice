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
{{% /examples %}}
