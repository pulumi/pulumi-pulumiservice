{{% examples %}}
## Example Usage

{{% example %}}
### Reference a saved Custom Agent by durable ID

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as pulumiservice from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organization0 = config.require("organization");
const scoutAgentId = config.require("scoutAgentId");
const scout = pulumiservice.getAgentOutput({
    organization: organization0,
    agentId: scoutAgentId,
});
export const agentId = scout.agentId;
export const displayName = scout.displayName;
```

```python
import pulumi
import pulumi_pulumiservice as pulumiservice

config = pulumi.Config()
organization0 = config.require("organization")
scout_agent_id = config.require("scoutAgentId")
scout = pulumiservice.get_agent_output(organization=organization0,
    agent_id=scout_agent_id)
pulumi.export("agentId", scout.agent_id)
pulumi.export("displayName", scout.display_name)
```

```go
package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		organization0 := cfg.Require("organization")
		scoutAgentId := cfg.Require("scoutAgentId")
		scout := pulumiservice.GetAgentOutput(ctx, pulumiservice.GetAgentOutputArgs{
			Organization: pulumi.String(organization0),
			AgentId:      pulumi.String(scoutAgentId),
		}, nil)
		ctx.Export("agentId", scout.AgentId())
		ctx.Export("displayName", scout.DisplayName())
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
    var config = new Config();
    var organization0 = config.Require("organization");
    var scoutAgentId = config.Require("scoutAgentId");
    var scout = PulumiService.GetAgent.Invoke(new()
    {
        Organization = organization0,
        AgentId = scoutAgentId,
    });

    return new Dictionary<string, object?>
    {
        ["agentId"] = scout.Apply(getAgentResult => getAgentResult.AgentId),
        ["displayName"] = scout.Apply(getAgentResult => getAgentResult.DisplayName),
    };
});
```

```java
package generated_program;

import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
import com.pulumi.pulumiservice.inputs.GetAgentArgs;
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
        final var config = ctx.config();
        final var organization0 = config.require("organization");
        final var scoutAgentId = config.require("scoutAgentId");
        final var scout = PulumiserviceFunctions.getAgent(GetAgentArgs.builder()
            .organization(organization0)
            .agentId(scoutAgentId)
            .build());

        ctx.export("agentId", scout.applyValue(_scout -> _scout.agentId()));
        ctx.export("displayName", scout.applyValue(_scout -> _scout.displayName()));
    }
}
```

```yaml
config:
  organization:
    type: string
  scoutAgentId:
    type: string
variables:
  scout:
    fn::invoke:
      function: pulumiservice:getAgent
      arguments:
        organization: ${organization}
        agentId: ${scoutAgentId}
outputs:
  agentId: ${scout.agentId}
  displayName: ${scout.displayName}
```
{{% /example %}}
{{% /examples %}}
