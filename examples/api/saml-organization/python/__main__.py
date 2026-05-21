import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"

idp_descriptor = """<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata"
                     entityID="https://example-idp.invalid/metadata">
  <md:IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol"/>
</md:EntityDescriptor>"""

ps_api.auth.SAML(
    "saml",
    org_name=organization_name,
    new_idp_sso_descriptor=idp_descriptor,
)
