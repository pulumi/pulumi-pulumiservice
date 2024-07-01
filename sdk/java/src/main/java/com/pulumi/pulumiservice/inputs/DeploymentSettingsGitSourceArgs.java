// *** WARNING: this file was generated by pulumi. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package com.pulumi.pulumiservice.inputs;

import com.pulumi.core.Output;
import com.pulumi.core.annotations.Import;
import com.pulumi.pulumiservice.inputs.DeploymentSettingsGitSourceGitAuthArgs;
import java.lang.String;
import java.util.Objects;
import java.util.Optional;
import javax.annotation.Nullable;


/**
 * Git source settings for a deployment.
 * 
 */
public final class DeploymentSettingsGitSourceArgs extends com.pulumi.resources.ResourceArgs {

    public static final DeploymentSettingsGitSourceArgs Empty = new DeploymentSettingsGitSourceArgs();

    /**
     * The branch to deploy. One of either `branch` or `commit` must be specified.
     * 
     */
    @Import(name="branch")
    private @Nullable Output<String> branch;

    /**
     * @return The branch to deploy. One of either `branch` or `commit` must be specified.
     * 
     */
    public Optional<Output<String>> branch() {
        return Optional.ofNullable(this.branch);
    }

    /**
     * The commit to deploy. One of either `branch` or `commit` must be specified.
     * 
     */
    @Import(name="commit")
    private @Nullable Output<String> commit;

    /**
     * @return The commit to deploy. One of either `branch` or `commit` must be specified.
     * 
     */
    public Optional<Output<String>> commit() {
        return Optional.ofNullable(this.commit);
    }

    /**
     * Git authentication configuration for this deployment. Should not be specified if there are `gitHub` settings for this deployment.
     * 
     */
    @Import(name="gitAuth")
    private @Nullable Output<DeploymentSettingsGitSourceGitAuthArgs> gitAuth;

    /**
     * @return Git authentication configuration for this deployment. Should not be specified if there are `gitHub` settings for this deployment.
     * 
     */
    public Optional<Output<DeploymentSettingsGitSourceGitAuthArgs>> gitAuth() {
        return Optional.ofNullable(this.gitAuth);
    }

    /**
     * The directory within the repository where the Pulumi.yaml is located.
     * 
     */
    @Import(name="repoDir")
    private @Nullable Output<String> repoDir;

    /**
     * @return The directory within the repository where the Pulumi.yaml is located.
     * 
     */
    public Optional<Output<String>> repoDir() {
        return Optional.ofNullable(this.repoDir);
    }

    /**
     * The repository URL to use for git settings. Should not be specified if there are `gitHub` settings for this deployment.
     * 
     */
    @Import(name="repoUrl")
    private @Nullable Output<String> repoUrl;

    /**
     * @return The repository URL to use for git settings. Should not be specified if there are `gitHub` settings for this deployment.
     * 
     */
    public Optional<Output<String>> repoUrl() {
        return Optional.ofNullable(this.repoUrl);
    }

    private DeploymentSettingsGitSourceArgs() {}

    private DeploymentSettingsGitSourceArgs(DeploymentSettingsGitSourceArgs $) {
        this.branch = $.branch;
        this.commit = $.commit;
        this.gitAuth = $.gitAuth;
        this.repoDir = $.repoDir;
        this.repoUrl = $.repoUrl;
    }

    public static Builder builder() {
        return new Builder();
    }
    public static Builder builder(DeploymentSettingsGitSourceArgs defaults) {
        return new Builder(defaults);
    }

    public static final class Builder {
        private DeploymentSettingsGitSourceArgs $;

        public Builder() {
            $ = new DeploymentSettingsGitSourceArgs();
        }

        public Builder(DeploymentSettingsGitSourceArgs defaults) {
            $ = new DeploymentSettingsGitSourceArgs(Objects.requireNonNull(defaults));
        }

        /**
         * @param branch The branch to deploy. One of either `branch` or `commit` must be specified.
         * 
         * @return builder
         * 
         */
        public Builder branch(@Nullable Output<String> branch) {
            $.branch = branch;
            return this;
        }

        /**
         * @param branch The branch to deploy. One of either `branch` or `commit` must be specified.
         * 
         * @return builder
         * 
         */
        public Builder branch(String branch) {
            return branch(Output.of(branch));
        }

        /**
         * @param commit The commit to deploy. One of either `branch` or `commit` must be specified.
         * 
         * @return builder
         * 
         */
        public Builder commit(@Nullable Output<String> commit) {
            $.commit = commit;
            return this;
        }

        /**
         * @param commit The commit to deploy. One of either `branch` or `commit` must be specified.
         * 
         * @return builder
         * 
         */
        public Builder commit(String commit) {
            return commit(Output.of(commit));
        }

        /**
         * @param gitAuth Git authentication configuration for this deployment. Should not be specified if there are `gitHub` settings for this deployment.
         * 
         * @return builder
         * 
         */
        public Builder gitAuth(@Nullable Output<DeploymentSettingsGitSourceGitAuthArgs> gitAuth) {
            $.gitAuth = gitAuth;
            return this;
        }

        /**
         * @param gitAuth Git authentication configuration for this deployment. Should not be specified if there are `gitHub` settings for this deployment.
         * 
         * @return builder
         * 
         */
        public Builder gitAuth(DeploymentSettingsGitSourceGitAuthArgs gitAuth) {
            return gitAuth(Output.of(gitAuth));
        }

        /**
         * @param repoDir The directory within the repository where the Pulumi.yaml is located.
         * 
         * @return builder
         * 
         */
        public Builder repoDir(@Nullable Output<String> repoDir) {
            $.repoDir = repoDir;
            return this;
        }

        /**
         * @param repoDir The directory within the repository where the Pulumi.yaml is located.
         * 
         * @return builder
         * 
         */
        public Builder repoDir(String repoDir) {
            return repoDir(Output.of(repoDir));
        }

        /**
         * @param repoUrl The repository URL to use for git settings. Should not be specified if there are `gitHub` settings for this deployment.
         * 
         * @return builder
         * 
         */
        public Builder repoUrl(@Nullable Output<String> repoUrl) {
            $.repoUrl = repoUrl;
            return this;
        }

        /**
         * @param repoUrl The repository URL to use for git settings. Should not be specified if there are `gitHub` settings for this deployment.
         * 
         * @return builder
         * 
         */
        public Builder repoUrl(String repoUrl) {
            return repoUrl(Output.of(repoUrl));
        }

        public DeploymentSettingsGitSourceArgs build() {
            return $;
        }
    }

}
