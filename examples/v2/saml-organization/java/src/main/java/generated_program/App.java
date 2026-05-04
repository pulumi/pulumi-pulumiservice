package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.SAMLOrganization;
import com.pulumi.pulumiservice.v2.SAMLOrganizationArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");

            String idpDescriptor =
                "<md:EntityDescriptor xmlns:md=\"urn:oasis:names:tc:SAML:2.0:metadata\"\n" +
                "                     entityID=\"https://example-idp.invalid/metadata\">\n" +
                "  <md:IDPSSODescriptor protocolSupportEnumeration=\"urn:oasis:names:tc:SAML:2.0:protocol\"/>\n" +
                "</md:EntityDescriptor>";

            new SAMLOrganization("saml",
                SAMLOrganizationArgs.builder()
                    .orgName(serviceOrg)
                    .newIdpSsoDescriptor(idpDescriptor)
                    .build());
        });
    }
}
