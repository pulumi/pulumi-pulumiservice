using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";

    var idpDescriptor = @"<md:EntityDescriptor xmlns:md=""urn:oasis:names:tc:SAML:2.0:metadata""
                     entityID=""https://example-idp.invalid/metadata"">
  <md:IDPSSODescriptor protocolSupportEnumeration=""urn:oasis:names:tc:SAML:2.0:protocol""/>
</md:EntityDescriptor>";

    new Ps.V2.Auth.SAML("saml", new()
    {
        OrgName = organizationName,
        NewIdpSsoDescriptor = idpDescriptor,
    });
});
