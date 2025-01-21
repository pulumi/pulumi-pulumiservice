// *** WARNING: this file was generated by pulumi. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;

namespace Pulumi.PulumiService
{
    /// <summary>
    /// An OIDC issuer registration with Pulumi.
    /// </summary>
    [PulumiServiceResourceType("pulumiservice:index:OidcIssuer")]
    public partial class OidcIssuer : global::Pulumi.CustomResource
    {
        /// <summary>
        /// The maximum duration of the Pulumi access token working after an exchange, specified in seconds.
        /// </summary>
        [Output("maxExpirationSeconds")]
        public Output<int?> MaxExpirationSeconds { get; private set; } = null!;

        /// <summary>
        /// Issuer name.
        /// </summary>
        [Output("name")]
        public Output<string> Name { get; private set; } = null!;

        /// <summary>
        /// Organization name.
        /// </summary>
        [Output("organization")]
        public Output<string> Organization { get; private set; } = null!;

        /// <summary>
        /// The authorization policies for this Oidc Issuer.
        /// </summary>
        [Output("policies")]
        public Output<ImmutableArray<Outputs.AuthPolicyDefinition>> Policies { get; private set; } = null!;

        /// <summary>
        /// The thumbprints of issuer's TLS certificates. By default, Pulumi will store the thumbprint of the certificate used to serve the OpenID configuration. If the provider uses multiple certificates to serve content, it is required to manually configure these.
        /// </summary>
        [Output("thumbprints")]
        public Output<ImmutableArray<string>> Thumbprints { get; private set; } = null!;

        /// <summary>
        /// The OIDC issuer URL.
        /// </summary>
        [Output("url")]
        public Output<string> Url { get; private set; } = null!;


        /// <summary>
        /// Create a OidcIssuer resource with the given unique name, arguments, and options.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resource</param>
        /// <param name="args">The arguments used to populate this resource's properties</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public OidcIssuer(string name, OidcIssuerArgs args, CustomResourceOptions? options = null)
            : base("pulumiservice:index:OidcIssuer", name, args ?? new OidcIssuerArgs(), MakeResourceOptions(options, ""))
        {
        }

        private OidcIssuer(string name, Input<string> id, CustomResourceOptions? options = null)
            : base("pulumiservice:index:OidcIssuer", name, null, MakeResourceOptions(options, id))
        {
        }

        private static CustomResourceOptions MakeResourceOptions(CustomResourceOptions? options, Input<string>? id)
        {
            var defaultOptions = new CustomResourceOptions
            {
                Version = Utilities.Version,
                ReplaceOnChanges =
                {
                    "organization",
                    "url",
                },
            };
            var merged = CustomResourceOptions.Merge(defaultOptions, options);
            // Override the ID if one was specified for consistency with other language SDKs.
            merged.Id = id ?? merged.Id;
            return merged;
        }
        /// <summary>
        /// Get an existing OidcIssuer resource's state with the given name, ID, and optional extra
        /// properties used to qualify the lookup.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resulting resource.</param>
        /// <param name="id">The unique provider ID of the resource to lookup.</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public static OidcIssuer Get(string name, Input<string> id, CustomResourceOptions? options = null)
        {
            return new OidcIssuer(name, id, options);
        }
    }

    public sealed class OidcIssuerArgs : global::Pulumi.ResourceArgs
    {
        /// <summary>
        /// The maximum duration of the Pulumi access token working after an exchange, specified in seconds.
        /// </summary>
        [Input("maxExpirationSeconds")]
        public Input<int>? MaxExpirationSeconds { get; set; }

        /// <summary>
        /// Issuer name.
        /// </summary>
        [Input("name", required: true)]
        public Input<string> Name { get; set; } = null!;

        /// <summary>
        /// Organization name.
        /// </summary>
        [Input("organization", required: true)]
        public Input<string> Organization { get; set; } = null!;

        [Input("policies")]
        private InputList<Inputs.AuthPolicyDefinitionArgs>? _policies;

        /// <summary>
        /// The auth policies for this Oidc Issuer.
        /// </summary>
        public InputList<Inputs.AuthPolicyDefinitionArgs> Policies
        {
            get => _policies ?? (_policies = new InputList<Inputs.AuthPolicyDefinitionArgs>());
            set => _policies = value;
        }

        [Input("thumbprints")]
        private InputList<string>? _thumbprints;

        /// <summary>
        /// The thumbprints of issuer's TLS certificates. By default, Pulumi will store the thumbprint of the certificate used to serve the OpenID configuration. If the provider uses multiple certificates to serve content, it is required to manually configure these.
        /// </summary>
        public InputList<string> Thumbprints
        {
            get => _thumbprints ?? (_thumbprints = new InputList<string>());
            set => _thumbprints = value;
        }

        /// <summary>
        /// The OIDC issuer URL.
        /// </summary>
        [Input("url", required: true)]
        public Input<string> Url { get; set; } = null!;

        public OidcIssuerArgs()
        {
        }
        public static new OidcIssuerArgs Empty => new OidcIssuerArgs();
    }
}
