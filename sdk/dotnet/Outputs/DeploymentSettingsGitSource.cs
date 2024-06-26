// *** WARNING: this file was generated by pulumi. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;

namespace Pulumi.PulumiService.Outputs
{

    /// <summary>
    /// Git source settings for a deployment.
    /// </summary>
    [OutputType]
    public sealed class DeploymentSettingsGitSource
    {
        /// <summary>
        /// The branch to deploy. One of either `branch` or `commit` must be specified.
        /// </summary>
        public readonly string? Branch;
        /// <summary>
        /// The commit to deploy. One of either `branch` or `commit` must be specified.
        /// </summary>
        public readonly string? Commit;
        /// <summary>
        /// Git authentication configuration for this deployment. Should not be specified if there are `gitHub` settings for this deployment.
        /// </summary>
        public readonly Outputs.DeploymentSettingsGitSourceGitAuth? GitAuth;
        /// <summary>
        /// The directory within the repository where the Pulumi.yaml is located.
        /// </summary>
        public readonly string? RepoDir;
        /// <summary>
        /// The repository URL to use for git settings. Should not be specified if there are `gitHub` settings for this deployment.
        /// </summary>
        public readonly string? RepoUrl;

        [OutputConstructor]
        private DeploymentSettingsGitSource(
            string? branch,

            string? commit,

            Outputs.DeploymentSettingsGitSourceGitAuth? gitAuth,

            string? repoDir,

            string? repoUrl)
        {
            Branch = branch;
            Commit = commit;
            GitAuth = gitAuth;
            RepoDir = repoDir;
            RepoUrl = repoUrl;
        }
    }
}
