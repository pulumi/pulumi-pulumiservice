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

__all__ = ['TeamArgs', 'Team']

@pulumi.input_type
class TeamArgs:
    def __init__(__self__, *,
                 organization_name: pulumi.Input[str],
                 team_type: pulumi.Input[str],
                 description: Optional[pulumi.Input[str]] = None,
                 display_name: Optional[pulumi.Input[str]] = None,
                 github_team_id: Optional[pulumi.Input[float]] = None,
                 members: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None,
                 name: Optional[pulumi.Input[str]] = None):
        """
        The set of arguments for constructing a Team resource.
        :param pulumi.Input[str] organization_name: The name of the Pulumi organization the team belongs to.
        :param pulumi.Input[str] team_type: The type of team. Must be either `pulumi` or `github`.
        :param pulumi.Input[str] description: Optional. Team description.
        :param pulumi.Input[str] display_name: Optional. Team display name.
        :param pulumi.Input[float] github_team_id: The GitHub ID of the team to mirror. Must be in the same GitHub organization that the Pulumi org is backed by. Required for "github" teams.
        :param pulumi.Input[Sequence[pulumi.Input[str]]] members: List of Pulumi Cloud usernames of team members.
        :param pulumi.Input[str] name: The team's name. Required for "pulumi" teams.
        """
        pulumi.set(__self__, "organization_name", organization_name)
        pulumi.set(__self__, "team_type", team_type)
        if description is not None:
            pulumi.set(__self__, "description", description)
        if display_name is not None:
            pulumi.set(__self__, "display_name", display_name)
        if github_team_id is not None:
            pulumi.set(__self__, "github_team_id", github_team_id)
        if members is not None:
            pulumi.set(__self__, "members", members)
        if name is not None:
            pulumi.set(__self__, "name", name)

    @property
    @pulumi.getter(name="organizationName")
    def organization_name(self) -> pulumi.Input[str]:
        """
        The name of the Pulumi organization the team belongs to.
        """
        return pulumi.get(self, "organization_name")

    @organization_name.setter
    def organization_name(self, value: pulumi.Input[str]):
        pulumi.set(self, "organization_name", value)

    @property
    @pulumi.getter(name="teamType")
    def team_type(self) -> pulumi.Input[str]:
        """
        The type of team. Must be either `pulumi` or `github`.
        """
        return pulumi.get(self, "team_type")

    @team_type.setter
    def team_type(self, value: pulumi.Input[str]):
        pulumi.set(self, "team_type", value)

    @property
    @pulumi.getter
    def description(self) -> Optional[pulumi.Input[str]]:
        """
        Optional. Team description.
        """
        return pulumi.get(self, "description")

    @description.setter
    def description(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "description", value)

    @property
    @pulumi.getter(name="displayName")
    def display_name(self) -> Optional[pulumi.Input[str]]:
        """
        Optional. Team display name.
        """
        return pulumi.get(self, "display_name")

    @display_name.setter
    def display_name(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "display_name", value)

    @property
    @pulumi.getter(name="githubTeamId")
    def github_team_id(self) -> Optional[pulumi.Input[float]]:
        """
        The GitHub ID of the team to mirror. Must be in the same GitHub organization that the Pulumi org is backed by. Required for "github" teams.
        """
        return pulumi.get(self, "github_team_id")

    @github_team_id.setter
    def github_team_id(self, value: Optional[pulumi.Input[float]]):
        pulumi.set(self, "github_team_id", value)

    @property
    @pulumi.getter
    def members(self) -> Optional[pulumi.Input[Sequence[pulumi.Input[str]]]]:
        """
        List of Pulumi Cloud usernames of team members.
        """
        return pulumi.get(self, "members")

    @members.setter
    def members(self, value: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]]):
        pulumi.set(self, "members", value)

    @property
    @pulumi.getter
    def name(self) -> Optional[pulumi.Input[str]]:
        """
        The team's name. Required for "pulumi" teams.
        """
        return pulumi.get(self, "name")

    @name.setter
    def name(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "name", value)


class Team(pulumi.CustomResource):
    @overload
    def __init__(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 description: Optional[pulumi.Input[str]] = None,
                 display_name: Optional[pulumi.Input[str]] = None,
                 github_team_id: Optional[pulumi.Input[float]] = None,
                 members: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None,
                 name: Optional[pulumi.Input[str]] = None,
                 organization_name: Optional[pulumi.Input[str]] = None,
                 team_type: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        """
        The Pulumi Cloud offers role-based access control (RBAC) using teams. Teams allow organization admins to assign a set of stack permissions to a group of users.

        :param str resource_name: The name of the resource.
        :param pulumi.ResourceOptions opts: Options for the resource.
        :param pulumi.Input[str] description: Optional. Team description.
        :param pulumi.Input[str] display_name: Optional. Team display name.
        :param pulumi.Input[float] github_team_id: The GitHub ID of the team to mirror. Must be in the same GitHub organization that the Pulumi org is backed by. Required for "github" teams.
        :param pulumi.Input[Sequence[pulumi.Input[str]]] members: List of Pulumi Cloud usernames of team members.
        :param pulumi.Input[str] name: The team's name. Required for "pulumi" teams.
        :param pulumi.Input[str] organization_name: The name of the Pulumi organization the team belongs to.
        :param pulumi.Input[str] team_type: The type of team. Must be either `pulumi` or `github`.
        """
        ...
    @overload
    def __init__(__self__,
                 resource_name: str,
                 args: TeamArgs,
                 opts: Optional[pulumi.ResourceOptions] = None):
        """
        The Pulumi Cloud offers role-based access control (RBAC) using teams. Teams allow organization admins to assign a set of stack permissions to a group of users.

        :param str resource_name: The name of the resource.
        :param TeamArgs args: The arguments to use to populate this resource's properties.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        ...
    def __init__(__self__, resource_name: str, *args, **kwargs):
        resource_args, opts = _utilities.get_resource_args_opts(TeamArgs, pulumi.ResourceOptions, *args, **kwargs)
        if resource_args is not None:
            __self__._internal_init(resource_name, opts, **resource_args.__dict__)
        else:
            __self__._internal_init(resource_name, *args, **kwargs)

    def _internal_init(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 description: Optional[pulumi.Input[str]] = None,
                 display_name: Optional[pulumi.Input[str]] = None,
                 github_team_id: Optional[pulumi.Input[float]] = None,
                 members: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None,
                 name: Optional[pulumi.Input[str]] = None,
                 organization_name: Optional[pulumi.Input[str]] = None,
                 team_type: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        opts = pulumi.ResourceOptions.merge(_utilities.get_resource_opts_defaults(), opts)
        if not isinstance(opts, pulumi.ResourceOptions):
            raise TypeError('Expected resource options to be a ResourceOptions instance')
        if opts.id is None:
            if __props__ is not None:
                raise TypeError('__props__ is only valid when passed in combination with a valid opts.id to get an existing resource')
            __props__ = TeamArgs.__new__(TeamArgs)

            __props__.__dict__["description"] = description
            __props__.__dict__["display_name"] = display_name
            __props__.__dict__["github_team_id"] = github_team_id
            __props__.__dict__["members"] = members
            __props__.__dict__["name"] = name
            if organization_name is None and not opts.urn:
                raise TypeError("Missing required property 'organization_name'")
            __props__.__dict__["organization_name"] = organization_name
            if team_type is None and not opts.urn:
                raise TypeError("Missing required property 'team_type'")
            __props__.__dict__["team_type"] = team_type
        super(Team, __self__).__init__(
            'pulumiservice:index:Team',
            resource_name,
            __props__,
            opts)

    @staticmethod
    def get(resource_name: str,
            id: pulumi.Input[str],
            opts: Optional[pulumi.ResourceOptions] = None) -> 'Team':
        """
        Get an existing Team resource's state with the given name, id, and optional extra
        properties used to qualify the lookup.

        :param str resource_name: The unique name of the resulting resource.
        :param pulumi.Input[str] id: The unique provider ID of the resource to lookup.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        opts = pulumi.ResourceOptions.merge(opts, pulumi.ResourceOptions(id=id))

        __props__ = TeamArgs.__new__(TeamArgs)

        __props__.__dict__["description"] = None
        __props__.__dict__["display_name"] = None
        __props__.__dict__["github_team_id"] = None
        __props__.__dict__["members"] = None
        __props__.__dict__["name"] = None
        __props__.__dict__["organization_name"] = None
        __props__.__dict__["team_type"] = None
        return Team(resource_name, opts=opts, __props__=__props__)

    @property
    @pulumi.getter
    def description(self) -> pulumi.Output[Optional[str]]:
        """
        Optional. Team description.
        """
        return pulumi.get(self, "description")

    @property
    @pulumi.getter(name="displayName")
    def display_name(self) -> pulumi.Output[Optional[str]]:
        """
        Optional. Team display name.
        """
        return pulumi.get(self, "display_name")

    @property
    @pulumi.getter(name="githubTeamId")
    def github_team_id(self) -> pulumi.Output[Optional[float]]:
        """
        The GitHub ID of the team to mirror. Must be in the same GitHub organization that the Pulumi org is backed by. Required for "github" teams.
        """
        return pulumi.get(self, "github_team_id")

    @property
    @pulumi.getter
    def members(self) -> pulumi.Output[Sequence[str]]:
        """
        List of Pulumi Cloud usernames of team members.
        """
        return pulumi.get(self, "members")

    @property
    @pulumi.getter
    def name(self) -> pulumi.Output[Optional[str]]:
        """
        The team's name. Required for "pulumi" teams.
        """
        return pulumi.get(self, "name")

    @property
    @pulumi.getter(name="organizationName")
    def organization_name(self) -> pulumi.Output[str]:
        """
        The name of the Pulumi organization the team belongs to.
        """
        return pulumi.get(self, "organization_name")

    @property
    @pulumi.getter(name="teamType")
    def team_type(self) -> pulumi.Output[str]:
        """
        The type of team. Must be either `pulumi` or `github`.
        """
        return pulumi.get(self, "team_type")

