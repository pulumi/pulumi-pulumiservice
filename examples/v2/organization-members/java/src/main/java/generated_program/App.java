package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.OrganizationMember;
import com.pulumi.pulumiservice.v2.OrganizationMemberArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var memberLogin = config.get("memberLogin").orElse("pulumi-bot");
            var memberRole = config.get("memberRole").orElse("member");

            var member = new OrganizationMember("member",
                OrganizationMemberArgs.builder()
                    .orgName(serviceOrg)
                    .userLogin(memberLogin)
                    .role(memberRole)
                    .build());

            ctx.export("memberId", member.id());
        });
    }
}
