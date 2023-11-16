// *** WARNING: this file was generated by pulumigen. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as utilities from "./utilities";

/**
 * The Pulumi Cloud allows users to create access tokens scoped to team. Team access tokens is a resource to create them and assign them to a team
 */
export class TeamAccessToken extends pulumi.CustomResource {
    /**
     * Get an existing TeamAccessToken resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param opts Optional settings to control the behavior of the CustomResource.
     */
    public static get(name: string, id: pulumi.Input<pulumi.ID>, opts?: pulumi.CustomResourceOptions): TeamAccessToken {
        return new TeamAccessToken(name, undefined as any, { ...opts, id: id });
    }

    /** @internal */
    public static readonly __pulumiType = 'pulumiservice:index:TeamAccessToken';

    /**
     * Returns true if the given object is an instance of TeamAccessToken.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    public static isInstance(obj: any): obj is TeamAccessToken {
        if (obj === undefined || obj === null) {
            return false;
        }
        return obj['__pulumiType'] === TeamAccessToken.__pulumiType;
    }

    /**
     * Optional. Description for the token.
     */
    public readonly description!: pulumi.Output<string | undefined>;
    /**
     * The name for the token. This must be unique amongst all machine tokens within your organization.
     */
    public readonly name!: pulumi.Output<string>;
    /**
     * The organization's name.
     */
    public readonly organizationName!: pulumi.Output<string>;
    /**
     * The team name.
     */
    public readonly teamName!: pulumi.Output<string>;
    /**
     * The token's value.
     */
    public /*out*/ readonly value!: pulumi.Output<string>;

    /**
     * Create a TeamAccessToken resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: TeamAccessTokenArgs, opts?: pulumi.CustomResourceOptions) {
        let resourceInputs: pulumi.Inputs = {};
        opts = opts || {};
        if (!opts.id) {
            if ((!args || args.name === undefined) && !opts.urn) {
                throw new Error("Missing required property 'name'");
            }
            if ((!args || args.organizationName === undefined) && !opts.urn) {
                throw new Error("Missing required property 'organizationName'");
            }
            if ((!args || args.teamName === undefined) && !opts.urn) {
                throw new Error("Missing required property 'teamName'");
            }
            resourceInputs["description"] = args ? args.description : undefined;
            resourceInputs["name"] = args ? args.name : undefined;
            resourceInputs["organizationName"] = args ? args.organizationName : undefined;
            resourceInputs["teamName"] = args ? args.teamName : undefined;
            resourceInputs["value"] = undefined /*out*/;
        } else {
            resourceInputs["description"] = undefined /*out*/;
            resourceInputs["name"] = undefined /*out*/;
            resourceInputs["organizationName"] = undefined /*out*/;
            resourceInputs["teamName"] = undefined /*out*/;
            resourceInputs["value"] = undefined /*out*/;
        }
        opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts);
        const secretOpts = { additionalSecretOutputs: ["value"] };
        opts = pulumi.mergeOptions(opts, secretOpts);
        super(TeamAccessToken.__pulumiType, name, resourceInputs, opts);
    }
}

/**
 * The set of arguments for constructing a TeamAccessToken resource.
 */
export interface TeamAccessTokenArgs {
    /**
     * Optional. Team description.
     */
    description?: pulumi.Input<string>;
    /**
     * The name for the token. This must be unique amongst all machine tokens within your organization.
     */
    name: pulumi.Input<string>;
    /**
     * The organization's name.
     */
    organizationName: pulumi.Input<string>;
    /**
     * The team name.
     */
    teamName: pulumi.Input<string>;
}
