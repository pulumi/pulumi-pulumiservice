// *** WARNING: this file was generated by pulumi. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package com.pulumi.pulumiservice;

import com.pulumi.core.Output;
import com.pulumi.core.annotations.Import;
import com.pulumi.exceptions.MissingRequiredPropertyException;
import java.lang.Boolean;
import java.lang.String;
import java.util.Objects;
import java.util.Optional;
import javax.annotation.Nullable;


public final class StackArgs extends com.pulumi.resources.ResourceArgs {

    public static final StackArgs Empty = new StackArgs();

    /**
     * Optional. Flag indicating whether to delete the stack even if it still contains resources.
     * 
     */
    @Import(name="forceDestroy")
    private @Nullable Output<Boolean> forceDestroy;

    /**
     * @return Optional. Flag indicating whether to delete the stack even if it still contains resources.
     * 
     */
    public Optional<Output<Boolean>> forceDestroy() {
        return Optional.ofNullable(this.forceDestroy);
    }

    /**
     * The name of the organization.
     * 
     */
    @Import(name="organizationName", required=true)
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
    @Import(name="projectName", required=true)
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
    @Import(name="stackName", required=true)
    private Output<String> stackName;

    /**
     * @return The name of the stack.
     * 
     */
    public Output<String> stackName() {
        return this.stackName;
    }

    private StackArgs() {}

    private StackArgs(StackArgs $) {
        this.forceDestroy = $.forceDestroy;
        this.organizationName = $.organizationName;
        this.projectName = $.projectName;
        this.stackName = $.stackName;
    }

    public static Builder builder() {
        return new Builder();
    }
    public static Builder builder(StackArgs defaults) {
        return new Builder(defaults);
    }

    public static final class Builder {
        private StackArgs $;

        public Builder() {
            $ = new StackArgs();
        }

        public Builder(StackArgs defaults) {
            $ = new StackArgs(Objects.requireNonNull(defaults));
        }

        /**
         * @param forceDestroy Optional. Flag indicating whether to delete the stack even if it still contains resources.
         * 
         * @return builder
         * 
         */
        public Builder forceDestroy(@Nullable Output<Boolean> forceDestroy) {
            $.forceDestroy = forceDestroy;
            return this;
        }

        /**
         * @param forceDestroy Optional. Flag indicating whether to delete the stack even if it still contains resources.
         * 
         * @return builder
         * 
         */
        public Builder forceDestroy(Boolean forceDestroy) {
            return forceDestroy(Output.of(forceDestroy));
        }

        /**
         * @param organizationName The name of the organization.
         * 
         * @return builder
         * 
         */
        public Builder organizationName(Output<String> organizationName) {
            $.organizationName = organizationName;
            return this;
        }

        /**
         * @param organizationName The name of the organization.
         * 
         * @return builder
         * 
         */
        public Builder organizationName(String organizationName) {
            return organizationName(Output.of(organizationName));
        }

        /**
         * @param projectName The name of the project.
         * 
         * @return builder
         * 
         */
        public Builder projectName(Output<String> projectName) {
            $.projectName = projectName;
            return this;
        }

        /**
         * @param projectName The name of the project.
         * 
         * @return builder
         * 
         */
        public Builder projectName(String projectName) {
            return projectName(Output.of(projectName));
        }

        /**
         * @param stackName The name of the stack.
         * 
         * @return builder
         * 
         */
        public Builder stackName(Output<String> stackName) {
            $.stackName = stackName;
            return this;
        }

        /**
         * @param stackName The name of the stack.
         * 
         * @return builder
         * 
         */
        public Builder stackName(String stackName) {
            return stackName(Output.of(stackName));
        }

        public StackArgs build() {
            if ($.organizationName == null) {
                throw new MissingRequiredPropertyException("StackArgs", "organizationName");
            }
            if ($.projectName == null) {
                throw new MissingRequiredPropertyException("StackArgs", "projectName");
            }
            if ($.stackName == null) {
                throw new MissingRequiredPropertyException("StackArgs", "stackName");
            }
            return $;
        }
    }

}
