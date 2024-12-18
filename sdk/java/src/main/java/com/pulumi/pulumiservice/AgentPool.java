// *** WARNING: this file was generated by pulumi. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package com.pulumi.pulumiservice;

import com.pulumi.core.Output;
import com.pulumi.core.annotations.Export;
import com.pulumi.core.annotations.ResourceType;
import com.pulumi.core.internal.Codegen;
import com.pulumi.pulumiservice.AgentPoolArgs;
import com.pulumi.pulumiservice.Utilities;
import java.lang.Boolean;
import java.lang.String;
import java.util.List;
import java.util.Optional;
import javax.annotation.Nullable;

/**
 * Agent Pool for customer managed deployments
 * 
 */
@ResourceType(type="pulumiservice:index:AgentPool")
public class AgentPool extends com.pulumi.resources.CustomResource {
    /**
     * The agent pool identifier.
     * 
     */
    @Export(name="agentPoolId", refs={String.class}, tree="[0]")
    private Output<String> agentPoolId;

    /**
     * @return The agent pool identifier.
     * 
     */
    public Output<String> agentPoolId() {
        return this.agentPoolId;
    }
    /**
     * Description of the agent pool.
     * 
     */
    @Export(name="description", refs={String.class}, tree="[0]")
    private Output</* @Nullable */ String> description;

    /**
     * @return Description of the agent pool.
     * 
     */
    public Output<Optional<String>> description() {
        return Codegen.optional(this.description);
    }
    /**
     * Optional. Flag indicating whether to delete the agent pool even if stacks are configured to use it.
     * 
     */
    @Export(name="forceDestroy", refs={Boolean.class}, tree="[0]")
    private Output</* @Nullable */ Boolean> forceDestroy;

    /**
     * @return Optional. Flag indicating whether to delete the agent pool even if stacks are configured to use it.
     * 
     */
    public Output<Optional<Boolean>> forceDestroy() {
        return Codegen.optional(this.forceDestroy);
    }
    /**
     * The name of the agent pool.
     * 
     */
    @Export(name="name", refs={String.class}, tree="[0]")
    private Output<String> name;

    /**
     * @return The name of the agent pool.
     * 
     */
    public Output<String> name() {
        return this.name;
    }
    /**
     * The organization&#39;s name.
     * 
     */
    @Export(name="organizationName", refs={String.class}, tree="[0]")
    private Output<String> organizationName;

    /**
     * @return The organization&#39;s name.
     * 
     */
    public Output<String> organizationName() {
        return this.organizationName;
    }
    /**
     * The agent pool&#39;s token&#39;s value.
     * 
     */
    @Export(name="tokenValue", refs={String.class}, tree="[0]")
    private Output<String> tokenValue;

    /**
     * @return The agent pool&#39;s token&#39;s value.
     * 
     */
    public Output<String> tokenValue() {
        return this.tokenValue;
    }

    /**
     *
     * @param name The _unique_ name of the resulting resource.
     */
    public AgentPool(java.lang.String name) {
        this(name, AgentPoolArgs.Empty);
    }
    /**
     *
     * @param name The _unique_ name of the resulting resource.
     * @param args The arguments to use to populate this resource's properties.
     */
    public AgentPool(java.lang.String name, AgentPoolArgs args) {
        this(name, args, null);
    }
    /**
     *
     * @param name The _unique_ name of the resulting resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param options A bag of options that control this resource's behavior.
     */
    public AgentPool(java.lang.String name, AgentPoolArgs args, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        super("pulumiservice:index:AgentPool", name, makeArgs(args, options), makeResourceOptions(options, Codegen.empty()), false);
    }

    private AgentPool(java.lang.String name, Output<java.lang.String> id, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        super("pulumiservice:index:AgentPool", name, null, makeResourceOptions(options, id), false);
    }

    private static AgentPoolArgs makeArgs(AgentPoolArgs args, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        if (options != null && options.getUrn().isPresent()) {
            return null;
        }
        return args == null ? AgentPoolArgs.Empty : args;
    }

    private static com.pulumi.resources.CustomResourceOptions makeResourceOptions(@Nullable com.pulumi.resources.CustomResourceOptions options, @Nullable Output<java.lang.String> id) {
        var defaultOptions = com.pulumi.resources.CustomResourceOptions.builder()
            .version(Utilities.getVersion())
            .additionalSecretOutputs(List.of(
                "tokenValue"
            ))
            .build();
        return com.pulumi.resources.CustomResourceOptions.merge(defaultOptions, options, id);
    }

    /**
     * Get an existing Host resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param options Optional settings to control the behavior of the CustomResource.
     */
    public static AgentPool get(java.lang.String name, Output<java.lang.String> id, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        return new AgentPool(name, id, options);
    }
}
