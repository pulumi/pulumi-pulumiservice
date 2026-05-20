package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v1_teams.Team;
import com.pulumi.pulumiservice.v1_teams.TeamArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var teamSuffix = config.get("teamSuffix").orElse("dev");
            var teamDescription = config.get("teamDescription").orElse("A team created by the v1 example.");

            var team = new Team("team",
                TeamArgs.builder()
                    .orgName(organizationName)
                    .name("v1-team-" + teamSuffix)
                    .displayName("v1 Team " + teamSuffix)
                    .description(teamDescription)
                    .build());

            ctx.export("teamName", team.name());
        });
    }
}
