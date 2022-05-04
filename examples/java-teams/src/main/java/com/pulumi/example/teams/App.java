package com.pulumi.example.teams;
import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.Team;
import com.pulumi.pulumiservice.TeamArgs;

public class App {
   public static void main(String[] args) {
       Pulumi.run(ctx -> {
           var team = new Team("Team", TeamArgs
                   .builder()
                   .name("brand-new-java-team")
                   .description("This was created with Pulumi")
                   .displayName("PulumiUP Team")
                   .organizationName("service-provider-test-org")
                   .teamType("pulumi")
                   .members("pulumi-bot", "service-provider-example-user")
                   .build());
           ctx.export("members", team.members());
       });
   }
}