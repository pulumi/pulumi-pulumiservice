package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api_teams.Team;
import com.pulumi.pulumiservice.api_teams.TeamArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var teamSuffix = config.get("teamSuffix").orElse("dev");
            var teamDescription = config.get("teamDescription").orElse("A team created by the api example.");

            var team = new Team("team",
                TeamArgs.builder()
                    .orgName(organizationName)
                    .name("api-team-" + teamSuffix)
                    .displayName("api Team " + teamSuffix)
                    .description(teamDescription)
                    .build());

            ctx.export("teamName", team.name());
        });
    }
}
