package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api_tokens.OrgToken;
import com.pulumi.pulumiservice.api_tokens.OrgTokenArgs;
import com.pulumi.pulumiservice.api_tokens.PersonalToken;
import com.pulumi.pulumiservice.api_tokens.PersonalTokenArgs;
import com.pulumi.pulumiservice.api_teams.Team;
import com.pulumi.pulumiservice.api_teams.TeamArgs;
import com.pulumi.pulumiservice.api_tokens.TeamToken;
import com.pulumi.pulumiservice.api_tokens.TeamTokenArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var tokenSuffix = config.get("tokenSuffix").orElse("dev");
            var tokenDescription = config.get("tokenDescription").orElse("example api access token");

            var team = new Team("team",
                TeamArgs.builder()
                    .orgName(organizationName)
                    .name("api-tokens-team-" + tokenSuffix)
                    .displayName("api Tokens Team " + tokenSuffix)
                    .description("Owner team for the api access-tokens example")
                    .build());

            var orgToken = new OrgToken("orgToken",
                OrgTokenArgs.builder()
                    .orgName(organizationName)
                    .name("api-org-token-" + tokenSuffix)
                    .description(tokenDescription)
                    .admin(false)
                    .expires(0)
                    .build());

            var teamToken = new TeamToken("teamToken",
                TeamTokenArgs.builder()
                    .orgName(organizationName)
                    .teamName(team.name())
                    .name("api-team-token-" + tokenSuffix)
                    .description(tokenDescription)
                    .expires(0)
                    .build());

            new PersonalToken("personalToken",
                PersonalTokenArgs.builder()
                    .description(tokenDescription)
                    .expires(0)
                    .build());

            ctx.export("orgTokenId", orgToken.tokenId());
            ctx.export("teamTokenId", teamToken.tokenId());
        });
    }
}
