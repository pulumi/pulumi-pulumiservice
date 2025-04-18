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
from . import outputs
from ._enums import *
from ._inputs import *

__all__ = ['OidcIssuerArgs', 'OidcIssuer']

@pulumi.input_type
class OidcIssuerArgs:
    def __init__(__self__, *,
                 name: pulumi.Input[str],
                 organization: pulumi.Input[str],
                 url: pulumi.Input[str],
                 max_expiration_seconds: Optional[pulumi.Input[int]] = None,
                 policies: Optional[pulumi.Input[Sequence[pulumi.Input['AuthPolicyDefinitionArgs']]]] = None,
                 thumbprints: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None):
        """
        The set of arguments for constructing a OidcIssuer resource.
        :param pulumi.Input[str] name: Issuer name.
        :param pulumi.Input[str] organization: Organization name.
        :param pulumi.Input[str] url: The OIDC issuer URL.
        :param pulumi.Input[int] max_expiration_seconds: The maximum duration of the Pulumi access token working after an exchange, specified in seconds.
        :param pulumi.Input[Sequence[pulumi.Input['AuthPolicyDefinitionArgs']]] policies: The auth policies for this Oidc Issuer.
        :param pulumi.Input[Sequence[pulumi.Input[str]]] thumbprints: The thumbprints of issuer's TLS certificates. By default, Pulumi will store the thumbprint of the certificate used to serve the OpenID configuration. If the provider uses multiple certificates to serve content, it is required to manually configure these.
        """
        pulumi.set(__self__, "name", name)
        pulumi.set(__self__, "organization", organization)
        pulumi.set(__self__, "url", url)
        if max_expiration_seconds is not None:
            pulumi.set(__self__, "max_expiration_seconds", max_expiration_seconds)
        if policies is not None:
            pulumi.set(__self__, "policies", policies)
        if thumbprints is not None:
            pulumi.set(__self__, "thumbprints", thumbprints)

    @property
    @pulumi.getter
    def name(self) -> pulumi.Input[str]:
        """
        Issuer name.
        """
        return pulumi.get(self, "name")

    @name.setter
    def name(self, value: pulumi.Input[str]):
        pulumi.set(self, "name", value)

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
    def url(self) -> pulumi.Input[str]:
        """
        The OIDC issuer URL.
        """
        return pulumi.get(self, "url")

    @url.setter
    def url(self, value: pulumi.Input[str]):
        pulumi.set(self, "url", value)

    @property
    @pulumi.getter(name="maxExpirationSeconds")
    def max_expiration_seconds(self) -> Optional[pulumi.Input[int]]:
        """
        The maximum duration of the Pulumi access token working after an exchange, specified in seconds.
        """
        return pulumi.get(self, "max_expiration_seconds")

    @max_expiration_seconds.setter
    def max_expiration_seconds(self, value: Optional[pulumi.Input[int]]):
        pulumi.set(self, "max_expiration_seconds", value)

    @property
    @pulumi.getter
    def policies(self) -> Optional[pulumi.Input[Sequence[pulumi.Input['AuthPolicyDefinitionArgs']]]]:
        """
        The auth policies for this Oidc Issuer.
        """
        return pulumi.get(self, "policies")

    @policies.setter
    def policies(self, value: Optional[pulumi.Input[Sequence[pulumi.Input['AuthPolicyDefinitionArgs']]]]):
        pulumi.set(self, "policies", value)

    @property
    @pulumi.getter
    def thumbprints(self) -> Optional[pulumi.Input[Sequence[pulumi.Input[str]]]]:
        """
        The thumbprints of issuer's TLS certificates. By default, Pulumi will store the thumbprint of the certificate used to serve the OpenID configuration. If the provider uses multiple certificates to serve content, it is required to manually configure these.
        """
        return pulumi.get(self, "thumbprints")

    @thumbprints.setter
    def thumbprints(self, value: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]]):
        pulumi.set(self, "thumbprints", value)


class OidcIssuer(pulumi.CustomResource):
    @overload
    def __init__(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 max_expiration_seconds: Optional[pulumi.Input[int]] = None,
                 name: Optional[pulumi.Input[str]] = None,
                 organization: Optional[pulumi.Input[str]] = None,
                 policies: Optional[pulumi.Input[Sequence[pulumi.Input[Union['AuthPolicyDefinitionArgs', 'AuthPolicyDefinitionArgsDict']]]]] = None,
                 thumbprints: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None,
                 url: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        """
        Register an OIDC Provider to establish a trust relationship between third-party systems like GitHub Actions and Pulumi Cloud, obviating the need to store a hard-coded Pulumi Cloud token in systems that need to run Pulumi commands or consume Pulumi Cloud APIs. Instead of a hard-coded, static token that must be manually rotated, trusted systems are granted temporary Pulumi Cloud tokens on an as-needed basis, which is more secure than static tokens.

        :param str resource_name: The name of the resource.
        :param pulumi.ResourceOptions opts: Options for the resource.
        :param pulumi.Input[int] max_expiration_seconds: The maximum duration of the Pulumi access token working after an exchange, specified in seconds.
        :param pulumi.Input[str] name: Issuer name.
        :param pulumi.Input[str] organization: Organization name.
        :param pulumi.Input[Sequence[pulumi.Input[Union['AuthPolicyDefinitionArgs', 'AuthPolicyDefinitionArgsDict']]]] policies: The auth policies for this Oidc Issuer.
        :param pulumi.Input[Sequence[pulumi.Input[str]]] thumbprints: The thumbprints of issuer's TLS certificates. By default, Pulumi will store the thumbprint of the certificate used to serve the OpenID configuration. If the provider uses multiple certificates to serve content, it is required to manually configure these.
        :param pulumi.Input[str] url: The OIDC issuer URL.
        """
        ...
    @overload
    def __init__(__self__,
                 resource_name: str,
                 args: OidcIssuerArgs,
                 opts: Optional[pulumi.ResourceOptions] = None):
        """
        Register an OIDC Provider to establish a trust relationship between third-party systems like GitHub Actions and Pulumi Cloud, obviating the need to store a hard-coded Pulumi Cloud token in systems that need to run Pulumi commands or consume Pulumi Cloud APIs. Instead of a hard-coded, static token that must be manually rotated, trusted systems are granted temporary Pulumi Cloud tokens on an as-needed basis, which is more secure than static tokens.

        :param str resource_name: The name of the resource.
        :param OidcIssuerArgs args: The arguments to use to populate this resource's properties.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        ...
    def __init__(__self__, resource_name: str, *args, **kwargs):
        resource_args, opts = _utilities.get_resource_args_opts(OidcIssuerArgs, pulumi.ResourceOptions, *args, **kwargs)
        if resource_args is not None:
            __self__._internal_init(resource_name, opts, **resource_args.__dict__)
        else:
            __self__._internal_init(resource_name, *args, **kwargs)

    def _internal_init(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 max_expiration_seconds: Optional[pulumi.Input[int]] = None,
                 name: Optional[pulumi.Input[str]] = None,
                 organization: Optional[pulumi.Input[str]] = None,
                 policies: Optional[pulumi.Input[Sequence[pulumi.Input[Union['AuthPolicyDefinitionArgs', 'AuthPolicyDefinitionArgsDict']]]]] = None,
                 thumbprints: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None,
                 url: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        opts = pulumi.ResourceOptions.merge(_utilities.get_resource_opts_defaults(), opts)
        if not isinstance(opts, pulumi.ResourceOptions):
            raise TypeError('Expected resource options to be a ResourceOptions instance')
        if opts.id is None:
            if __props__ is not None:
                raise TypeError('__props__ is only valid when passed in combination with a valid opts.id to get an existing resource')
            __props__ = OidcIssuerArgs.__new__(OidcIssuerArgs)

            __props__.__dict__["max_expiration_seconds"] = max_expiration_seconds
            if name is None and not opts.urn:
                raise TypeError("Missing required property 'name'")
            __props__.__dict__["name"] = name
            if organization is None and not opts.urn:
                raise TypeError("Missing required property 'organization'")
            __props__.__dict__["organization"] = organization
            __props__.__dict__["policies"] = policies
            __props__.__dict__["thumbprints"] = thumbprints
            if url is None and not opts.urn:
                raise TypeError("Missing required property 'url'")
            __props__.__dict__["url"] = url
        replace_on_changes = pulumi.ResourceOptions(replace_on_changes=["organization", "url"])
        opts = pulumi.ResourceOptions.merge(opts, replace_on_changes)
        super(OidcIssuer, __self__).__init__(
            'pulumiservice:index:OidcIssuer',
            resource_name,
            __props__,
            opts)

    @staticmethod
    def get(resource_name: str,
            id: pulumi.Input[str],
            opts: Optional[pulumi.ResourceOptions] = None) -> 'OidcIssuer':
        """
        Get an existing OidcIssuer resource's state with the given name, id, and optional extra
        properties used to qualify the lookup.

        :param str resource_name: The unique name of the resulting resource.
        :param pulumi.Input[str] id: The unique provider ID of the resource to lookup.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        opts = pulumi.ResourceOptions.merge(opts, pulumi.ResourceOptions(id=id))

        __props__ = OidcIssuerArgs.__new__(OidcIssuerArgs)

        __props__.__dict__["max_expiration_seconds"] = None
        __props__.__dict__["name"] = None
        __props__.__dict__["organization"] = None
        __props__.__dict__["policies"] = None
        __props__.__dict__["thumbprints"] = None
        __props__.__dict__["url"] = None
        return OidcIssuer(resource_name, opts=opts, __props__=__props__)

    @property
    @pulumi.getter(name="maxExpirationSeconds")
    def max_expiration_seconds(self) -> pulumi.Output[Optional[int]]:
        """
        The maximum duration of the Pulumi access token working after an exchange, specified in seconds.
        """
        return pulumi.get(self, "max_expiration_seconds")

    @property
    @pulumi.getter
    def name(self) -> pulumi.Output[str]:
        """
        Issuer name.
        """
        return pulumi.get(self, "name")

    @property
    @pulumi.getter
    def organization(self) -> pulumi.Output[str]:
        """
        Organization name.
        """
        return pulumi.get(self, "organization")

    @property
    @pulumi.getter
    def policies(self) -> pulumi.Output[Sequence['outputs.AuthPolicyDefinition']]:
        """
        The authorization policies for this Oidc Issuer.
        """
        return pulumi.get(self, "policies")

    @property
    @pulumi.getter
    def thumbprints(self) -> pulumi.Output[Sequence[str]]:
        """
        The thumbprints of issuer's TLS certificates. By default, Pulumi will store the thumbprint of the certificate used to serve the OpenID configuration. If the provider uses multiple certificates to serve content, it is required to manually configure these.
        """
        return pulumi.get(self, "thumbprints")

    @property
    @pulumi.getter
    def url(self) -> pulumi.Output[str]:
        """
        The OIDC issuer URL.
        """
        return pulumi.get(self, "url")

