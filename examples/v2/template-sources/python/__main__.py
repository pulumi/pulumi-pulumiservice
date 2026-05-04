import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
template_suffix = config.get("templateSuffix") or "dev"
source_url = config.get("sourceUrl") or "https://github.com/pulumi/examples"

source = ps_v2.OrgTemplateCollection(
    "source",
    org_name=service_org,
    name=f"v2-templates-{template_suffix}",
    source_url=source_url,
)

pulumi.export("collectionName", source.name)
