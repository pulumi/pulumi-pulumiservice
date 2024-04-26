// *** WARNING: this file was generated by pulumi-language-nodejs. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as utilities from "./utilities";

export class Provider extends pulumi.ProviderResource {
    /** @internal */
    public static readonly __pulumiType = 'pulumiservice';

    /**
     * Returns true if the given object is an instance of Provider.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    public static isInstance(obj: any): obj is Provider {
        if (obj === undefined || obj === null) {
            return false;
        }
        return obj['__pulumiType'] === "pulumi:providers:" + Provider.__pulumiType;
    }

    /**
     * Access Token to authenticate with Pulumi Cloud.
     */
    public readonly accessToken!: pulumi.Output<string | undefined>;
    /**
     * The service URL used to reach Pulumi Cloud.
     */
    public readonly serviceURL!: pulumi.Output<string | undefined>;

    /**
     * Create a Provider resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args?: ProviderArgs, opts?: pulumi.ResourceOptions) {
        let resourceInputs: pulumi.Inputs = {};
        opts = opts || {};
        {
            resourceInputs["accessToken"] = (args?.accessToken ? pulumi.secret(args.accessToken) : undefined) ?? utilities.getEnv("PULUMI_ACCESS_TOKEN");
            resourceInputs["serviceURL"] = (args ? args.serviceURL : undefined) ?? (utilities.getEnv("PULUMI_BACKEND_URL") || "https://api.pulumi.com");
        }
        opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts);
        const secretOpts = { additionalSecretOutputs: ["accessToken"] };
        opts = pulumi.mergeOptions(opts, secretOpts);
        super(Provider.__pulumiType, name, resourceInputs, opts);
    }
}

/**
 * The set of arguments for constructing a Provider resource.
 */
export interface ProviderArgs {
    /**
     * Access Token to authenticate with Pulumi Cloud.
     */
    accessToken?: pulumi.Input<string>;
    /**
     * The service URL used to reach Pulumi Cloud.
     */
    serviceURL?: pulumi.Input<string>;
}
