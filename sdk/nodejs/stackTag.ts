// *** WARNING: this file was generated by pulumi-language-nodejs. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as utilities from "./utilities";

/**
 * Stacks have associated metadata in the form of tags. Each tag consists of a name and value.
 */
export class StackTag extends pulumi.CustomResource {
    /**
     * Get an existing StackTag resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param opts Optional settings to control the behavior of the CustomResource.
     */
    public static get(name: string, id: pulumi.Input<pulumi.ID>, opts?: pulumi.CustomResourceOptions): StackTag {
        return new StackTag(name, undefined as any, { ...opts, id: id });
    }

    /** @internal */
    public static readonly __pulumiType = 'pulumiservice:index:StackTag';

    /**
     * Returns true if the given object is an instance of StackTag.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    public static isInstance(obj: any): obj is StackTag {
        if (obj === undefined || obj === null) {
            return false;
        }
        return obj['__pulumiType'] === StackTag.__pulumiType;
    }

    /**
     * Name of the tag. The 'key' part of the key=value pair
     */
    public readonly name!: pulumi.Output<string>;
    /**
     * Organization name.
     */
    public readonly organization!: pulumi.Output<string>;
    /**
     * Project name.
     */
    public readonly project!: pulumi.Output<string>;
    /**
     * Stack name.
     */
    public readonly stack!: pulumi.Output<string>;
    /**
     * Value of the tag. The 'value' part of the key=value pair
     */
    public readonly value!: pulumi.Output<string>;

    /**
     * Create a StackTag resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: StackTagArgs, opts?: pulumi.CustomResourceOptions) {
        let resourceInputs: pulumi.Inputs = {};
        opts = opts || {};
        if (!opts.id) {
            if ((!args || args.name === undefined) && !opts.urn) {
                throw new Error("Missing required property 'name'");
            }
            if ((!args || args.organization === undefined) && !opts.urn) {
                throw new Error("Missing required property 'organization'");
            }
            if ((!args || args.project === undefined) && !opts.urn) {
                throw new Error("Missing required property 'project'");
            }
            if ((!args || args.stack === undefined) && !opts.urn) {
                throw new Error("Missing required property 'stack'");
            }
            if ((!args || args.value === undefined) && !opts.urn) {
                throw new Error("Missing required property 'value'");
            }
            resourceInputs["name"] = args ? args.name : undefined;
            resourceInputs["organization"] = args ? args.organization : undefined;
            resourceInputs["project"] = args ? args.project : undefined;
            resourceInputs["stack"] = args ? args.stack : undefined;
            resourceInputs["value"] = args ? args.value : undefined;
        } else {
            resourceInputs["name"] = undefined /*out*/;
            resourceInputs["organization"] = undefined /*out*/;
            resourceInputs["project"] = undefined /*out*/;
            resourceInputs["stack"] = undefined /*out*/;
            resourceInputs["value"] = undefined /*out*/;
        }
        opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts);
        super(StackTag.__pulumiType, name, resourceInputs, opts);
    }
}

/**
 * The set of arguments for constructing a StackTag resource.
 */
export interface StackTagArgs {
    /**
     * Name of the tag. The 'key' part of the key=value pair
     */
    name: pulumi.Input<string>;
    /**
     * Organization name.
     */
    organization: pulumi.Input<string>;
    /**
     * Project name.
     */
    project: pulumi.Input<string>;
    /**
     * Stack name.
     */
    stack: pulumi.Input<string>;
    /**
     * Value of the tag. The 'value' part of the key=value pair
     */
    value: pulumi.Input<string>;
}
