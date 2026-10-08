package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
import com.pulumi.pulumiservice.api.Role;
import com.pulumi.pulumiservice.api.RoleArgs;
import com.pulumi.pulumiservice.api_teams.Team;
import com.pulumi.pulumiservice.api_teams.TeamArgs;
import com.pulumi.pulumiservice.enums.RbacEntityType;
import com.pulumi.pulumiservice.inputs.BuildAllowPermissionsArgs;
import com.pulumi.pulumiservice.inputs.BuildComposePermissionsArgs;
import com.pulumi.pulumiservice.inputs.BuildGroupPermissionsArgs;
import com.pulumi.pulumiservice.inputs.BuildTagConditionalPermissionsArgs;

import java.util.List;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var nameSuffix = config.get("nameSuffix").orElse("manual");
            var roleDescription = config.get("roleDescription").orElse("Stack read everywhere, plus write on stacks tagged team=platform.");

            // Pulumi Cloud models a role in three layers: a set grants scopes, a policy
            // composes sets (optionally gated on a condition), and a role composes policies.
            var stackReadSet = new Role("stackReadSet",
                RoleArgs.builder()
                    .orgName(organizationName)
                    .name("api-rbac-stack-read-" + nameSuffix)
                    .uxPurpose("set")
                    .resourceType("stack")
                    .details(PulumiserviceFunctions.buildAllowPermissions(BuildAllowPermissionsArgs.builder()
                            .permissions("stack:read")
                            .build())
                        .applyValue(r -> (Object) r.permissions()))
                    .build());

            var stackWriteSet = new Role("stackWriteSet",
                RoleArgs.builder()
                    .orgName(organizationName)
                    .name("api-rbac-stack-write-" + nameSuffix)
                    .uxPurpose("set")
                    .resourceType("stack")
                    .details(PulumiserviceFunctions.buildAllowPermissions(BuildAllowPermissionsArgs.builder()
                            .permissions("stack:read", "stack:write")
                            .build())
                        .applyValue(r -> (Object) r.permissions()))
                    .build());

            var allStacksRead = PulumiserviceFunctions.buildComposePermissions(BuildComposePermissionsArgs.builder()
                    .permissionDescriptorIds(stackReadSet.roleID().applyValue(List::of))
                    .build())
                .applyValue(r -> r.permissions());

            var platformStacksWrite = PulumiserviceFunctions.buildTagConditionalPermissions(BuildTagConditionalPermissionsArgs.builder()
                    .entityType(RbacEntityType.Stack)
                    .tagKey("team")
                    .tagValue("platform")
                    .setIds(stackWriteSet.roleID().applyValue(List::of))
                    .build())
                .applyValue(r -> r.permissions());

            var platformPolicy = new Role("platformPolicy",
                RoleArgs.builder()
                    .orgName(organizationName)
                    .name("api-rbac-policy-" + nameSuffix)
                    .uxPurpose("policy")
                    .details(PulumiserviceFunctions.buildGroupPermissions(BuildGroupPermissionsArgs.builder()
                            .entries(Output.all(allStacksRead, platformStacksWrite))
                            .build())
                        .applyValue(r -> (Object) r.permissions()))
                    .build());

            var platformRole = new Role("platformRole",
                RoleArgs.builder()
                    .orgName(organizationName)
                    .name("api-rbac-role-" + nameSuffix)
                    .description(roleDescription)
                    .uxPurpose("role")
                    .details(PulumiserviceFunctions.buildComposePermissions(BuildComposePermissionsArgs.builder()
                            .permissionDescriptorIds(platformPolicy.roleID().applyValue(List::of))
                            .build())
                        .applyValue(r -> (Object) r.permissions()))
                    .build());

            var rbacTeam = new Team("rbacTeam",
                TeamArgs.builder()
                    .orgName(organizationName)
                    .name("api-rbac-team-" + nameSuffix)
                    .displayName("api RBAC Team " + nameSuffix)
                    .description("Team scaffold used by the api rbac example.")
                    .build());

            ctx.export("roleName", platformRole.name());
            ctx.export("teamName", rbacTeam.name());
        });
    }
}
