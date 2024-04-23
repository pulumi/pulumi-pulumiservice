# coding=utf-8
# *** WARNING: this file was generated by pulumi-language-python. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

import copy
import warnings
import pulumi
import pulumi.runtime
from typing import Any, Mapping, Optional, Sequence, Union, overload
from . import _utilities

__all__ = ['DriftScheduleArgs', 'DriftSchedule']

@pulumi.input_type
class DriftScheduleArgs:
    def __init__(__self__, *,
                 organization: pulumi.Input[str],
                 project: pulumi.Input[str],
                 schedule_cron: pulumi.Input[str],
                 stack: pulumi.Input[str],
                 auto_remediate: Optional[pulumi.Input[bool]] = None):
        """
        The set of arguments for constructing a DriftSchedule resource.
        :param pulumi.Input[str] organization: Organization name.
        :param pulumi.Input[str] project: Project name.
        :param pulumi.Input[str] schedule_cron: Cron expression for when to run drift detection.
        :param pulumi.Input[str] stack: Stack name.
        :param pulumi.Input[bool] auto_remediate: Whether any drift detected should be remediated after a drift run.
        """
        pulumi.set(__self__, "organization", organization)
        pulumi.set(__self__, "project", project)
        pulumi.set(__self__, "schedule_cron", schedule_cron)
        pulumi.set(__self__, "stack", stack)
        if auto_remediate is not None:
            pulumi.set(__self__, "auto_remediate", auto_remediate)

    @property
    @pulumi.getter
    def organization(self) -> pulumi.Input[str]:
        """
        Organization name.
        """
        return pulumi.get(self, "organization")

    @organization.setter
    def organization(self, value: pulumi.Input[str]):
        pulumi.set(self, "organization", value)

    @property
    @pulumi.getter
    def project(self) -> pulumi.Input[str]:
        """
        Project name.
        """
        return pulumi.get(self, "project")

    @project.setter
    def project(self, value: pulumi.Input[str]):
        pulumi.set(self, "project", value)

    @property
    @pulumi.getter(name="scheduleCron")
    def schedule_cron(self) -> pulumi.Input[str]:
        """
        Cron expression for when to run drift detection.
        """
        return pulumi.get(self, "schedule_cron")

    @schedule_cron.setter
    def schedule_cron(self, value: pulumi.Input[str]):
        pulumi.set(self, "schedule_cron", value)

    @property
    @pulumi.getter
    def stack(self) -> pulumi.Input[str]:
        """
        Stack name.
        """
        return pulumi.get(self, "stack")

    @stack.setter
    def stack(self, value: pulumi.Input[str]):
        pulumi.set(self, "stack", value)

    @property
    @pulumi.getter(name="autoRemediate")
    def auto_remediate(self) -> Optional[pulumi.Input[bool]]:
        """
        Whether any drift detected should be remediated after a drift run.
        """
        return pulumi.get(self, "auto_remediate")

    @auto_remediate.setter
    def auto_remediate(self, value: Optional[pulumi.Input[bool]]):
        pulumi.set(self, "auto_remediate", value)


class DriftSchedule(pulumi.CustomResource):
    @overload
    def __init__(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 auto_remediate: Optional[pulumi.Input[bool]] = None,
                 organization: Optional[pulumi.Input[str]] = None,
                 project: Optional[pulumi.Input[str]] = None,
                 schedule_cron: Optional[pulumi.Input[str]] = None,
                 stack: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        """
        A cron schedule to run drift detection.

        :param str resource_name: The name of the resource.
        :param pulumi.ResourceOptions opts: Options for the resource.
        :param pulumi.Input[bool] auto_remediate: Whether any drift detected should be remediated after a drift run.
        :param pulumi.Input[str] organization: Organization name.
        :param pulumi.Input[str] project: Project name.
        :param pulumi.Input[str] schedule_cron: Cron expression for when to run drift detection.
        :param pulumi.Input[str] stack: Stack name.
        """
        ...
    @overload
    def __init__(__self__,
                 resource_name: str,
                 args: DriftScheduleArgs,
                 opts: Optional[pulumi.ResourceOptions] = None):
        """
        A cron schedule to run drift detection.

        :param str resource_name: The name of the resource.
        :param DriftScheduleArgs args: The arguments to use to populate this resource's properties.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        ...
    def __init__(__self__, resource_name: str, *args, **kwargs):
        resource_args, opts = _utilities.get_resource_args_opts(DriftScheduleArgs, pulumi.ResourceOptions, *args, **kwargs)
        if resource_args is not None:
            __self__._internal_init(resource_name, opts, **resource_args.__dict__)
        else:
            __self__._internal_init(resource_name, *args, **kwargs)

    def _internal_init(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 auto_remediate: Optional[pulumi.Input[bool]] = None,
                 organization: Optional[pulumi.Input[str]] = None,
                 project: Optional[pulumi.Input[str]] = None,
                 schedule_cron: Optional[pulumi.Input[str]] = None,
                 stack: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        opts = pulumi.ResourceOptions.merge(_utilities.get_resource_opts_defaults(), opts)
        if not isinstance(opts, pulumi.ResourceOptions):
            raise TypeError('Expected resource options to be a ResourceOptions instance')
        if opts.id is None:
            if __props__ is not None:
                raise TypeError('__props__ is only valid when passed in combination with a valid opts.id to get an existing resource')
            __props__ = DriftScheduleArgs.__new__(DriftScheduleArgs)

            __props__.__dict__["auto_remediate"] = auto_remediate
            if organization is None and not opts.urn:
                raise TypeError("Missing required property 'organization'")
            __props__.__dict__["organization"] = organization
            if project is None and not opts.urn:
                raise TypeError("Missing required property 'project'")
            __props__.__dict__["project"] = project
            if schedule_cron is None and not opts.urn:
                raise TypeError("Missing required property 'schedule_cron'")
            __props__.__dict__["schedule_cron"] = schedule_cron
            if stack is None and not opts.urn:
                raise TypeError("Missing required property 'stack'")
            __props__.__dict__["stack"] = stack
            __props__.__dict__["schedule_id"] = None
        super(DriftSchedule, __self__).__init__(
            'pulumiservice:index:DriftSchedule',
            resource_name,
            __props__,
            opts)

    @staticmethod
    def get(resource_name: str,
            id: pulumi.Input[str],
            opts: Optional[pulumi.ResourceOptions] = None) -> 'DriftSchedule':
        """
        Get an existing DriftSchedule resource's state with the given name, id, and optional extra
        properties used to qualify the lookup.

        :param str resource_name: The unique name of the resulting resource.
        :param pulumi.Input[str] id: The unique provider ID of the resource to lookup.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        opts = pulumi.ResourceOptions.merge(opts, pulumi.ResourceOptions(id=id))

        __props__ = DriftScheduleArgs.__new__(DriftScheduleArgs)

        __props__.__dict__["auto_remediate"] = None
        __props__.__dict__["organization"] = None
        __props__.__dict__["project"] = None
        __props__.__dict__["schedule_cron"] = None
        __props__.__dict__["schedule_id"] = None
        __props__.__dict__["stack"] = None
        return DriftSchedule(resource_name, opts=opts, __props__=__props__)

    @property
    @pulumi.getter(name="autoRemediate")
    def auto_remediate(self) -> pulumi.Output[Optional[bool]]:
        """
        Whether any drift detected should be remediated after a drift run.
        """
        return pulumi.get(self, "auto_remediate")

    @property
    @pulumi.getter
    def organization(self) -> pulumi.Output[str]:
        """
        Organization name.
        """
        return pulumi.get(self, "organization")

    @property
    @pulumi.getter
    def project(self) -> pulumi.Output[str]:
        """
        Project name.
        """
        return pulumi.get(self, "project")

    @property
    @pulumi.getter(name="scheduleCron")
    def schedule_cron(self) -> pulumi.Output[str]:
        """
        Cron expression for when to run drift detection.
        """
        return pulumi.get(self, "schedule_cron")

    @property
    @pulumi.getter(name="scheduleId")
    def schedule_id(self) -> pulumi.Output[str]:
        """
        Schedule ID of the created schedule, assigned by Pulumi Cloud.
        """
        return pulumi.get(self, "schedule_id")

    @property
    @pulumi.getter
    def stack(self) -> pulumi.Output[str]:
        """
        Stack name.
        """
        return pulumi.get(self, "stack")

