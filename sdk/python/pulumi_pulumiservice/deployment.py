# coding=utf-8
# *** WARNING: this file was generated by pulumigen. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

import copy
import warnings
import pulumi
import pulumi.runtime
from typing import Any, Mapping, Optional, Sequence, Union, overload
from . import _utilities
from ._inputs import *

__all__ = ['DeploymentArgs', 'Deployment']

@pulumi.input_type
class DeploymentArgs:
    def __init__(__self__, *,
                 stack: pulumi.Input[str],
                 config: Optional[pulumi.Input[Mapping[str, Any]]] = None,
                 settings: Optional[pulumi.Input['StackDeploymentSettingsArgs']] = None):
        """
        The set of arguments for constructing a Deployment resource.
        :param pulumi.Input[str] stack: The fully-qualified name of the stack to deploy.
        :param pulumi.Input[Mapping[str, Any]] config: The stack configuration to use for the deployment.
        :param pulumi.Input['StackDeploymentSettingsArgs'] settings: The settings to use for the deployment.
        """
        pulumi.set(__self__, "stack", stack)
        if config is not None:
            pulumi.set(__self__, "config", config)
        if settings is not None:
            pulumi.set(__self__, "settings", settings)

    @property
    @pulumi.getter
    def stack(self) -> pulumi.Input[str]:
        """
        The fully-qualified name of the stack to deploy.
        """
        return pulumi.get(self, "stack")

    @stack.setter
    def stack(self, value: pulumi.Input[str]):
        pulumi.set(self, "stack", value)

    @property
    @pulumi.getter
    def config(self) -> Optional[pulumi.Input[Mapping[str, Any]]]:
        """
        The stack configuration to use for the deployment.
        """
        return pulumi.get(self, "config")

    @config.setter
    def config(self, value: Optional[pulumi.Input[Mapping[str, Any]]]):
        pulumi.set(self, "config", value)

    @property
    @pulumi.getter
    def settings(self) -> Optional[pulumi.Input['StackDeploymentSettingsArgs']]:
        """
        The settings to use for the deployment.
        """
        return pulumi.get(self, "settings")

    @settings.setter
    def settings(self, value: Optional[pulumi.Input['StackDeploymentSettingsArgs']]):
        pulumi.set(self, "settings", value)


class Deployment(pulumi.CustomResource):
    @overload
    def __init__(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 config: Optional[pulumi.Input[Mapping[str, Any]]] = None,
                 settings: Optional[pulumi.Input[pulumi.InputType['StackDeploymentSettingsArgs']]] = None,
                 stack: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        """
        A deployment deploys a stack. Deployments will always have changes.

        :param str resource_name: The name of the resource.
        :param pulumi.ResourceOptions opts: Options for the resource.
        :param pulumi.Input[Mapping[str, Any]] config: The stack configuration to use for the deployment.
        :param pulumi.Input[pulumi.InputType['StackDeploymentSettingsArgs']] settings: The settings to use for the deployment.
        :param pulumi.Input[str] stack: The fully-qualified name of the stack to deploy.
        """
        ...
    @overload
    def __init__(__self__,
                 resource_name: str,
                 args: DeploymentArgs,
                 opts: Optional[pulumi.ResourceOptions] = None):
        """
        A deployment deploys a stack. Deployments will always have changes.

        :param str resource_name: The name of the resource.
        :param DeploymentArgs args: The arguments to use to populate this resource's properties.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        ...
    def __init__(__self__, resource_name: str, *args, **kwargs):
        resource_args, opts = _utilities.get_resource_args_opts(DeploymentArgs, pulumi.ResourceOptions, *args, **kwargs)
        if resource_args is not None:
            __self__._internal_init(resource_name, opts, **resource_args.__dict__)
        else:
            __self__._internal_init(resource_name, *args, **kwargs)

    def _internal_init(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 config: Optional[pulumi.Input[Mapping[str, Any]]] = None,
                 settings: Optional[pulumi.Input[pulumi.InputType['StackDeploymentSettingsArgs']]] = None,
                 stack: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        opts = pulumi.ResourceOptions.merge(_utilities.get_resource_opts_defaults(), opts)
        if not isinstance(opts, pulumi.ResourceOptions):
            raise TypeError('Expected resource options to be a ResourceOptions instance')
        if opts.id is None:
            if __props__ is not None:
                raise TypeError('__props__ is only valid when passed in combination with a valid opts.id to get an existing resource')
            __props__ = DeploymentArgs.__new__(DeploymentArgs)

            __props__.__dict__["config"] = config
            __props__.__dict__["settings"] = settings
            if stack is None and not opts.urn:
                raise TypeError("Missing required property 'stack'")
            __props__.__dict__["stack"] = stack
            __props__.__dict__["outputs"] = None
            __props__.__dict__["version"] = None
        super(Deployment, __self__).__init__(
            'pulumiservice:index:Deployment',
            resource_name,
            __props__,
            opts)

    @staticmethod
    def get(resource_name: str,
            id: pulumi.Input[str],
            opts: Optional[pulumi.ResourceOptions] = None) -> 'Deployment':
        """
        Get an existing Deployment resource's state with the given name, id, and optional extra
        properties used to qualify the lookup.

        :param str resource_name: The unique name of the resulting resource.
        :param pulumi.Input[str] id: The unique provider ID of the resource to lookup.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        opts = pulumi.ResourceOptions.merge(opts, pulumi.ResourceOptions(id=id))

        __props__ = DeploymentArgs.__new__(DeploymentArgs)

        __props__.__dict__["outputs"] = None
        __props__.__dict__["stack"] = None
        __props__.__dict__["version"] = None
        return Deployment(resource_name, opts=opts, __props__=__props__)

    @property
    @pulumi.getter
    def outputs(self) -> pulumi.Output[Mapping[str, Any]]:
        """
        The outputs of the deployed stack.
        """
        return pulumi.get(self, "outputs")

    @property
    @pulumi.getter
    def stack(self) -> pulumi.Output[str]:
        """
        The fully-qualified name of the stack to deploy.
        """
        return pulumi.get(self, "stack")

    @property
    @pulumi.getter
    def version(self) -> pulumi.Output[int]:
        """
        The version of the deployed stack.
        """
        return pulumi.get(self, "version")
