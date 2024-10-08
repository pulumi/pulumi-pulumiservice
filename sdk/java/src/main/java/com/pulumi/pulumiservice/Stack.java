// *** WARNING: this file was generated by pulumi. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package com.pulumi.pulumiservice;

import com.pulumi.core.Output;
import com.pulumi.core.annotations.Export;
import com.pulumi.core.annotations.ResourceType;
import com.pulumi.core.internal.Codegen;
import com.pulumi.pulumiservice.StackArgs;
import com.pulumi.pulumiservice.Utilities;
import java.lang.Boolean;
import java.lang.String;
import java.util.Optional;
import javax.annotation.Nullable;

/**
 * A stack is a collection of resources that share a common lifecycle. Stacks are uniquely identified by their name and the project they belong to.
 * 
 */
@ResourceType(type="pulumiservice:index:Stack")
public class Stack extends com.pulumi.resources.CustomResource {
    /**
     * Optional. Flag indicating whether to delete the stack even if it still contains resources.
     * 
     */
    @Export(name="forceDestroy", refs={Boolean.class}, tree="[0]")
    private Output</* @Nullable */ Boolean> forceDestroy;

    /**
     * @return Optional. Flag indicating whether to delete the stack even if it still contains resources.
     * 
     */
    public Output<Optional<Boolean>> forceDestroy() {
        return Codegen.optional(this.forceDestroy);
    }
    /**
     * The name of the organization.
     * 
     */
    @Export(name="organizationName", refs={String.class}, tree="[0]")
    private Output<String> organizationName;

    /**
     * @return The name of the organization.
     * 
     */
    public Output<String> organizationName() {
        return this.organizationName;
    }
    /**
     * The name of the project.
     * 
     */
    @Export(name="projectName", refs={String.class}, tree="[0]")
    private Output<String> projectName;

    /**
     * @return The name of the project.
     * 
     */
    public Output<String> projectName() {
        return this.projectName;
    }
    /**
     * The name of the stack.
     * 
     */
    @Export(name="stackName", refs={String.class}, tree="[0]")
    private Output<String> stackName;

    /**
     * @return The name of the stack.
     * 
     */
    public Output<String> stackName() {
        return this.stackName;
    }

    /**
     *
     * @param name The _unique_ name of the resulting resource.
     */
    public Stack(java.lang.String name) {
        this(name, StackArgs.Empty);
    }
    /**
     *
     * @param name The _unique_ name of the resulting resource.
     * @param args The arguments to use to populate this resource's properties.
     */
    public Stack(java.lang.String name, StackArgs args) {
        this(name, args, null);
    }
    /**
     *
     * @param name The _unique_ name of the resulting resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param options A bag of options that control this resource's behavior.
     */
    public Stack(java.lang.String name, StackArgs args, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        super("pulumiservice:index:Stack", name, makeArgs(args, options), makeResourceOptions(options, Codegen.empty()), false);
    }

    private Stack(java.lang.String name, Output<java.lang.String> id, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        super("pulumiservice:index:Stack", name, null, makeResourceOptions(options, id), false);
    }

    private static StackArgs makeArgs(StackArgs args, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        if (options != null && options.getUrn().isPresent()) {
            return null;
        }
        return args == null ? StackArgs.Empty : args;
    }

    private static com.pulumi.resources.CustomResourceOptions makeResourceOptions(@Nullable com.pulumi.resources.CustomResourceOptions options, @Nullable Output<java.lang.String> id) {
        var defaultOptions = com.pulumi.resources.CustomResourceOptions.builder()
            .version(Utilities.getVersion())
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
    public static Stack get(java.lang.String name, Output<java.lang.String> id, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        return new Stack(name, id, options);
    }
}
