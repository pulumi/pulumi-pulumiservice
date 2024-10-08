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
    /// A source for Pulumi templates
    /// </summary>
    [PulumiServiceResourceType("pulumiservice:index:TemplateSource")]
    public partial class TemplateSource : global::Pulumi.CustomResource
    {
        /// <summary>
        /// The default destination for projects using templates from this source.
        /// </summary>
        [Output("destination")]
        public Output<Outputs.TemplateSourceDestination?> Destination { get; private set; } = null!;

        /// <summary>
        /// Organization name.
        /// </summary>
        [Output("organizationName")]
        public Output<string> OrganizationName { get; private set; } = null!;

        /// <summary>
        /// Source name.
        /// </summary>
        [Output("sourceName")]
        public Output<string> SourceName { get; private set; } = null!;

        /// <summary>
        /// Github URL of the repository from which to grab templates.
        /// </summary>
        [Output("sourceURL")]
        public Output<string> SourceURL { get; private set; } = null!;


        /// <summary>
        /// Create a TemplateSource resource with the given unique name, arguments, and options.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resource</param>
        /// <param name="args">The arguments used to populate this resource's properties</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public TemplateSource(string name, TemplateSourceArgs args, CustomResourceOptions? options = null)
            : base("pulumiservice:index:TemplateSource", name, args ?? new TemplateSourceArgs(), MakeResourceOptions(options, ""))
        {
        }

        private TemplateSource(string name, Input<string> id, CustomResourceOptions? options = null)
            : base("pulumiservice:index:TemplateSource", name, null, MakeResourceOptions(options, id))
        {
        }

        private static CustomResourceOptions MakeResourceOptions(CustomResourceOptions? options, Input<string>? id)
        {
            var defaultOptions = new CustomResourceOptions
            {
                Version = Utilities.Version,
            };
            var merged = CustomResourceOptions.Merge(defaultOptions, options);
            // Override the ID if one was specified for consistency with other language SDKs.
            merged.Id = id ?? merged.Id;
            return merged;
        }
        /// <summary>
        /// Get an existing TemplateSource resource's state with the given name, ID, and optional extra
        /// properties used to qualify the lookup.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resulting resource.</param>
        /// <param name="id">The unique provider ID of the resource to lookup.</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public static TemplateSource Get(string name, Input<string> id, CustomResourceOptions? options = null)
        {
            return new TemplateSource(name, id, options);
        }
    }

    public sealed class TemplateSourceArgs : global::Pulumi.ResourceArgs
    {
        /// <summary>
        /// The default destination for projects using templates from this source.
        /// </summary>
        [Input("destination")]
        public Input<Inputs.TemplateSourceDestinationArgs>? Destination { get; set; }

        /// <summary>
        /// Organization name.
        /// </summary>
        [Input("organizationName", required: true)]
        public Input<string> OrganizationName { get; set; } = null!;

        /// <summary>
        /// Source name.
        /// </summary>
        [Input("sourceName", required: true)]
        public Input<string> SourceName { get; set; } = null!;

        /// <summary>
        /// Github URL of the repository from which to grab templates.
        /// </summary>
        [Input("sourceURL", required: true)]
        public Input<string> SourceURL { get; set; } = null!;

        public TemplateSourceArgs()
        {
        }
        public static new TemplateSourceArgs Empty => new TemplateSourceArgs();
    }
}
