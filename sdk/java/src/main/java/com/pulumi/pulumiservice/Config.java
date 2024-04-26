// *** WARNING: this file was generated by pulumi. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package com.pulumi.pulumiservice;

import com.pulumi.core.internal.Codegen;
import java.lang.String;
import java.util.Optional;

public final class Config {

    private static final com.pulumi.Config config = com.pulumi.Config.of("pulumiservice");
/**
 * Access Token to authenticate with Pulumi Cloud.
 * 
 */
    public Optional<String> accessToken() {
        return Codegen.stringProp("accessToken").config(config).env("PULUMI_ACCESS_TOKEN").get();
    }
/**
 * The service URL used to reach Pulumi Cloud.
 * 
 */
    public Optional<String> serviceURL() {
        return Codegen.stringProp("serviceURL").config(config).env("PULUMI_BACKEND_URL").def("https://api.pulumi.com").get();
    }
}
