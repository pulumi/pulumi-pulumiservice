// *** WARNING: this file was generated by pulumigen. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;

namespace Pulumi.PulumiService.Inputs
{

    public sealed class AzureOIDCConfigurationArgs : global::Pulumi.ResourceArgs
    {
        /// <summary>
        /// The client ID of the federated workload identity.
        /// </summary>
        [Input("clientId")]
        public Input<string>? ClientId { get; set; }

        /// <summary>
        /// The subscription ID of the federated workload identity.
        /// </summary>
        [Input("subscriptionID")]
        public Input<string>? SubscriptionID { get; set; }

        /// <summary>
        /// The tenant ID of the federated workload identity.
        /// </summary>
        [Input("tenantId")]
        public Input<string>? TenantId { get; set; }

        public AzureOIDCConfigurationArgs()
        {
        }
        public static new AzureOIDCConfigurationArgs Empty => new AzureOIDCConfigurationArgs();
    }
}
