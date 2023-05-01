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
    public sealed class DeploymentSettingsGitAuthBasicAuth
    {
        /// <summary>
        /// Password for git basic authentication.
        /// </summary>
        public readonly string Password;
        /// <summary>
        /// User name for git basic authentication.
        /// </summary>
        public readonly string Username;

        [OutputConstructor]
        private DeploymentSettingsGitAuthBasicAuth(
            string password,

            string username)
        {
            Password = password;
            Username = username;
        }
    }
}
