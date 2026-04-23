// Java variant of canonical/01-organization-bootstrap.
// Day-0 provisioning: three teams, a team-scoped CI token, a baseline
// policy group. Functionally equivalent to the sibling YAML program.

import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.orgs_teams.Team;
import com.pulumi.pulumiservice.orgs_teams.TeamArgs;
import com.pulumi.pulumiservice.orgs_tokens.TeamAccessToken;
import com.pulumi.pulumiservice.orgs_tokens.TeamAccessTokenArgs;
import com.pulumi.pulumiservice.orgs_policies.PolicyGroup;
import com.pulumi.pulumiservice.orgs_policies.PolicyGroupArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(App::stack);
    }

    private static void stack(Context ctx) {
        var cfg = ctx.config();
        var organizationName = cfg.get("organizationName").orElse("service-provider-test-org");
        var digits = cfg.get("digits").orElse("00000");

        new Team("admins", TeamArgs.builder()
            .organizationName(organizationName)
            .teamType("pulumi")
            .name("admins-" + digits)
            .displayName("Organization Admins")
            .description("Full org control; rotate this membership quarterly.")
            .build());

        var deployers = new Team("deployers", TeamArgs.builder()
            .organizationName(organizationName)
            .teamType("pulumi")
            .name("deployers-" + digits)
            .displayName("CI Deployers")
            .description("Automation-only team. Human members discouraged — use the CI token.")
            .build());

        new Team("readers", TeamArgs.builder()
            .organizationName(organizationName)
            .teamType("pulumi")
            .name("readers-" + digits)
            .displayName("Developers (read-only)")
            .description("Default team for new org members; grants stack read access.")
            .build());

        var ciToken = new TeamAccessToken("ciToken", TeamAccessTokenArgs.builder()
            .organizationName(organizationName)
            .teamName(deployers.name().applyValue(n -> n))
            .description("Used by GitHub Actions to deploy non-production stacks.")
            .build());

        new PolicyGroup("defaultGuardrails", PolicyGroupArgs.builder()
            .organizationName(organizationName)
            .name("baseline-guardrails-" + digits)
            .build());

        ctx.export("ciTokenValue", ciToken.value());
    }
}
