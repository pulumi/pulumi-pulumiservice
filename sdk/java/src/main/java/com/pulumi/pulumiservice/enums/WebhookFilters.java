// *** WARNING: this file was generated by pulumi. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package com.pulumi.pulumiservice.enums;

import com.pulumi.core.annotations.EnumType;
import java.lang.String;
import java.util.Objects;
import java.util.StringJoiner;

    @EnumType
    public enum WebhookFilters {
        /**
         * Trigger a webhook when a stack is created. Only valid for org webhooks.
         * 
         */
        StackCreated("stack_created"),
        /**
         * Trigger a webhook when a stack is deleted. Only valid for org webhooks.
         * 
         */
        StackDeleted("stack_deleted"),
        /**
         * Trigger a webhook when a stack update succeeds.
         * 
         */
        UpdateSucceeded("update_succeeded"),
        /**
         * Trigger a webhook when a stack update fails.
         * 
         */
        UpdateFailed("update_failed"),
        /**
         * Trigger a webhook when a stack preview succeeds.
         * 
         */
        PreviewSucceeded("preview_succeeded"),
        /**
         * Trigger a webhook when a stack preview fails.
         * 
         */
        PreviewFailed("preview_failed"),
        /**
         * Trigger a webhook when a stack destroy succeeds.
         * 
         */
        DestroySucceeded("destroy_succeeded"),
        /**
         * Trigger a webhook when a stack destroy fails.
         * 
         */
        DestroyFailed("destroy_failed"),
        /**
         * Trigger a webhook when a stack refresh succeeds.
         * 
         */
        RefreshSucceeded("refresh_succeeded"),
        /**
         * Trigger a webhook when a stack refresh fails.
         * 
         */
        RefreshFailed("refresh_failed"),
        /**
         * Trigger a webhook when a deployment is queued.
         * 
         */
        DeploymentQueued("deployment_queued"),
        /**
         * Trigger a webhook when a deployment starts running.
         * 
         */
        DeploymentStarted("deployment_started"),
        /**
         * Trigger a webhook when a deployment succeeds.
         * 
         */
        DeploymentSucceeded("deployment_succeeded"),
        /**
         * Trigger a webhook when a deployment fails.
         * 
         */
        DeploymentFailed("deployment_failed"),
        /**
         * Trigger a webhook when drift is detected.
         * 
         */
        DriftDetected("drift_detected"),
        /**
         * Trigger a webhook when a drift detection run succeeds, regardless of whether drift is detected.
         * 
         */
        DriftDetectionSucceeded("drift_detection_succeeded"),
        /**
         * Trigger a webhook when a drift detection run fails.
         * 
         */
        DriftDetectionFailed("drift_detection_failed"),
        /**
         * Trigger a webhook when a drift remediation run succeeds.
         * 
         */
        DriftRemediationSucceeded("drift_remediation_succeeded"),
        /**
         * Trigger a webhook when a drift remediation run fails.
         * 
         */
        DriftRemediationFailed("drift_remediation_failed"),
        /**
         * Trigger a webhook when a new environment is created.
         * 
         */
        EnvironmentCreated("environment_created"),
        /**
         * Trigger a webhook when an environment is deleted.
         * 
         */
        EnvironmentDeleted("environment_deleted"),
        /**
         * Trigger a webhook when a new revision is created on an environment.
         * 
         */
        EnvironmentRevisionCreated("environment_revision_created"),
        /**
         * Trigger a webhook when a revision is retracted on an environment.
         * 
         */
        EnvironmentRevisionRetracted("environment_revision_retracted"),
        /**
         * Trigger a webhook when a revision tag is created on an environment.
         * 
         */
        EnvironmentRevisionTagCreated("environment_revision_tag_created"),
        /**
         * Trigger a webhook when a revision tag is deleted on an environment.
         * 
         */
        EnvironmentRevisionTagDeleted("environment_revision_tag_deleted"),
        /**
         * Trigger a webhook when a revision tag is updated on an environment.
         * 
         */
        EnvironmentRevisionTagUpdated("environment_revision_tag_updated"),
        /**
         * Trigger a webhook when an environment tag is created.
         * 
         */
        EnvironmentTagCreated("environment_tag_created"),
        /**
         * Trigger a webhook when an environment tag is deleted.
         * 
         */
        EnvironmentTagDeleted("environment_tag_deleted"),
        /**
         * Trigger a webhook when an environment tag is updated.
         * 
         */
        EnvironmentTagUpdated("environment_tag_updated"),
        /**
         * Trigger a webhook when an imported environment has changed.
         * 
         */
        ImportedEnvironmentChanged("imported_environment_changed");

        private final String value;

        WebhookFilters(String value) {
            this.value = Objects.requireNonNull(value);
        }

        @EnumType.Converter
        public String getValue() {
            return this.value;
        }

        @Override
        public java.lang.String toString() {
            return new StringJoiner(", ", "WebhookFilters[", "]")
                .add("value='" + this.value + "'")
                .toString();
        }
    }
