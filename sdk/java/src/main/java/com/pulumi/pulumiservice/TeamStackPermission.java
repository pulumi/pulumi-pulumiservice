// *** WARNING: this file was generated by pulumi. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package com.pulumi.pulumiservice;

import com.pulumi.core.Output;
import com.pulumi.core.annotations.ResourceType;
import com.pulumi.core.internal.Codegen;
import com.pulumi.pulumiservice.TeamStackPermissionArgs;
import com.pulumi.pulumiservice.Utilities;
import javax.annotation.Nullable;

/**
 * Grants a team permissions to the specified stack.
 * 
 */
@ResourceType(type="pulumiservice:index:TeamStackPermission")
public class TeamStackPermission extends com.pulumi.resources.CustomResource {
    /**
     *
     * @param name The _unique_ name of the resulting resource.
     */
    public TeamStackPermission(String name) {
        this(name, TeamStackPermissionArgs.Empty);
    }
    /**
     *
     * @param name The _unique_ name of the resulting resource.
     * @param args The arguments to use to populate this resource's properties.
     */
    public TeamStackPermission(String name, TeamStackPermissionArgs args) {
        this(name, args, null);
    }
    /**
     *
     * @param name The _unique_ name of the resulting resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param options A bag of options that control this resource's behavior.
     */
    public TeamStackPermission(String name, TeamStackPermissionArgs args, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        super("pulumiservice:index:TeamStackPermission", name, args == null ? TeamStackPermissionArgs.Empty : args, makeResourceOptions(options, Codegen.empty()));
    }

    private TeamStackPermission(String name, Output<String> id, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        super("pulumiservice:index:TeamStackPermission", name, null, makeResourceOptions(options, id));
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
    public static TeamStackPermission get(String name, Output<String> id, @Nullable com.pulumi.resources.CustomResourceOptions options) {
        return new TeamStackPermission(name, id, options);
    }
}
