package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api.OrganizationMember;
import com.pulumi.pulumiservice.api.OrganizationMemberArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var memberLogin = config.get("memberLogin").orElse("pulumi-bot");
            var memberRole = config.get("memberRole").orElse("member");

            var member = new OrganizationMember("member",
                OrganizationMemberArgs.builder()
                    .orgName(organizationName)
                    .userLogin(memberLogin)
                    .role(memberRole)
                    .build());

            ctx.export("memberId", member.id());
        });
    }
}
