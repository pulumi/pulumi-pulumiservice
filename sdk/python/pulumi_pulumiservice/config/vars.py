# coding=utf-8
# *** WARNING: this file was generated by pulumi-language-python. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

import copy
import warnings
import pulumi
import pulumi.runtime
from typing import Any, Mapping, Optional, Sequence, Union, overload
from .. import _utilities

import types

__config__ = pulumi.Config('pulumiservice')


class _ExportableConfig(types.ModuleType):
    @property
    def access_token(self) -> Optional[str]:
        """
        Access Token to authenticate with Pulumi Cloud.
        """
        return __config__.get('accessToken') or _utilities.get_env('PULUMI_ACCESS_TOKEN')

    @property
    def service_url(self) -> str:
        """
        The service URL used to reach Pulumi Cloud.
        """
        return __config__.get('serviceURL') or (_utilities.get_env('PULUMI_BACKEND_URL') or 'https://api.pulumi.com')

