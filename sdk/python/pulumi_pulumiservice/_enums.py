# coding=utf-8
# *** WARNING: this file was generated by pulumi-language-python. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

from enum import Enum

__all__ = [
    'AuthPolicyDecision',
    'AuthPolicyPermissionLevel',
    'AuthPolicyTokenType',
    'EnvironmentPermission',
    'PulumiOperation',
    'TeamStackPermissionScope',
    'WebhookFilters',
    'WebhookFormat',
    'WebhookGroup',
]


class AuthPolicyDecision(str, Enum):
    DENY = "deny"
    """
    A deny rule for Oidc Issuer Policy.
    """
    ALLOW = "allow"
    """
    An allow rule for Oidc Issuer Policy.
    """


class AuthPolicyPermissionLevel(str, Enum):
    STANDARD = "standard"
    """
    Standard level of permissions.
    """
    ADMIN = "admin"
    """
    Admin level of permissions.
    """


class AuthPolicyTokenType(str, Enum):
    PERSONAL = "personal"
    """
    Personal Pulumi token. Requires userLogin field to be filled.
    """
    TEAM = "team"
    """
    Team Pulumi token. Requires teamName field to be filled.
    """
    ORGANIZATION = "organization"
    """
    Organization Pulumi token. Requires authorizedPermissions field to be filled.
    """
    RUNNER = "runner"
    """
    Deployment Runner Pulumi token. Requires runnerID field to be filled.
    """


class EnvironmentPermission(str, Enum):
    NONE = "none"
    """
    No permissions.
    """
    READ = "read"
    """
    Permission to read environment definition only.
    """
    OPEN = "open"
    """
    Permission to open and read the environment.
    """
    WRITE = "write"
    """
    Permission to open, read and update the environment.
    """
    ADMIN = "admin"
    """
    Permission for all operations on the environment.
    """


class PulumiOperation(str, Enum):
    UPDATE = "update"
    """
    Analogous to `pulumi up` command.
    """
    PREVIEW = "preview"
    """
    Analogous to `pulumi preview` command.
    """
    REFRESH = "refresh"
    """
    Analogous to `pulumi refresh` command.
    """
    DESTROY = "destroy"
    """
    Analogous to `pulumi destroy` command.
    """


class TeamStackPermissionScope(float, Enum):
    READ = 101
    """
    Grants read permissions to stack.
    """
    EDIT = 102
    """
    Grants edit permissions to stack.
    """
    ADMIN = 103
    """
    Grants admin permissions to stack.
    """


class WebhookFilters(str, Enum):
    STACK_CREATED = "stack_created"
    """
    Trigger a webhook when a stack is created. Only valid for org webhooks.
    """
    STACK_DELETED = "stack_deleted"
    """
    Trigger a webhook when a stack is deleted. Only valid for org webhooks.
    """
    UPDATE_SUCCEEDED = "update_succeeded"
    """
    Trigger a webhook when a stack update succeeds.
    """
    UPDATE_FAILED = "update_failed"
    """
    Trigger a webhook when a stack update fails.
    """
    PREVIEW_SUCCEEDED = "preview_succeeded"
    """
    Trigger a webhook when a stack preview succeeds.
    """
    PREVIEW_FAILED = "preview_failed"
    """
    Trigger a webhook when a stack preview fails.
    """
    DESTROY_SUCCEEDED = "destroy_succeeded"
    """
    Trigger a webhook when a stack destroy succeeds.
    """
    DESTROY_FAILED = "destroy_failed"
    """
    Trigger a webhook when a stack destroy fails.
    """
    REFRESH_SUCCEEDED = "refresh_succeeded"
    """
    Trigger a webhook when a stack refresh succeeds.
    """
    REFRESH_FAILED = "refresh_failed"
    """
    Trigger a webhook when a stack refresh fails.
    """
    DEPLOYMENT_QUEUED = "deployment_queued"
    """
    Trigger a webhook when a deployment is queued.
    """
    DEPLOYMENT_STARTED = "deployment_started"
    """
    Trigger a webhook when a deployment starts running.
    """
    DEPLOYMENT_SUCCEEDED = "deployment_succeeded"
    """
    Trigger a webhook when a deployment succeeds.
    """
    DEPLOYMENT_FAILED = "deployment_failed"
    """
    Trigger a webhook when a deployment fails.
    """
    DRIFT_DETECTED = "drift_detected"
    """
    Trigger a webhook when drift is detected.
    """
    DRIFT_DETECTION_SUCCEEDED = "drift_detection_succeeded"
    """
    Trigger a webhook when a drift detection run succeeds, regardless of whether drift is detected.
    """
    DRIFT_DETECTION_FAILED = "drift_detection_failed"
    """
    Trigger a webhook when a drift detection run fails.
    """
    DRIFT_REMEDIATION_SUCCEEDED = "drift_remediation_succeeded"
    """
    Trigger a webhook when a drift remediation run succeeds.
    """
    DRIFT_REMEDIATION_FAILED = "drift_remediation_failed"
    """
    Trigger a webhook when a drift remediation run fails.
    """
    ENVIRONMENT_CREATED = "environment_created"
    """
    Trigger a webhook when a new environment is created.
    """
    ENVIRONMENT_DELETED = "environment_deleted"
    """
    Trigger a webhook when an environment is deleted.
    """
    ENVIRONMENT_REVISION_CREATED = "environment_revision_created"
    """
    Trigger a webhook when a new revision is created on an environment.
    """
    ENVIRONMENT_REVISION_RETRACTED = "environment_revision_retracted"
    """
    Trigger a webhook when a revision is retracted on an environment.
    """
    ENVIRONMENT_REVISION_TAG_CREATED = "environment_revision_tag_created"
    """
    Trigger a webhook when a revision tag is created on an environment.
    """
    ENVIRONMENT_REVISION_TAG_DELETED = "environment_revision_tag_deleted"
    """
    Trigger a webhook when a revision tag is deleted on an environment.
    """
    ENVIRONMENT_REVISION_TAG_UPDATED = "environment_revision_tag_updated"
    """
    Trigger a webhook when a revision tag is updated on an environment.
    """
    ENVIRONMENT_TAG_CREATED = "environment_tag_created"
    """
    Trigger a webhook when an environment tag is created.
    """
    ENVIRONMENT_TAG_DELETED = "environment_tag_deleted"
    """
    Trigger a webhook when an environment tag is deleted.
    """
    ENVIRONMENT_TAG_UPDATED = "environment_tag_updated"
    """
    Trigger a webhook when an environment tag is updated.
    """
    IMPORTED_ENVIRONMENT_CHANGED = "imported_environment_changed"
    """
    Trigger a webhook when an imported environment has changed.
    """


class WebhookFormat(str, Enum):
    RAW = "raw"
    """
    The default webhook format.
    """
    SLACK = "slack"
    """
    Messages formatted for consumption by Slack incoming webhooks.
    """
    PULUMI_DEPLOYMENTS = "pulumi_deployments"
    """
    Initiate deployments on a stack from a Pulumi Cloud webhook.
    """
    MICROSOFT_TEAMS = "ms_teams"
    """
    Messages formatted for consumption by Microsoft Teams incoming webhooks.
    """


class WebhookGroup(str, Enum):
    STACKS = "stacks"
    """
    A group of webhooks containing all stack events.
    """
    DEPLOYMENTS = "deployments"
    """
    A group of webhooks containing all deployment events.
    """
    ENVIRONMENTS = "environments"
    """
    A group of webhooks containing all environment events.
    """
