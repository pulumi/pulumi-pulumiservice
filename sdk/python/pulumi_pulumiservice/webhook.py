# coding=utf-8
# *** WARNING: this file was generated by pulumigen. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

import warnings
import pulumi
import pulumi.runtime
from typing import Any, Mapping, Optional, Sequence, Union, overload
from . import _utilities

__all__ = ['WebhookArgs', 'Webhook']

@pulumi.input_type
class WebhookArgs:
    def __init__(__self__, *,
                 active: pulumi.Input[bool],
                 display_name: pulumi.Input[str],
                 organization_name: pulumi.Input[str],
                 payload_url: pulumi.Input[str],
                 secret: Optional[pulumi.Input[str]] = None):
        """
        The set of arguments for constructing a Webhook resource.
        :param pulumi.Input[bool] active: Indicates whether this webhook is enabled or not.
        :param pulumi.Input[str] display_name: The friendly name displayed in the Pulumi Service.
        :param pulumi.Input[str] organization_name: Name of the organization.
        :param pulumi.Input[str] payload_url: URL to send request to.
        :param pulumi.Input[str] secret: Optional. secret used as the HMAC key. See [webhook docs](https://www.pulumi.com/docs/intro/pulumi-service/webhooks/#headers) for more information.
        """
        pulumi.set(__self__, "active", active)
        pulumi.set(__self__, "display_name", display_name)
        pulumi.set(__self__, "organization_name", organization_name)
        pulumi.set(__self__, "payload_url", payload_url)
        if secret is not None:
            pulumi.set(__self__, "secret", secret)

    @property
    @pulumi.getter
    def active(self) -> pulumi.Input[bool]:
        """
        Indicates whether this webhook is enabled or not.
        """
        return pulumi.get(self, "active")

    @active.setter
    def active(self, value: pulumi.Input[bool]):
        pulumi.set(self, "active", value)

    @property
    @pulumi.getter(name="displayName")
    def display_name(self) -> pulumi.Input[str]:
        """
        The friendly name displayed in the Pulumi Service.
        """
        return pulumi.get(self, "display_name")

    @display_name.setter
    def display_name(self, value: pulumi.Input[str]):
        pulumi.set(self, "display_name", value)

    @property
    @pulumi.getter(name="organizationName")
    def organization_name(self) -> pulumi.Input[str]:
        """
        Name of the organization.
        """
        return pulumi.get(self, "organization_name")

    @organization_name.setter
    def organization_name(self, value: pulumi.Input[str]):
        pulumi.set(self, "organization_name", value)

    @property
    @pulumi.getter(name="payloadUrl")
    def payload_url(self) -> pulumi.Input[str]:
        """
        URL to send request to.
        """
        return pulumi.get(self, "payload_url")

    @payload_url.setter
    def payload_url(self, value: pulumi.Input[str]):
        pulumi.set(self, "payload_url", value)

    @property
    @pulumi.getter
    def secret(self) -> Optional[pulumi.Input[str]]:
        """
        Optional. secret used as the HMAC key. See [webhook docs](https://www.pulumi.com/docs/intro/pulumi-service/webhooks/#headers) for more information.
        """
        return pulumi.get(self, "secret")

    @secret.setter
    def secret(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "secret", value)


class Webhook(pulumi.CustomResource):
    @overload
    def __init__(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 active: Optional[pulumi.Input[bool]] = None,
                 display_name: Optional[pulumi.Input[str]] = None,
                 organization_name: Optional[pulumi.Input[str]] = None,
                 payload_url: Optional[pulumi.Input[str]] = None,
                 secret: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        """
        Pulumi Webhooks allow you to notify external services of events happening within your Pulumi organization or stack. For example, you can trigger a notification whenever a stack is updated. Whenever an event occurs, Pulumi will send an HTTP POST request to all registered webhooks. The webhook can then be used to emit some notification, start running integration tests, or even update additional stacks.

        :param str resource_name: The name of the resource.
        :param pulumi.ResourceOptions opts: Options for the resource.
        :param pulumi.Input[bool] active: Indicates whether this webhook is enabled or not.
        :param pulumi.Input[str] display_name: The friendly name displayed in the Pulumi Service.
        :param pulumi.Input[str] organization_name: Name of the organization.
        :param pulumi.Input[str] payload_url: URL to send request to.
        :param pulumi.Input[str] secret: Optional. secret used as the HMAC key. See [webhook docs](https://www.pulumi.com/docs/intro/pulumi-service/webhooks/#headers) for more information.
        """
        ...
    @overload
    def __init__(__self__,
                 resource_name: str,
                 args: WebhookArgs,
                 opts: Optional[pulumi.ResourceOptions] = None):
        """
        Pulumi Webhooks allow you to notify external services of events happening within your Pulumi organization or stack. For example, you can trigger a notification whenever a stack is updated. Whenever an event occurs, Pulumi will send an HTTP POST request to all registered webhooks. The webhook can then be used to emit some notification, start running integration tests, or even update additional stacks.

        :param str resource_name: The name of the resource.
        :param WebhookArgs args: The arguments to use to populate this resource's properties.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        ...
    def __init__(__self__, resource_name: str, *args, **kwargs):
        resource_args, opts = _utilities.get_resource_args_opts(WebhookArgs, pulumi.ResourceOptions, *args, **kwargs)
        if resource_args is not None:
            __self__._internal_init(resource_name, opts, **resource_args.__dict__)
        else:
            __self__._internal_init(resource_name, *args, **kwargs)

    def _internal_init(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 active: Optional[pulumi.Input[bool]] = None,
                 display_name: Optional[pulumi.Input[str]] = None,
                 organization_name: Optional[pulumi.Input[str]] = None,
                 payload_url: Optional[pulumi.Input[str]] = None,
                 secret: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        if opts is None:
            opts = pulumi.ResourceOptions()
        if not isinstance(opts, pulumi.ResourceOptions):
            raise TypeError('Expected resource options to be a ResourceOptions instance')
        if opts.version is None:
            opts.version = _utilities.get_version()
        if opts.id is None:
            if __props__ is not None:
                raise TypeError('__props__ is only valid when passed in combination with a valid opts.id to get an existing resource')
            __props__ = WebhookArgs.__new__(WebhookArgs)

            if active is None and not opts.urn:
                raise TypeError("Missing required property 'active'")
            __props__.__dict__["active"] = active
            if display_name is None and not opts.urn:
                raise TypeError("Missing required property 'display_name'")
            __props__.__dict__["display_name"] = display_name
            if organization_name is None and not opts.urn:
                raise TypeError("Missing required property 'organization_name'")
            __props__.__dict__["organization_name"] = organization_name
            if payload_url is None and not opts.urn:
                raise TypeError("Missing required property 'payload_url'")
            __props__.__dict__["payload_url"] = payload_url
            __props__.__dict__["secret"] = None if secret is None else pulumi.Output.secret(secret)
            __props__.__dict__["name"] = None
        secret_opts = pulumi.ResourceOptions(additional_secret_outputs=["secret"])
        opts = pulumi.ResourceOptions.merge(opts, secret_opts)
        super(Webhook, __self__).__init__(
            'pulumiservice:index:Webhook',
            resource_name,
            __props__,
            opts)

    @staticmethod
    def get(resource_name: str,
            id: pulumi.Input[str],
            opts: Optional[pulumi.ResourceOptions] = None) -> 'Webhook':
        """
        Get an existing Webhook resource's state with the given name, id, and optional extra
        properties used to qualify the lookup.

        :param str resource_name: The unique name of the resulting resource.
        :param pulumi.Input[str] id: The unique provider ID of the resource to lookup.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        opts = pulumi.ResourceOptions.merge(opts, pulumi.ResourceOptions(id=id))

        __props__ = WebhookArgs.__new__(WebhookArgs)

        __props__.__dict__["active"] = None
        __props__.__dict__["display_name"] = None
        __props__.__dict__["name"] = None
        __props__.__dict__["organization_name"] = None
        __props__.__dict__["payload_url"] = None
        __props__.__dict__["secret"] = None
        return Webhook(resource_name, opts=opts, __props__=__props__)

    @property
    @pulumi.getter
    def active(self) -> pulumi.Output[Optional[bool]]:
        """
        Indicates whether this webhook is enabled or not.
        """
        return pulumi.get(self, "active")

    @property
    @pulumi.getter(name="displayName")
    def display_name(self) -> pulumi.Output[Optional[str]]:
        """
        The friendly name displayed in the Pulumi Service.
        """
        return pulumi.get(self, "display_name")

    @property
    @pulumi.getter
    def name(self) -> pulumi.Output[Optional[str]]:
        """
        Webhook identifier generated by pulumi service.
        """
        return pulumi.get(self, "name")

    @property
    @pulumi.getter(name="organizationName")
    def organization_name(self) -> pulumi.Output[Optional[str]]:
        """
        Name of the organization.
        """
        return pulumi.get(self, "organization_name")

    @property
    @pulumi.getter(name="payloadUrl")
    def payload_url(self) -> pulumi.Output[Optional[str]]:
        """
        URL to send request to.
        """
        return pulumi.get(self, "payload_url")

    @property
    @pulumi.getter
    def secret(self) -> pulumi.Output[Optional[str]]:
        """
        Optional. secret used as the HMAC key. See [webhook docs](https://www.pulumi.com/docs/intro/pulumi-service/webhooks/#headers) for more information.
        """
        return pulumi.get(self, "secret")

