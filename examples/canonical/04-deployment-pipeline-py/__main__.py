"""Python variant of canonical/04-deployment-pipeline.

A git-driven Pulumi Deployments pipeline: source + executor settings,
drift detection on a cron, TTL on ephemeral stacks. Behavioral twin of
the sibling YAML program.
"""

import pulumi
from pulumi_pulumiservice.stacks import deployments

cfg = pulumi.Config()
organization_name = cfg.get("organizationName") or "service-provider-test-org"
project = cfg.get("project") or "infrastructure"
stack = cfg.get("stack") or "production"

deployments.Settings(
    "deploymentSettings",
    organization=organization_name,
    project=project,
    stack=stack,
    source_context={
        "git": {
            "repo_url": "https://github.com/acme-corp/infrastructure.git",
            "branch": "refs/heads/main",
            "repo_dir": "stacks/production",
        },
    },
    github={
        "repository": "acme-corp/infrastructure",
        "deploy_commits": True,
        "preview_pull_requests": True,
        "pull_request_template": True,
    },
    executor_context={"executor_image": {"image": "pulumi/pulumi:latest"}},
    operation_context={
        "pre_run_commands": ["npm ci"],
        "environment_variables": {"NODE_ENV": "production"},
    },
)

deployments.DriftSchedule(
    "driftCheck",
    organization=organization_name,
    project=project,
    stack=stack,
    schedule_cron="0 */6 * * *",
    auto_remediate=False,
)

deployments.TtlSchedule(
    "ephemeralTtl",
    organization=organization_name,
    project=project,
    stack=stack,
    timestamp="2026-12-31T00:00:00Z",
    delete_after_destroy=True,
)
