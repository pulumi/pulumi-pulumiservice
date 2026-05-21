import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
template_suffix = config.get("templateSuffix") or "dev"
source_url = config.get("sourceUrl") or "https://github.com/pulumi/examples"

source = ps_api.OrgTemplateCollection(
    "source",
    org_name=organization_name,
    name=f"api-templates-{template_suffix}",
    source_url=source_url,
)

pulumi.export("collectionName", source.name)
