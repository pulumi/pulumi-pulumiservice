// *** WARNING: this file was generated by pulumi-language-nodejs. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as inputs from "./types/input";
import * as outputs from "./types/output";
import * as enums from "./types/enums";
import * as utilities from "./utilities";

/**
 * An OIDC issuer registration with Pulumi.
 */
export class OidcIssuer extends pulumi.CustomResource {
    /**
     * Get an existing OidcIssuer resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param opts Optional settings to control the behavior of the CustomResource.
     */
    public static get(name: string, id: pulumi.Input<pulumi.ID>, opts?: pulumi.CustomResourceOptions): OidcIssuer {
        return new OidcIssuer(name, undefined as any, { ...opts, id: id });
    }

    /** @internal */
    public static readonly __pulumiType = 'pulumiservice:index:OidcIssuer';

    /**
     * Returns true if the given object is an instance of OidcIssuer.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    public static isInstance(obj: any): obj is OidcIssuer {
        if (obj === undefined || obj === null) {
            return false;
        }
        return obj['__pulumiType'] === OidcIssuer.__pulumiType;
    }

    /**
     * The maximum duration of the Pulumi access token working after an exchange, specified in seconds.
     */
    public readonly maxExpirationSeconds!: pulumi.Output<number | undefined>;
    /**
     * Issuer name.
     */
    public readonly name!: pulumi.Output<string>;
    /**
     * Organization name.
     */
    public readonly organization!: pulumi.Output<string>;
    /**
     * The authorization policies for this Oidc Issuer.
     */
    public readonly policies!: pulumi.Output<outputs.AuthPolicyDefinition[]>;
    /**
     * The thumbprints of issuer's TLS certificates. By default, Pulumi will store the thumbprint of the certificate used to serve the OpenID configuration. If the provider uses multiple certificates to serve content, it is required to manually configure these.
     */
    public readonly thumbprints!: pulumi.Output<string[]>;
    /**
     * The OIDC issuer URL.
     */
    public readonly url!: pulumi.Output<string>;

    /**
     * Create a OidcIssuer resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: OidcIssuerArgs, opts?: pulumi.CustomResourceOptions) {
        let resourceInputs: pulumi.Inputs = {};
        opts = opts || {};
        if (!opts.id) {
            if ((!args || args.name === undefined) && !opts.urn) {
                throw new Error("Missing required property 'name'");
            }
            if ((!args || args.organization === undefined) && !opts.urn) {
                throw new Error("Missing required property 'organization'");
            }
            if ((!args || args.url === undefined) && !opts.urn) {
                throw new Error("Missing required property 'url'");
            }
            resourceInputs["maxExpirationSeconds"] = args ? args.maxExpirationSeconds : undefined;
            resourceInputs["name"] = args ? args.name : undefined;
            resourceInputs["organization"] = args ? args.organization : undefined;
            resourceInputs["policies"] = args ? args.policies : undefined;
            resourceInputs["thumbprints"] = args ? args.thumbprints : undefined;
            resourceInputs["url"] = args ? args.url : undefined;
        } else {
            resourceInputs["maxExpirationSeconds"] = undefined /*out*/;
            resourceInputs["name"] = undefined /*out*/;
            resourceInputs["organization"] = undefined /*out*/;
            resourceInputs["policies"] = undefined /*out*/;
            resourceInputs["thumbprints"] = undefined /*out*/;
            resourceInputs["url"] = undefined /*out*/;
        }
        opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts);
        super(OidcIssuer.__pulumiType, name, resourceInputs, opts);
    }
}

/**
 * The set of arguments for constructing a OidcIssuer resource.
 */
export interface OidcIssuerArgs {
    /**
     * The maximum duration of the Pulumi access token working after an exchange, specified in seconds.
     */
    maxExpirationSeconds?: pulumi.Input<number>;
    /**
     * Issuer name.
     */
    name: pulumi.Input<string>;
    /**
     * Organization name.
     */
    organization: pulumi.Input<string>;
    /**
     * The auth policies for this Oidc Issuer.
     */
    policies?: pulumi.Input<pulumi.Input<inputs.AuthPolicyDefinitionArgs>[]>;
    /**
     * The thumbprints of issuer's TLS certificates. By default, Pulumi will store the thumbprint of the certificate used to serve the OpenID configuration. If the provider uses multiple certificates to serve content, it is required to manually configure these.
     */
    thumbprints?: pulumi.Input<pulumi.Input<string>[]>;
    /**
     * The OIDC issuer URL.
     */
    url: pulumi.Input<string>;
}
