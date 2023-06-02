// *** WARNING: this file was generated by pulumigen. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;

namespace Pulumi.PulumiService
{
    /// <summary>
    /// Pulumi Webhooks allow you to notify external services of events happening within your Pulumi organization or stack. For example, you can trigger a notification whenever a stack is updated. Whenever an event occurs, Pulumi will send an HTTP POST request to all registered webhooks. The webhook can then be used to emit some notification, start running integration tests, or even update additional stacks.
    /// 
    /// ### Import
    /// 
    /// Pulumi webhooks can be imported using the `id`, which for webhooks is `{org}/{project}/{stack}/{webhook-name}` e.g.,
    /// 
    /// ```sh
    ///  $ pulumi import pulumiservice:index:Webhook my_webhook my-org/my-project/my-stack/4b0d0671
    /// ```
    /// </summary>
    [PulumiServiceResourceType("pulumiservice:index:Webhook")]
    public partial class Webhook : global::Pulumi.CustomResource
    {
        /// <summary>
        /// Indicates whether this webhook is enabled or not.
        /// </summary>
        [Output("active")]
        public Output<bool?> Active { get; private set; } = null!;

        /// <summary>
        /// The friendly name displayed in the Pulumi Cloud.
        /// </summary>
        [Output("displayName")]
        public Output<string?> DisplayName { get; private set; } = null!;

        /// <summary>
        /// Optional set of filters to apply to the webhook. See [webhook docs](https://www.pulumi.com/docs/intro/pulumi-service/webhooks/#filters) for more information.
        /// </summary>
        [Output("filters")]
        public Output<ImmutableArray<Pulumi.PulumiService.WebhookFilters>> Filters { get; private set; } = null!;

        /// <summary>
        /// Format of the webhook payload. Can be either `raw` or `slack`. Defaults to `raw`.
        /// </summary>
        [Output("format")]
        public Output<Pulumi.PulumiService.WebhookFormat?> Format { get; private set; } = null!;

        /// <summary>
        /// Webhook identifier generated by Pulumi Cloud.
        /// </summary>
        [Output("name")]
        public Output<string?> Name { get; private set; } = null!;

        /// <summary>
        /// Name of the organization.
        /// </summary>
        [Output("organizationName")]
        public Output<string?> OrganizationName { get; private set; } = null!;

        /// <summary>
        /// URL to send request to.
        /// </summary>
        [Output("payloadUrl")]
        public Output<string?> PayloadUrl { get; private set; } = null!;

        /// <summary>
        /// Name of the project. Only specified if this is a stack webhook.
        /// </summary>
        [Output("projectName")]
        public Output<string?> ProjectName { get; private set; } = null!;

        /// <summary>
        /// Optional. secret used as the HMAC key. See [webhook docs](https://www.pulumi.com/docs/intro/pulumi-service/webhooks/#headers) for more information.
        /// </summary>
        [Output("secret")]
        public Output<string?> Secret { get; private set; } = null!;

        /// <summary>
        /// Name of the stack. Only specified if this is a stack webhook.
        /// </summary>
        [Output("stackName")]
        public Output<string?> StackName { get; private set; } = null!;


        /// <summary>
        /// Create a Webhook resource with the given unique name, arguments, and options.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resource</param>
        /// <param name="args">The arguments used to populate this resource's properties</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public Webhook(string name, WebhookArgs args, CustomResourceOptions? options = null)
            : base("pulumiservice:index:Webhook", name, args ?? new WebhookArgs(), MakeResourceOptions(options, ""))
        {
        }

        private Webhook(string name, Input<string> id, CustomResourceOptions? options = null)
            : base("pulumiservice:index:Webhook", name, null, MakeResourceOptions(options, id))
        {
        }

        private static CustomResourceOptions MakeResourceOptions(CustomResourceOptions? options, Input<string>? id)
        {
            var defaultOptions = new CustomResourceOptions
            {
                Version = Utilities.Version,
                AdditionalSecretOutputs =
                {
                    "secret",
                },
            };
            var merged = CustomResourceOptions.Merge(defaultOptions, options);
            // Override the ID if one was specified for consistency with other language SDKs.
            merged.Id = id ?? merged.Id;
            return merged;
        }
        /// <summary>
        /// Get an existing Webhook resource's state with the given name, ID, and optional extra
        /// properties used to qualify the lookup.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resulting resource.</param>
        /// <param name="id">The unique provider ID of the resource to lookup.</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public static Webhook Get(string name, Input<string> id, CustomResourceOptions? options = null)
        {
            return new Webhook(name, id, options);
        }
    }

    public sealed class WebhookArgs : global::Pulumi.ResourceArgs
    {
        /// <summary>
        /// Indicates whether this webhook is enabled or not.
        /// </summary>
        [Input("active", required: true)]
        public Input<bool> Active { get; set; } = null!;

        /// <summary>
        /// The friendly name displayed in the Pulumi Cloud.
        /// </summary>
        [Input("displayName", required: true)]
        public Input<string> DisplayName { get; set; } = null!;

        [Input("filters")]
        private InputList<Pulumi.PulumiService.WebhookFilters>? _filters;

        /// <summary>
        /// Optional set of filters to apply to the webhook. See [webhook docs](https://www.pulumi.com/docs/intro/pulumi-service/webhooks/#filters) for more information.
        /// </summary>
        public InputList<Pulumi.PulumiService.WebhookFilters> Filters
        {
            get => _filters ?? (_filters = new InputList<Pulumi.PulumiService.WebhookFilters>());
            set => _filters = value;
        }

        /// <summary>
        /// Format of the webhook payload. Can be either `raw` or `slack`. Defaults to `raw`.
        /// </summary>
        [Input("format")]
        public Input<Pulumi.PulumiService.WebhookFormat>? Format { get; set; }

        /// <summary>
        /// Name of the organization.
        /// </summary>
        [Input("organizationName", required: true)]
        public Input<string> OrganizationName { get; set; } = null!;

        /// <summary>
        /// URL to send request to.
        /// </summary>
        [Input("payloadUrl", required: true)]
        public Input<string> PayloadUrl { get; set; } = null!;

        /// <summary>
        /// Name of the project. Only needed if this is a stack webhook.
        /// </summary>
        [Input("projectName")]
        public Input<string>? ProjectName { get; set; }

        [Input("secret")]
        private Input<string>? _secret;

        /// <summary>
        /// Optional. secret used as the HMAC key. See [webhook docs](https://www.pulumi.com/docs/intro/pulumi-service/webhooks/#headers) for more information.
        /// </summary>
        public Input<string>? Secret
        {
            get => _secret;
            set
            {
                var emptySecret = Output.CreateSecret(0);
                _secret = Output.Tuple<Input<string>?, int>(value, emptySecret).Apply(t => t.Item1);
            }
        }

        /// <summary>
        /// Name of the stack. Only needed if this is a stack webhook.
        /// </summary>
        [Input("stackName")]
        public Input<string>? StackName { get; set; }

        public WebhookArgs()
        {
            Format = Pulumi.PulumiService.WebhookFormat.Raw;
        }
        public static new WebhookArgs Empty => new WebhookArgs();
    }
}
