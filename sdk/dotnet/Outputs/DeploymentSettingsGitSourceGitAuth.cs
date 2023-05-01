// *** WARNING: this file was generated by pulumigen. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;

namespace Pulumi.PulumiService.Outputs
{

    /// <summary>
    /// Git source settings for a deployment.
    /// </summary>
    [OutputType]
    public sealed class DeploymentSettingsGitSourceGitAuth
    {
        /// <summary>
        /// Basic auth for git authentication. Only one of `personalAccessToken`, `sshAuth`, or `basicAuth` must be defined.
        /// </summary>
        public readonly Outputs.DeploymentSettingsGitAuthBasicAuth? BasicAuth;
        /// <summary>
        /// SSH auth for git authentication. Only one of `personalAccessToken`, `sshAuth`, or `basicAuth` must be defined.
        /// </summary>
        public readonly Outputs.DeploymentSettingsGitAuthSSHAuth? SshAuth;

        [OutputConstructor]
        private DeploymentSettingsGitSourceGitAuth(
            Outputs.DeploymentSettingsGitAuthBasicAuth? basicAuth,

            Outputs.DeploymentSettingsGitAuthSSHAuth? sshAuth)
        {
            BasicAuth = basicAuth;
            SshAuth = sshAuth;
        }
    }
}
