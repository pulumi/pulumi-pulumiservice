// *** WARNING: this file was generated by pulumi. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package com.pulumi.pulumiservice;

import com.pulumi.core.Output;
import com.pulumi.core.annotations.Export;
import com.pulumi.core.annotations.ResourceType;
import com.pulumi.core.internal.Codegen;
import com.pulumi.pulumiservice.TeamStackPermissionArgs;
import com.pulumi.pulumiservice.Utilities;
import com.pulumi.pulumiservice.enums.TeamStackPermissionScope;
import java.lang.String;
import javax.annotation.Nullable;

/**
 * Grants a team permissions to the specified stack.
 * 
 */
@ResourceType(type="pulumiservice:index:TeamStackPermission")
public class TeamStackPermission extends com.pulumi.resources.CustomResource {
    /**
     * The organization or the personal account name of the stack.
     * 
     */
    @Export(name="organization", refs={String.class}, tree="[0]")
    private Output<String> organization;

    /**
     * @return The organization or the personal account name of the stack.
     * 
     */
    public Output<String> organization() {
        return this.organization;
    }
    /**
     * Sets the permission level that this team will be granted to the stack.
     * 
     */
    @Export(name="permission", refs={TeamStackPermissionScope.class}, tree="[0]")
    private Output<TeamStackPermissionScope> permission;

    /**
     * @return Sets the permission level that this team will be granted to the stack.
     * 
     */
    public Output<TeamStackPermissionScope> permission() {
        return this.permission;
    }
    /**
     * The project name for this stack.
     * 
     */
    @Export(name="project", refs={String.class}, tree="[0]")
    private Output<String> project;

    /**
     * @return The project name for this stack.
     * 
     */
    public Output<String> project() {
        return this.project;
    }
    /**
     * The name of the stack that the team will be granted permissions to.
     * 
     */
    @Export(name="stack", refs={String.class}, tree="[0]")
    private Output<String> stack;

    /**
     * @return The name of the stack that the team will be granted permissions to.
     * 
     */
    public Output<String> stack() {
        return this.stack;
    }
    /**
     * The name of the team to grant this stack permissions to. This is not the display name.
     * 
     */
    @Export(name="team", refs={String.class}, tree="[0]")
    private Output<String> team;

    /**
     * @return The name of the team to grant this stack permissions to. This is not the display name.
     * 
     */
    public Output<String> team() {
        return this.team;
    }

    /**
     *
     * @param name The _unique_ name of the resulting resource.
     */
    public TeamStackPermission(java.lang.String name) {
        this(name, TeamStackPermissionArgs.Empty);
    }
    /**
     *
     * @param name The _unique_ name of the resulting resource.
     * @param args The arguments to use to populate this resource's properties.
     */
    public TeamStackPermission(java.lang.String name, TeamStackPermissionArgs args) {
        this(name, args, null);
    }
    /**
     *
     * @param name The _unique_ name of the resulting resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param options A bag of options that control this resource's behavior.
     */
    public TeamStackPermission(java.lang.String name, TeamStackPermissionArgs args, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        super("pulumiservice:index:TeamStackPermission", name, makeArgs(args, options), makeResourceOptions(options, Codegen.empty()), false);
    }

    private TeamStackPermission(java.lang.String name, Output<java.lang.String> id, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        super("pulumiservice:index:TeamStackPermission", name, null, makeResourceOptions(options, id), false);
    }

    private static TeamStackPermissionArgs makeArgs(TeamStackPermissionArgs args, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        if (options != null && options.getUrn().isPresent()) {
            return null;
        }
        return args == null ? TeamStackPermissionArgs.Empty : args;
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
    public static TeamStackPermission get(java.lang.String name, Output<java.lang.String> id, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        return new TeamStackPermission(name, id, options);
    }
}
