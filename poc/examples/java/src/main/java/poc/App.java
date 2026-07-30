package poc;

import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.pulumiservice.api.Role;
import com.pulumi.pulumiservice.api.RoleArgs;
import com.pulumi.pulumiservice.api.inputs.PermissionDescriptorAllowArgs;
import com.pulumi.pulumiservice.api.inputs.PermissionDescriptorGroupArgs;

import java.util.List;

public final class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var org = System.getenv().getOrDefault("POC_ORG", "poc-org");

            // The union slot is Object (6 members exceeds Either<L,R>), but the
            // typed variant builders exist; __type is a legal Java identifier.
            var allow = PermissionDescriptorAllowArgs.builder()
                    .__type("PermissionDescriptorAllow")
                    .permissions(List.of("organization:read_usage"))
                    .build();

            var group = PermissionDescriptorGroupArgs.builder()
                    .__type("PermissionDescriptorGroup")
                    .entries(List.of(allow))
                    .build();

            var role = new Role("poc-union-role", RoleArgs.builder()
                    .orgName(org)
                    .name("poc-union-role")
                    .uxPurpose("set")
                    .resourceType("global")
                    .description("POC: discriminated union permission tree")
                    .details(group)
                    .build());

            ctx.export("roleId", role.id());
            ctx.export("detailsOut", role.details());
        });
    }
}
