// *** WARNING: this file was generated by pulumi. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package com.pulumi.pulumiservice.outputs;

import com.pulumi.core.annotations.CustomType;
import com.pulumi.pulumiservice.outputs.DeploymentSettingsGitSourceGitAuth;
import java.lang.String;
import java.util.Objects;
import java.util.Optional;
import javax.annotation.Nullable;

@CustomType
public final class DeploymentSettingsGitSource {
    /**
     * @return The branch to deploy. One of either `branch` or `commit` must be specified.
     * 
     */
    private @Nullable String branch;
    /**
     * @return The commit to deploy. One of either `branch` or `commit` must be specified.
     * 
     */
    private @Nullable String commit;
    /**
     * @return Git authentication configuration for this deployment. Should not be specified if there are `gitHub` settings for this deployment.
     * 
     */
    private @Nullable DeploymentSettingsGitSourceGitAuth gitAuth;
    /**
     * @return The directory within the repository where the Pulumi.yaml is located.
     * 
     */
    private @Nullable String repoDir;
    /**
     * @return The repository URL to use for git settings. Should not be specified if there are `gitHub` settings for this deployment.
     * 
     */
    private @Nullable String repoUrl;

    private DeploymentSettingsGitSource() {}
    /**
     * @return The branch to deploy. One of either `branch` or `commit` must be specified.
     * 
     */
    public Optional<String> branch() {
        return Optional.ofNullable(this.branch);
    }
    /**
     * @return The commit to deploy. One of either `branch` or `commit` must be specified.
     * 
     */
    public Optional<String> commit() {
        return Optional.ofNullable(this.commit);
    }
    /**
     * @return Git authentication configuration for this deployment. Should not be specified if there are `gitHub` settings for this deployment.
     * 
     */
    public Optional<DeploymentSettingsGitSourceGitAuth> gitAuth() {
        return Optional.ofNullable(this.gitAuth);
    }
    /**
     * @return The directory within the repository where the Pulumi.yaml is located.
     * 
     */
    public Optional<String> repoDir() {
        return Optional.ofNullable(this.repoDir);
    }
    /**
     * @return The repository URL to use for git settings. Should not be specified if there are `gitHub` settings for this deployment.
     * 
     */
    public Optional<String> repoUrl() {
        return Optional.ofNullable(this.repoUrl);
    }

    public static Builder builder() {
        return new Builder();
    }

    public static Builder builder(DeploymentSettingsGitSource defaults) {
        return new Builder(defaults);
    }
    @CustomType.Builder
    public static final class Builder {
        private @Nullable String branch;
        private @Nullable String commit;
        private @Nullable DeploymentSettingsGitSourceGitAuth gitAuth;
        private @Nullable String repoDir;
        private @Nullable String repoUrl;
        public Builder() {}
        public Builder(DeploymentSettingsGitSource defaults) {
    	      Objects.requireNonNull(defaults);
    	      this.branch = defaults.branch;
    	      this.commit = defaults.commit;
    	      this.gitAuth = defaults.gitAuth;
    	      this.repoDir = defaults.repoDir;
    	      this.repoUrl = defaults.repoUrl;
        }

        @CustomType.Setter
        public Builder branch(@Nullable String branch) {

            this.branch = branch;
            return this;
        }
        @CustomType.Setter
        public Builder commit(@Nullable String commit) {

            this.commit = commit;
            return this;
        }
        @CustomType.Setter
        public Builder gitAuth(@Nullable DeploymentSettingsGitSourceGitAuth gitAuth) {

            this.gitAuth = gitAuth;
            return this;
        }
        @CustomType.Setter
        public Builder repoDir(@Nullable String repoDir) {

            this.repoDir = repoDir;
            return this;
        }
        @CustomType.Setter
        public Builder repoUrl(@Nullable String repoUrl) {

            this.repoUrl = repoUrl;
            return this;
        }
        public DeploymentSettingsGitSource build() {
            final var _resultValue = new DeploymentSettingsGitSource();
            _resultValue.branch = branch;
            _resultValue.commit = commit;
            _resultValue.gitAuth = gitAuth;
            _resultValue.repoDir = repoDir;
            _resultValue.repoUrl = repoUrl;
            return _resultValue;
        }
    }
}
