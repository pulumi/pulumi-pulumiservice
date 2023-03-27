// *** WARNING: this file was generated by pulumi-java-gen. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package com.pulumi.pulumiservice.inputs;

import com.pulumi.core.Output;
import com.pulumi.core.annotations.Import;
import java.lang.Integer;
import java.lang.String;
import java.util.Objects;
import java.util.Optional;
import javax.annotation.Nullable;


public final class GCPOIDCConfigurationArgs extends com.pulumi.resources.ResourceArgs {

    public static final GCPOIDCConfigurationArgs Empty = new GCPOIDCConfigurationArgs();

    /**
     * The numerical ID of the GCP project.
     * 
     */
    @Import(name="projectId")
    private @Nullable Output<String> projectId;

    /**
     * @return The numerical ID of the GCP project.
     * 
     */
    public Optional<Output<String>> projectId() {
        return Optional.ofNullable(this.projectId);
    }

    /**
     * The ID of the identity provider associated with the workload pool.
     * 
     */
    @Import(name="providerId")
    private @Nullable Output<String> providerId;

    /**
     * @return The ID of the identity provider associated with the workload pool.
     * 
     */
    public Optional<Output<String>> providerId() {
        return Optional.ofNullable(this.providerId);
    }

    /**
     * The region of the GCP project.
     * 
     */
    @Import(name="region")
    private @Nullable Output<String> region;

    /**
     * @return The region of the GCP project.
     * 
     */
    public Optional<Output<String>> region() {
        return Optional.ofNullable(this.region);
    }

    /**
     * The email address of the service account to use.
     * 
     */
    @Import(name="serviceAccount")
    private @Nullable Output<String> serviceAccount;

    /**
     * @return The email address of the service account to use.
     * 
     */
    public Optional<Output<String>> serviceAccount() {
        return Optional.ofNullable(this.serviceAccount);
    }

    /**
     * The lifetime of the temporary credentials.
     * 
     */
    @Import(name="tokenLifetime")
    private @Nullable Output<Integer> tokenLifetime;

    /**
     * @return The lifetime of the temporary credentials.
     * 
     */
    public Optional<Output<Integer>> tokenLifetime() {
        return Optional.ofNullable(this.tokenLifetime);
    }

    /**
     * The ID of the workload pool to use.
     * 
     */
    @Import(name="workloadPoolId")
    private @Nullable Output<String> workloadPoolId;

    /**
     * @return The ID of the workload pool to use.
     * 
     */
    public Optional<Output<String>> workloadPoolId() {
        return Optional.ofNullable(this.workloadPoolId);
    }

    private GCPOIDCConfigurationArgs() {}

    private GCPOIDCConfigurationArgs(GCPOIDCConfigurationArgs $) {
        this.projectId = $.projectId;
        this.providerId = $.providerId;
        this.region = $.region;
        this.serviceAccount = $.serviceAccount;
        this.tokenLifetime = $.tokenLifetime;
        this.workloadPoolId = $.workloadPoolId;
    }

    public static Builder builder() {
        return new Builder();
    }
    public static Builder builder(GCPOIDCConfigurationArgs defaults) {
        return new Builder(defaults);
    }

    public static final class Builder {
        private GCPOIDCConfigurationArgs $;

        public Builder() {
            $ = new GCPOIDCConfigurationArgs();
        }

        public Builder(GCPOIDCConfigurationArgs defaults) {
            $ = new GCPOIDCConfigurationArgs(Objects.requireNonNull(defaults));
        }

        /**
         * @param projectId The numerical ID of the GCP project.
         * 
         * @return builder
         * 
         */
        public Builder projectId(@Nullable Output<String> projectId) {
            $.projectId = projectId;
            return this;
        }

        /**
         * @param projectId The numerical ID of the GCP project.
         * 
         * @return builder
         * 
         */
        public Builder projectId(String projectId) {
            return projectId(Output.of(projectId));
        }

        /**
         * @param providerId The ID of the identity provider associated with the workload pool.
         * 
         * @return builder
         * 
         */
        public Builder providerId(@Nullable Output<String> providerId) {
            $.providerId = providerId;
            return this;
        }

        /**
         * @param providerId The ID of the identity provider associated with the workload pool.
         * 
         * @return builder
         * 
         */
        public Builder providerId(String providerId) {
            return providerId(Output.of(providerId));
        }

        /**
         * @param region The region of the GCP project.
         * 
         * @return builder
         * 
         */
        public Builder region(@Nullable Output<String> region) {
            $.region = region;
            return this;
        }

        /**
         * @param region The region of the GCP project.
         * 
         * @return builder
         * 
         */
        public Builder region(String region) {
            return region(Output.of(region));
        }

        /**
         * @param serviceAccount The email address of the service account to use.
         * 
         * @return builder
         * 
         */
        public Builder serviceAccount(@Nullable Output<String> serviceAccount) {
            $.serviceAccount = serviceAccount;
            return this;
        }

        /**
         * @param serviceAccount The email address of the service account to use.
         * 
         * @return builder
         * 
         */
        public Builder serviceAccount(String serviceAccount) {
            return serviceAccount(Output.of(serviceAccount));
        }

        /**
         * @param tokenLifetime The lifetime of the temporary credentials.
         * 
         * @return builder
         * 
         */
        public Builder tokenLifetime(@Nullable Output<Integer> tokenLifetime) {
            $.tokenLifetime = tokenLifetime;
            return this;
        }

        /**
         * @param tokenLifetime The lifetime of the temporary credentials.
         * 
         * @return builder
         * 
         */
        public Builder tokenLifetime(Integer tokenLifetime) {
            return tokenLifetime(Output.of(tokenLifetime));
        }

        /**
         * @param workloadPoolId The ID of the workload pool to use.
         * 
         * @return builder
         * 
         */
        public Builder workloadPoolId(@Nullable Output<String> workloadPoolId) {
            $.workloadPoolId = workloadPoolId;
            return this;
        }

        /**
         * @param workloadPoolId The ID of the workload pool to use.
         * 
         * @return builder
         * 
         */
        public Builder workloadPoolId(String workloadPoolId) {
            return workloadPoolId(Output.of(workloadPoolId));
        }

        public GCPOIDCConfigurationArgs build() {
            return $;
        }
    }

}
