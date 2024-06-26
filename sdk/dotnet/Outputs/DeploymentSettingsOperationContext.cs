// *** WARNING: this file was generated by pulumi. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;

namespace Pulumi.PulumiService.Outputs
{

    /// <summary>
    /// Settings related to the Pulumi operation environment during the deployment.
    /// </summary>
    [OutputType]
    public sealed class DeploymentSettingsOperationContext
    {
        /// <summary>
        /// Environment variables to set for the deployment.
        /// </summary>
        public readonly ImmutableDictionary<string, string>? EnvironmentVariables;
        /// <summary>
        /// OIDC configuration to use during the deployment.
        /// </summary>
        public readonly Outputs.OperationContextOIDC? Oidc;
        /// <summary>
        /// Options to override default behavior during the deployment.
        /// </summary>
        public readonly Outputs.OperationContextOptions? Options;
        /// <summary>
        /// Shell commands to run before the Pulumi operation executes.
        /// </summary>
        public readonly ImmutableArray<string> PreRunCommands;

        [OutputConstructor]
        private DeploymentSettingsOperationContext(
            ImmutableDictionary<string, string>? environmentVariables,

            Outputs.OperationContextOIDC? oidc,

            Outputs.OperationContextOptions? options,

            ImmutableArray<string> preRunCommands)
        {
            EnvironmentVariables = environmentVariables;
            Oidc = oidc;
            Options = options;
            PreRunCommands = preRunCommands;
        }
    }
}
