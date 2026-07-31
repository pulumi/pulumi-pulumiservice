package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api.Role;
import com.pulumi.pulumiservice.api.RoleArgs;
import com.pulumi.pulumiservice.api.inputs.PermissionDescriptorAllowArgs;
import com.pulumi.pulumiservice.api.inputs.PermissionDescriptorConditionArgs;
import com.pulumi.pulumiservice.api.inputs.PermissionDescriptorGroupArgs;
import com.pulumi.pulumiservice.api.inputs.PermissionExpressionAndArgs;
import com.pulumi.pulumiservice.api.inputs.PermissionExpressionEnvironmentArgs;
import com.pulumi.pulumiservice.api.inputs.PermissionExpressionEqualArgs;
import com.pulumi.pulumiservice.api.inputs.PermissionExpressionNotArgs;
import com.pulumi.pulumiservice.api.inputs.PermissionLiteralExpressionEnvironmentArgs;

import java.util.List;

public class App {
    private static PermissionExpressionEqualArgs envMatch(String identity) {
        return PermissionExpressionEqualArgs.builder()
            .__type("PermissionExpressionEqual")
            .left(PermissionExpressionEnvironmentArgs.builder()
                .__type("PermissionExpressionEnvironment")
                .build())
            .right(PermissionLiteralExpressionEnvironmentArgs.builder()
                .__type("PermissionLiteralExpressionEnvironment")
                .identity(identity)
                .build())
            .build();
    }

    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var nameSuffix = config.get("nameSuffix").orElse("manual");
            var roleDescription = config.get("roleDescription")
                .orElse("Environment-scoped read access, created by the api rbac-scoped example.");
            var allowedEnvironmentId = config.get("allowedEnvironmentId")
                .orElse("c5549aa1-87db-4d67-a195-455b56772900");
            var deniedEnvironmentId = config.get("deniedEnvironmentId")
                .orElse("3cb9b7ad-0848-4e0d-aeff-8e9f093fd2d9");

            var scopedRole = new Role("scopedRole",
                RoleArgs.builder()
                    .orgName(organizationName)
                    .name("api-rbac-scoped-" + nameSuffix)
                    .description(roleDescription)
                    .uxPurpose("role")
                    .details(PermissionDescriptorGroupArgs.builder()
                        .__type("PermissionDescriptorGroup")
                        .entries(List.of(PermissionDescriptorConditionArgs.builder()
                            .__type("PermissionDescriptorCondition")
                            .condition(PermissionExpressionAndArgs.builder()
                                .__type("PermissionExpressionAnd")
                                .left(envMatch(allowedEnvironmentId))
                                .right(PermissionExpressionNotArgs.builder()
                                    .__type("PermissionExpressionNot")
                                    .node(envMatch(deniedEnvironmentId))
                                    .build())
                                .build())
                            .subNode(PermissionDescriptorAllowArgs.builder()
                                .__type("PermissionDescriptorAllow")
                                .permissions("environment:read", "environment:open")
                                .build())
                            .build()))
                        .build())
                    .build());

            ctx.export("roleName", scopedRole.name());
        });
    }
}
