// *** WARNING: this file was generated by pulumi-java-gen. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package com.pulumi.pulumiservice;

import com.pulumi.core.Output;
import com.pulumi.core.annotations.Export;
import com.pulumi.core.annotations.ResourceType;
import com.pulumi.core.internal.Codegen;
import com.pulumi.pulumiservice.TeamAccessTokenArgs;
import com.pulumi.pulumiservice.Utilities;
import java.lang.String;
import java.util.Optional;
import javax.annotation.Nullable;

/**
 * The Pulumi Service allows users to create access tokens scoped to team. Team access tokens is a resource to create them and assign them to a team
 * 
 */
@ResourceType(type="pulumiservice:index:TeamAccessToken")
public class TeamAccessToken extends com.pulumi.resources.CustomResource {
    /**
     * Optional. Description for the token.
     * 
     */
    @Export(name="description", refs={String.class}, tree="[0]")
    private Output</* @Nullable */ String> description;

    /**
     * @return Optional. Description for the token.
     * 
     */
    public Output<Optional<String>> description() {
        return Codegen.optional(this.description);
    }
    /**
     * The name for the token. This must be unique amongst all machine tokens within your organization.
     * 
     */
    @Export(name="name", refs={String.class}, tree="[0]")
    private Output</* @Nullable */ String> name;

    /**
     * @return The name for the token. This must be unique amongst all machine tokens within your organization.
     * 
     */
    public Output<Optional<String>> name() {
        return Codegen.optional(this.name);
    }
    /**
     * The organization&#39;s name.
     * 
     */
    @Export(name="organizationName", refs={String.class}, tree="[0]")
    private Output</* @Nullable */ String> organizationName;

    /**
     * @return The organization&#39;s name.
     * 
     */
    public Output<Optional<String>> organizationName() {
        return Codegen.optional(this.organizationName);
    }
    /**
     * The team name.
     * 
     */
    @Export(name="teamName", refs={String.class}, tree="[0]")
    private Output</* @Nullable */ String> teamName;

    /**
     * @return The team name.
     * 
     */
    public Output<Optional<String>> teamName() {
        return Codegen.optional(this.teamName);
    }

    /**
     *
     * @param name The _unique_ name of the resulting resource.
     */
    public TeamAccessToken(String name) {
        this(name, TeamAccessTokenArgs.Empty);
    }
    /**
     *
     * @param name The _unique_ name of the resulting resource.
     * @param args The arguments to use to populate this resource's properties.
     */
    public TeamAccessToken(String name, TeamAccessTokenArgs args) {
        this(name, args, null);
    }
    /**
     *
     * @param name The _unique_ name of the resulting resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param options A bag of options that control this resource's behavior.
     */
    public TeamAccessToken(String name, TeamAccessTokenArgs args, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        super("pulumiservice:index:TeamAccessToken", name, args == null ? TeamAccessTokenArgs.Empty : args, makeResourceOptions(options, Codegen.empty()));
    }

    private TeamAccessToken(String name, Output<String> id, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        super("pulumiservice:index:TeamAccessToken", name, null, makeResourceOptions(options, id));
    }

    private static com.pulumi.resources.CustomResourceOptions makeResourceOptions(@Nullable com.pulumi.resources.CustomResourceOptions options, @Nullable Output<String> id) {
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
    public static TeamAccessToken get(String name, Output<String> id, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        return new TeamAccessToken(name, id, options);
    }
}
