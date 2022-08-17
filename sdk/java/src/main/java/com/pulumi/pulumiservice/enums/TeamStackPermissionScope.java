// *** WARNING: this file was generated by pulumi-java-gen. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package com.pulumi.pulumiservice.enums;

import com.pulumi.core.annotations.EnumType;
import java.lang.Double;
import java.util.StringJoiner;

    @EnumType
    public enum TeamStackPermissionScope {
        /**
         * Grants read permissions to stack.
         * 
         */
        Read(101.000000),
        /**
         * Grants edit permissions to stack.
         * 
         */
        Edit(102.000000),
        /**
         * Grants admin permissions to stack.
         * 
         */
        Admin(103.000000);

        private final Double value;

        TeamStackPermissionScope(Double value) {
            this.value = value;
        }

        @EnumType.Converter
        public Double getValue() {
            return this.value;
        }

        @Override
        public String toString() {
            return new StringJoiner(", ", "TeamStackPermissionScope[", "]")
                .add("value='" + this.value + "'")
                .toString();
        }
    }