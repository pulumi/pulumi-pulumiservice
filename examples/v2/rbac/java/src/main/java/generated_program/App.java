package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.Role;
import com.pulumi.pulumiservice.v2.RoleArgs;
import com.pulumi.pulumiservice.v2.Team;
import com.pulumi.pulumiservice.v2.TeamArgs;

import java.util.List;
import java.util.Map;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var nameSuffix = config.get("nameSuffix").orElse("manual");
            var roleDescription = config.get("roleDescription").orElse("Read-only access to stacks, created by the v2 rbac example.");

            var readOnlyRole = new Role("readOnlyRole",
                RoleArgs.builder()
                    .orgName(serviceOrg)
                    .name("v2-rbac-read-only-" + nameSuffix)
                    .description(roleDescription)
                    .uxPurpose("role")
                    .details(Map.of(
                        "__type", "PermissionDescriptorAllow",
                        "permissions", List.of("stack:read")))
                    .build());

            var rbacTeam = new Team("rbacTeam",
                TeamArgs.builder()
                    .orgName(serviceOrg)
                    .name("v2-rbac-team-" + nameSuffix)
                    .displayName("v2 RBAC Team " + nameSuffix)
                    .description("Team scaffold used by the v2 rbac example.")
                    .build());

            ctx.export("roleName", readOnlyRole.name());
            ctx.export("teamName", rbacTeam.name());
        });
    }
}
