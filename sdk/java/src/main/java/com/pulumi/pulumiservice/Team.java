// *** WARNING: this file was generated by pulumi-java-gen. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package com.pulumi.pulumiservice;

import com.pulumi.core.Output;
import com.pulumi.core.annotations.Export;
import com.pulumi.core.annotations.ResourceType;
import com.pulumi.core.internal.Codegen;
import com.pulumi.pulumiservice.TeamArgs;
import com.pulumi.pulumiservice.Utilities;
import java.lang.String;
import java.util.List;
import java.util.Optional;
import javax.annotation.Nullable;

/**
 * The Pulumi Service offers role-based access control (RBAC) using teams. Teams allow organization admins to assign a set of stack permissions to a group of users.
 * 
 */
@ResourceType(type="pulumiservice:index:Team")
public class Team extends com.pulumi.resources.CustomResource {
    /**
     * Optional. Team description.
     * 
     */
    @Export(name="description", type=String.class, parameters={})
    private Output</* @Nullable */ String> description;

    /**
     * @return Optional. Team description.
     * 
     */
    public Output<Optional<String>> description() {
        return Codegen.optional(this.description);
    }
    /**
     * Optional. Team display name.
     * 
     */
    @Export(name="displayName", type=String.class, parameters={})
    private Output</* @Nullable */ String> displayName;

    /**
     * @return Optional. Team display name.
     * 
     */
    public Output<Optional<String>> displayName() {
        return Codegen.optional(this.displayName);
    }
    /**
     * List of team members.
     * 
     */
    @Export(name="members", type=List.class, parameters={String.class})
    private Output</* @Nullable */ List<String>> members;

    /**
     * @return List of team members.
     * 
     */
    public Output<Optional<List<String>>> members() {
        return Codegen.optional(this.members);
    }
    /**
     * The team name.
     * 
     */
    @Export(name="name", type=String.class, parameters={})
    private Output</* @Nullable */ String> name;

    /**
     * @return The team name.
     * 
     */
    public Output<Optional<String>> name() {
        return Codegen.optional(this.name);
    }
    /**
     * The organization&#39;s name.
     * 
     */
    @Export(name="organizationName", type=String.class, parameters={})
    private Output</* @Nullable */ String> organizationName;

    /**
     * @return The organization&#39;s name.
     * 
     */
    public Output<Optional<String>> organizationName() {
        return Codegen.optional(this.organizationName);
    }
    /**
     * The type of team. Must be either `pulumi` or `github`.
     * 
     */
    @Export(name="teamType", type=String.class, parameters={})
    private Output</* @Nullable */ String> teamType;

    /**
     * @return The type of team. Must be either `pulumi` or `github`.
     * 
     */
    public Output<Optional<String>> teamType() {
        return Codegen.optional(this.teamType);
    }

    /**
     *
     * @param name The _unique_ name of the resulting resource.
     */
    public Team(String name) {
        this(name, TeamArgs.Empty);
    }
    /**
     *
     * @param name The _unique_ name of the resulting resource.
     * @param args The arguments to use to populate this resource's properties.
     */
    public Team(String name, TeamArgs args) {
        this(name, args, null);
    }
    /**
     *
     * @param name The _unique_ name of the resulting resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param options A bag of options that control this resource's behavior.
     */
    public Team(String name, TeamArgs args, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        super("pulumiservice:index:Team", name, args == null ? TeamArgs.Empty : args, makeResourceOptions(options, Codegen.empty()));
    }

    private Team(String name, Output<String> id, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        super("pulumiservice:index:Team", name, null, makeResourceOptions(options, id));
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
    public static Team get(String name, Output<String> id, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        return new Team(name, id, options);
    }
}
