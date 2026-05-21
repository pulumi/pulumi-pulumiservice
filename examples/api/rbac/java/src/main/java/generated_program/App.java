package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api.Role;
import com.pulumi.pulumiservice.api.RoleArgs;
import com.pulumi.pulumiservice.api_teams.Team;
import com.pulumi.pulumiservice.api_teams.TeamArgs;

import java.util.List;
import java.util.Map;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var nameSuffix = config.get("nameSuffix").orElse("manual");
            var roleDescription = config.get("roleDescription").orElse("Read-only access to stacks, created by the api rbac example.");

            var readOnlyRole = new Role("readOnlyRole",
                RoleArgs.builder()
                    .orgName(organizationName)
                    .name("api-rbac-read-only-" + nameSuffix)
                    .description(roleDescription)
                    .uxPurpose("role")
                    .details(Map.of(
                        "__type", "PermissionDescriptorAllow",
                        "permissions", List.of("stack:read")))
                    .build());

            var rbacTeam = new Team("rbacTeam",
                TeamArgs.builder()
                    .orgName(organizationName)
                    .name("api-rbac-team-" + nameSuffix)
                    .displayName("api RBAC Team " + nameSuffix)
                    .description("Team scaffold used by the api rbac example.")
                    .build());

            ctx.export("roleName", readOnlyRole.name());
            ctx.export("teamName", rbacTeam.name());
        });
    }
}
