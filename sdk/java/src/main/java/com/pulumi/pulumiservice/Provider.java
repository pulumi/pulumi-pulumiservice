// *** WARNING: this file was generated by pulumi. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package com.pulumi.pulumiservice;

import com.pulumi.core.Output;
import com.pulumi.core.annotations.Export;
import com.pulumi.core.annotations.ResourceType;
import com.pulumi.core.internal.Codegen;
import com.pulumi.pulumiservice.ProviderArgs;
import com.pulumi.pulumiservice.Utilities;
import java.lang.String;
import java.util.List;
import java.util.Optional;
import javax.annotation.Nullable;

@ResourceType(type="pulumi:providers:pulumiservice")
public class Provider extends com.pulumi.resources.ProviderResource {
    /**
     * Access Token to authenticate with Pulumi Cloud.
     * 
     */
    @Export(name="accessToken", refs={String.class}, tree="[0]")
    private Output</* @Nullable */ String> accessToken;

    /**
     * @return Access Token to authenticate with Pulumi Cloud.
     * 
     */
    public Output<Optional<String>> accessToken() {
        return Codegen.optional(this.accessToken);
    }
    /**
     * The service URL used to reach Pulumi Cloud.
     * 
     */
    @Export(name="serviceURL", refs={String.class}, tree="[0]")
    private Output</* @Nullable */ String> serviceURL;

    /**
     * @return The service URL used to reach Pulumi Cloud.
     * 
     */
    public Output<Optional<String>> serviceURL() {
        return Codegen.optional(this.serviceURL);
    }

    /**
     *
     * @param name The _unique_ name of the resulting resource.
     */
    public Provider(String name) {
        this(name, ProviderArgs.Empty);
    }
    /**
     *
     * @param name The _unique_ name of the resulting resource.
     * @param args The arguments to use to populate this resource's properties.
     */
    public Provider(String name, @Nullable ProviderArgs args) {
        this(name, args, null);
    }
    /**
     *
     * @param name The _unique_ name of the resulting resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param options A bag of options that control this resource's behavior.
     */
    public Provider(String name, @Nullable ProviderArgs args, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        super("pulumiservice", name, args == null ? ProviderArgs.Empty : args, makeResourceOptions(options, Codegen.empty()));
    }

    private static com.pulumi.resources.CustomResourceOptions makeResourceOptions(@Nullable com.pulumi.resources.CustomResourceOptions options, @Nullable Output<String> id) {
        var defaultOptions = com.pulumi.resources.CustomResourceOptions.builder()
            .version(Utilities.getVersion())
            .additionalSecretOutputs(List.of(
                "accessToken"
            ))
            .build();
        return com.pulumi.resources.CustomResourceOptions.merge(defaultOptions, options, id);
    }

}
