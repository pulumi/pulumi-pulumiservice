// *** WARNING: this file was generated by pulumi. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package com.pulumi.pulumiservice.inputs;

import com.pulumi.core.Output;
import com.pulumi.core.annotations.Import;
import java.lang.Boolean;
import java.lang.String;
import java.util.Objects;
import java.util.Optional;
import javax.annotation.Nullable;


public final class OperationContextOptionsArgs extends com.pulumi.resources.ResourceArgs {

    public static final OperationContextOptionsArgs Empty = new OperationContextOptionsArgs();

    /**
     * Whether the stack should be deleted after it is destroyed.
     * 
     */
    @Import(name="deleteAfterDestroy")
    private @Nullable Output<Boolean> deleteAfterDestroy;

    /**
     * @return Whether the stack should be deleted after it is destroyed.
     * 
     */
    public Optional<Output<Boolean>> deleteAfterDestroy() {
        return Optional.ofNullable(this.deleteAfterDestroy);
    }

    /**
     * The shell to use to run commands during the deployment. Defaults to &#39;bash&#39;.
     * 
     */
    @Import(name="shell")
    private @Nullable Output<String> shell;

    /**
     * @return The shell to use to run commands during the deployment. Defaults to &#39;bash&#39;.
     * 
     */
    public Optional<Output<String>> shell() {
        return Optional.ofNullable(this.shell);
    }

    /**
     * Skip the default dependency installation step - use this to customize the dependency installation (e.g. if using yarn or poetry)
     * 
     */
    @Import(name="skipInstallDependencies")
    private @Nullable Output<Boolean> skipInstallDependencies;

    /**
     * @return Skip the default dependency installation step - use this to customize the dependency installation (e.g. if using yarn or poetry)
     * 
     */
    public Optional<Output<Boolean>> skipInstallDependencies() {
        return Optional.ofNullable(this.skipInstallDependencies);
    }

    /**
     * Skip intermediate deployments (Consolidate multiple deployments of the same type into one deployment)
     * 
     */
    @Import(name="skipIntermediateDeployments")
    private @Nullable Output<Boolean> skipIntermediateDeployments;

    /**
     * @return Skip intermediate deployments (Consolidate multiple deployments of the same type into one deployment)
     * 
     */
    public Optional<Output<Boolean>> skipIntermediateDeployments() {
        return Optional.ofNullable(this.skipIntermediateDeployments);
    }

    private OperationContextOptionsArgs() {}

    private OperationContextOptionsArgs(OperationContextOptionsArgs $) {
        this.deleteAfterDestroy = $.deleteAfterDestroy;
        this.shell = $.shell;
        this.skipInstallDependencies = $.skipInstallDependencies;
        this.skipIntermediateDeployments = $.skipIntermediateDeployments;
    }

    public static Builder builder() {
        return new Builder();
    }
    public static Builder builder(OperationContextOptionsArgs defaults) {
        return new Builder(defaults);
    }

    public static final class Builder {
        private OperationContextOptionsArgs $;

        public Builder() {
            $ = new OperationContextOptionsArgs();
        }

        public Builder(OperationContextOptionsArgs defaults) {
            $ = new OperationContextOptionsArgs(Objects.requireNonNull(defaults));
        }

        /**
         * @param deleteAfterDestroy Whether the stack should be deleted after it is destroyed.
         * 
         * @return builder
         * 
         */
        public Builder deleteAfterDestroy(@Nullable Output<Boolean> deleteAfterDestroy) {
            $.deleteAfterDestroy = deleteAfterDestroy;
            return this;
        }

        /**
         * @param deleteAfterDestroy Whether the stack should be deleted after it is destroyed.
         * 
         * @return builder
         * 
         */
        public Builder deleteAfterDestroy(Boolean deleteAfterDestroy) {
            return deleteAfterDestroy(Output.of(deleteAfterDestroy));
        }

        /**
         * @param shell The shell to use to run commands during the deployment. Defaults to &#39;bash&#39;.
         * 
         * @return builder
         * 
         */
        public Builder shell(@Nullable Output<String> shell) {
            $.shell = shell;
            return this;
        }

        /**
         * @param shell The shell to use to run commands during the deployment. Defaults to &#39;bash&#39;.
         * 
         * @return builder
         * 
         */
        public Builder shell(String shell) {
            return shell(Output.of(shell));
        }

        /**
         * @param skipInstallDependencies Skip the default dependency installation step - use this to customize the dependency installation (e.g. if using yarn or poetry)
         * 
         * @return builder
         * 
         */
        public Builder skipInstallDependencies(@Nullable Output<Boolean> skipInstallDependencies) {
            $.skipInstallDependencies = skipInstallDependencies;
            return this;
        }

        /**
         * @param skipInstallDependencies Skip the default dependency installation step - use this to customize the dependency installation (e.g. if using yarn or poetry)
         * 
         * @return builder
         * 
         */
        public Builder skipInstallDependencies(Boolean skipInstallDependencies) {
            return skipInstallDependencies(Output.of(skipInstallDependencies));
        }

        /**
         * @param skipIntermediateDeployments Skip intermediate deployments (Consolidate multiple deployments of the same type into one deployment)
         * 
         * @return builder
         * 
         */
        public Builder skipIntermediateDeployments(@Nullable Output<Boolean> skipIntermediateDeployments) {
            $.skipIntermediateDeployments = skipIntermediateDeployments;
            return this;
        }

        /**
         * @param skipIntermediateDeployments Skip intermediate deployments (Consolidate multiple deployments of the same type into one deployment)
         * 
         * @return builder
         * 
         */
        public Builder skipIntermediateDeployments(Boolean skipIntermediateDeployments) {
            return skipIntermediateDeployments(Output.of(skipIntermediateDeployments));
        }

        public OperationContextOptionsArgs build() {
            return $;
        }
    }

}
