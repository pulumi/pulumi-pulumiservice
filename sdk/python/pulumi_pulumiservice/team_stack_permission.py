# coding=utf-8
# *** WARNING: this file was generated by pulumi-language-python. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

import copy
import warnings
import sys
import pulumi
import pulumi.runtime
from typing import Any, Mapping, Optional, Sequence, Union, overload
if sys.version_info >= (3, 11):
    from typing import NotRequired, TypedDict, TypeAlias
else:
    from typing_extensions import NotRequired, TypedDict, TypeAlias
from . import _utilities
from ._enums import *

__all__ = ['TeamStackPermissionArgs', 'TeamStackPermission']

@pulumi.input_type
class TeamStackPermissionArgs:
    def __init__(__self__, *,
                 organization: pulumi.Input[str],
                 permission: 'TeamStackPermissionScope',
                 project: pulumi.Input[str],
                 stack: pulumi.Input[str],
                 team: pulumi.Input[str]):
        """
        The set of arguments for constructing a TeamStackPermission resource.
        :param pulumi.Input[str] organization: The organization or the personal account name of the stack.
        :param 'TeamStackPermissionScope' permission: Sets the permission level that this team will be granted to the stack.
        :param pulumi.Input[str] project: The project name for this stack.
        :param pulumi.Input[str] stack: The name of the stack that the team will be granted permissions to.
        :param pulumi.Input[str] team: The name of the team to grant this stack permissions to. This is not the display name.
        """
        pulumi.set(__self__, "organization", organization)
        pulumi.set(__self__, "permission", permission)
        pulumi.set(__self__, "project", project)
        pulumi.set(__self__, "stack", stack)
        pulumi.set(__self__, "team", team)

    @property
    @pulumi.getter
    def organization(self) -> pulumi.Input[str]:
        """
        The organization or the personal account name of the stack.
        """
        return pulumi.get(self, "organization")

    @organization.setter
    def organization(self, value: pulumi.Input[str]):
        pulumi.set(self, "organization", value)

    @property
    @pulumi.getter
    def permission(self) -> 'TeamStackPermissionScope':
        """
        Sets the permission level that this team will be granted to the stack.
        """
        return pulumi.get(self, "permission")

    @permission.setter
    def permission(self, value: 'TeamStackPermissionScope'):
        pulumi.set(self, "permission", value)

    @property
    @pulumi.getter
    def project(self) -> pulumi.Input[str]:
        """
        The project name for this stack.
        """
        return pulumi.get(self, "project")

    @project.setter
    def project(self, value: pulumi.Input[str]):
        pulumi.set(self, "project", value)

    @property
    @pulumi.getter
    def stack(self) -> pulumi.Input[str]:
        """
        The name of the stack that the team will be granted permissions to.
        """
        return pulumi.get(self, "stack")

    @stack.setter
    def stack(self, value: pulumi.Input[str]):
        pulumi.set(self, "stack", value)

    @property
    @pulumi.getter
    def team(self) -> pulumi.Input[str]:
        """
        The name of the team to grant this stack permissions to. This is not the display name.
        """
        return pulumi.get(self, "team")

    @team.setter
    def team(self, value: pulumi.Input[str]):
        pulumi.set(self, "team", value)


class TeamStackPermission(pulumi.CustomResource):
    @overload
    def __init__(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 organization: Optional[pulumi.Input[str]] = None,
                 permission: Optional['TeamStackPermissionScope'] = None,
                 project: Optional[pulumi.Input[str]] = None,
                 stack: Optional[pulumi.Input[str]] = None,
                 team: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        """
        Grants a team permissions to the specified stack.

        :param str resource_name: The name of the resource.
        :param pulumi.ResourceOptions opts: Options for the resource.
        :param pulumi.Input[str] organization: The organization or the personal account name of the stack.
        :param 'TeamStackPermissionScope' permission: Sets the permission level that this team will be granted to the stack.
        :param pulumi.Input[str] project: The project name for this stack.
        :param pulumi.Input[str] stack: The name of the stack that the team will be granted permissions to.
        :param pulumi.Input[str] team: The name of the team to grant this stack permissions to. This is not the display name.
        """
        ...
    @overload
    def __init__(__self__,
                 resource_name: str,
                 args: TeamStackPermissionArgs,
                 opts: Optional[pulumi.ResourceOptions] = None):
        """
        Grants a team permissions to the specified stack.

        :param str resource_name: The name of the resource.
        :param TeamStackPermissionArgs args: The arguments to use to populate this resource's properties.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        ...
    def __init__(__self__, resource_name: str, *args, **kwargs):
        resource_args, opts = _utilities.get_resource_args_opts(TeamStackPermissionArgs, pulumi.ResourceOptions, *args, **kwargs)
        if resource_args is not None:
            __self__._internal_init(resource_name, opts, **resource_args.__dict__)
        else:
            __self__._internal_init(resource_name, *args, **kwargs)

    def _internal_init(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 organization: Optional[pulumi.Input[str]] = None,
                 permission: Optional['TeamStackPermissionScope'] = None,
                 project: Optional[pulumi.Input[str]] = None,
                 stack: Optional[pulumi.Input[str]] = None,
                 team: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        opts = pulumi.ResourceOptions.merge(_utilities.get_resource_opts_defaults(), opts)
        if not isinstance(opts, pulumi.ResourceOptions):
            raise TypeError('Expected resource options to be a ResourceOptions instance')
        if opts.id is None:
            if __props__ is not None:
                raise TypeError('__props__ is only valid when passed in combination with a valid opts.id to get an existing resource')
            __props__ = TeamStackPermissionArgs.__new__(TeamStackPermissionArgs)

            if organization is None and not opts.urn:
                raise TypeError("Missing required property 'organization'")
            __props__.__dict__["organization"] = organization
            if permission is None and not opts.urn:
                raise TypeError("Missing required property 'permission'")
            __props__.__dict__["permission"] = permission
            if project is None and not opts.urn:
                raise TypeError("Missing required property 'project'")
            __props__.__dict__["project"] = project
            if stack is None and not opts.urn:
                raise TypeError("Missing required property 'stack'")
            __props__.__dict__["stack"] = stack
            if team is None and not opts.urn:
                raise TypeError("Missing required property 'team'")
            __props__.__dict__["team"] = team
        super(TeamStackPermission, __self__).__init__(
            'pulumiservice:index:TeamStackPermission',
            resource_name,
            __props__,
            opts)

    @staticmethod
    def get(resource_name: str,
            id: pulumi.Input[str],
            opts: Optional[pulumi.ResourceOptions] = None) -> 'TeamStackPermission':
        """
        Get an existing TeamStackPermission resource's state with the given name, id, and optional extra
        properties used to qualify the lookup.

        :param str resource_name: The unique name of the resulting resource.
        :param pulumi.Input[str] id: The unique provider ID of the resource to lookup.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        opts = pulumi.ResourceOptions.merge(opts, pulumi.ResourceOptions(id=id))

        __props__ = TeamStackPermissionArgs.__new__(TeamStackPermissionArgs)

        __props__.__dict__["organization"] = None
        __props__.__dict__["permission"] = None
        __props__.__dict__["project"] = None
        __props__.__dict__["stack"] = None
        __props__.__dict__["team"] = None
        return TeamStackPermission(resource_name, opts=opts, __props__=__props__)

    @property
    @pulumi.getter
    def organization(self) -> pulumi.Output[str]:
        """
        The organization or the personal account name of the stack.
        """
        return pulumi.get(self, "organization")

    @property
    @pulumi.getter
    def permission(self) -> pulumi.Output['TeamStackPermissionScope']:
        """
        Sets the permission level that this team will be granted to the stack.
        """
        return pulumi.get(self, "permission")

    @property
    @pulumi.getter
    def project(self) -> pulumi.Output[str]:
        """
        The project name for this stack.
        """
        return pulumi.get(self, "project")

    @property
    @pulumi.getter
    def stack(self) -> pulumi.Output[str]:
        """
        The name of the stack that the team will be granted permissions to.
        """
        return pulumi.get(self, "stack")

    @property
    @pulumi.getter
    def team(self) -> pulumi.Output[str]:
        """
        The name of the team to grant this stack permissions to. This is not the display name.
        """
        return pulumi.get(self, "team")

