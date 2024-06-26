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
        DriftRemediationFailed("drift_remediation_failed");

        private final String value;

        WebhookFilters(String value) {
            this.value = Objects.requireNonNull(value);
        }

        @EnumType.Converter
        public String getValue() {
            return this.value;
        }

        @Override
        public String toString() {
            return new StringJoiner(", ", "WebhookFilters[", "]")
                .add("value='" + this.value + "'")
                .toString();
        }
    }
