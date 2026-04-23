// Java variant of canonical/07-tiered-team-access.
// Three teams × two stacks × three permission tiers. Behavioral twin of
// the sibling YAML program.

import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.orgs_teams.Team;
import com.pulumi.pulumiservice.orgs_teams.TeamArgs;
import com.pulumi.pulumiservice.stacks_permissions.TeamStackPermission;
import com.pulumi.pulumiservice.stacks_permissions.TeamStackPermissionArgs;

public class App {
    // Team-stack permission levels: 0=none, 101=read, 102=edit, 103=admin.
    private static final int PERM_READ = 101;
    private static final int PERM_EDIT = 102;
    private static final int PERM_ADMIN = 103;

    public static void main(String[] args) {
        Pulumi.run(App::stack);
    }

    private static void stack(Context ctx) {
        var cfg = ctx.config();
        var org = cfg.get("organizationName").orElse("service-provider-test-org");
        var digits = cfg.get("digits").orElse("00000");

        var platformAdmins = new Team("platformAdmins", TeamArgs.builder()
            .organizationName(org).teamType("pulumi")
            .name("platform-admins-" + digits)
            .displayName("Platform Admins")
            .description("Break-glass access to everything. Keep small.")
            .build());
        var billingOwners = new Team("billingOwners", TeamArgs.builder()
            .organizationName(org).teamType("pulumi")
            .name("billing-owners-" + digits)
            .displayName("Billing Service Owners")
            .description("Owns the billing service stacks end-to-end.")
            .build());
        var developers = new Team("developers", TeamArgs.builder()
            .organizationName(org).teamType("pulumi")
            .name("developers-" + digits)
            .displayName("Developers (all)")
            .description("Read everything; deploy nothing without an explicit grant.")
            .build());

        grant(org, "platformAdminPerm", "platform", platformAdmins.name().applyValue(n -> n), PERM_ADMIN);
        grant(org, "platformDevRead", "platform", developers.name().applyValue(n -> n), PERM_READ);
        grant(org, "billingAdminPerm", "billing-service", platformAdmins.name().applyValue(n -> n), PERM_ADMIN);
        grant(org, "billingOwnerPerm", "billing-service", billingOwners.name().applyValue(n -> n), PERM_EDIT);
        grant(org, "billingDevRead", "billing-service", developers.name().applyValue(n -> n), PERM_READ);
    }

    private static void grant(String org, String name, String project, Output<String> team, int perm) {
        new TeamStackPermission(name, TeamStackPermissionArgs.builder()
            .organization(org).project(project).stack("prod")
            .team(team).permission(perm)
            .build());
    }
}
