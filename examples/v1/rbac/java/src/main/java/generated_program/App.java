package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v1.Role;
import com.pulumi.pulumiservice.v1.RoleArgs;
import com.pulumi.pulumiservice.v1_teams.Team;
import com.pulumi.pulumiservice.v1_teams.TeamArgs;

import java.util.List;
import java.util.Map;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var nameSuffix = config.get("nameSuffix").orElse("manual");
            var roleDescription = config.get("roleDescription").orElse("Read-only access to stacks, created by the v1 rbac example.");

            var readOnlyRole = new Role("readOnlyRole",
                RoleArgs.builder()
                    .orgName(organizationName)
                    .name("v1-rbac-read-only-" + nameSuffix)
                    .description(roleDescription)
                    .uxPurpose("role")
                    .details(Map.of(
                        "__type", "PermissionDescriptorAllow",
                        "permissions", List.of("stack:read")))
                    .build());

            var rbacTeam = new Team("rbacTeam",
                TeamArgs.builder()
                    .orgName(organizationName)
                    .name("v1-rbac-team-" + nameSuffix)
                    .displayName("v1 RBAC Team " + nameSuffix)
                    .description("Team scaffold used by the v1 rbac example.")
                    .build());

            ctx.export("roleName", readOnlyRole.name());
            ctx.export("teamName", rbacTeam.name());
        });
    }
}
