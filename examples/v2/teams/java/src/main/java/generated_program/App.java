package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2_teams.Team;
import com.pulumi.pulumiservice.v2_teams.TeamArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var teamSuffix = config.get("teamSuffix").orElse("dev");
            var teamDescription = config.get("teamDescription").orElse("A team created by the v2 example.");

            var team = new Team("team",
                TeamArgs.builder()
                    .orgName(serviceOrg)
                    .name("v2-team-" + teamSuffix)
                    .displayName("v2 Team " + teamSuffix)
                    .description(teamDescription)
                    .build());

            ctx.export("teamName", team.name());
        });
    }
}
