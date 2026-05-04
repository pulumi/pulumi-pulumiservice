package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.OrgToken;
import com.pulumi.pulumiservice.v2.OrgTokenArgs;
import com.pulumi.pulumiservice.v2.PersonalToken;
import com.pulumi.pulumiservice.v2.PersonalTokenArgs;
import com.pulumi.pulumiservice.v2.Team;
import com.pulumi.pulumiservice.v2.TeamArgs;
import com.pulumi.pulumiservice.v2.TeamToken;
import com.pulumi.pulumiservice.v2.TeamTokenArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var tokenSuffix = config.get("tokenSuffix").orElse("dev");
            var tokenDescription = config.get("tokenDescription").orElse("example v2 access token");

            var team = new Team("team",
                TeamArgs.builder()
                    .orgName(serviceOrg)
                    .name("v2-tokens-team-" + tokenSuffix)
                    .displayName("v2 Tokens Team " + tokenSuffix)
                    .description("Owner team for the v2 access-tokens example")
                    .build());

            var orgToken = new OrgToken("orgToken",
                OrgTokenArgs.builder()
                    .orgName(serviceOrg)
                    .name("v2-org-token-" + tokenSuffix)
                    .description(tokenDescription)
                    .admin(false)
                    .expires(0)
                    .build());

            var teamToken = new TeamToken("teamToken",
                TeamTokenArgs.builder()
                    .orgName(serviceOrg)
                    .teamName(team.name())
                    .name("v2-team-token-" + tokenSuffix)
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
