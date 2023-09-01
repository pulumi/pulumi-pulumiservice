// *** WARNING: this file was generated by pulumigen. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as utilities from "./utilities";

// Export members:
export { AccessTokenArgs } from "./accessToken";
export type AccessToken = import("./accessToken").AccessToken;
export const AccessToken: typeof import("./accessToken").AccessToken = null as any;
utilities.lazyLoad(exports, ["AccessToken"], () => require("./accessToken"));

export { DeploymentSettingsArgs } from "./deploymentSettings";
export type DeploymentSettings = import("./deploymentSettings").DeploymentSettings;
export const DeploymentSettings: typeof import("./deploymentSettings").DeploymentSettings = null as any;
utilities.lazyLoad(exports, ["DeploymentSettings"], () => require("./deploymentSettings"));

export { GetDeploymentArgs, GetDeploymentResult, GetDeploymentOutputArgs } from "./getDeployment";
export const getDeployment: typeof import("./getDeployment").getDeployment = null as any;
export const getDeploymentOutput: typeof import("./getDeployment").getDeploymentOutput = null as any;
utilities.lazyLoad(exports, ["getDeployment","getDeploymentOutput"], () => require("./getDeployment"));

export { OrgAccessTokenArgs } from "./orgAccessToken";
export type OrgAccessToken = import("./orgAccessToken").OrgAccessToken;
export const OrgAccessToken: typeof import("./orgAccessToken").OrgAccessToken = null as any;
utilities.lazyLoad(exports, ["OrgAccessToken"], () => require("./orgAccessToken"));

export { ProviderArgs } from "./provider";
export type Provider = import("./provider").Provider;
export const Provider: typeof import("./provider").Provider = null as any;
utilities.lazyLoad(exports, ["Provider"], () => require("./provider"));

export { RunDeploymentArgs, RunDeploymentResult, RunDeploymentOutputArgs } from "./runDeployment";
export const runDeployment: typeof import("./runDeployment").runDeployment = null as any;
export const runDeploymentOutput: typeof import("./runDeployment").runDeploymentOutput = null as any;
utilities.lazyLoad(exports, ["runDeployment","runDeploymentOutput"], () => require("./runDeployment"));

export { StackTagArgs } from "./stackTag";
export type StackTag = import("./stackTag").StackTag;
export const StackTag: typeof import("./stackTag").StackTag = null as any;
utilities.lazyLoad(exports, ["StackTag"], () => require("./stackTag"));

export { TeamArgs } from "./team";
export type Team = import("./team").Team;
export const Team: typeof import("./team").Team = null as any;
utilities.lazyLoad(exports, ["Team"], () => require("./team"));

export { TeamAccessTokenArgs } from "./teamAccessToken";
export type TeamAccessToken = import("./teamAccessToken").TeamAccessToken;
export const TeamAccessToken: typeof import("./teamAccessToken").TeamAccessToken = null as any;
utilities.lazyLoad(exports, ["TeamAccessToken"], () => require("./teamAccessToken"));

export { TeamStackPermissionArgs } from "./teamStackPermission";
export type TeamStackPermission = import("./teamStackPermission").TeamStackPermission;
export const TeamStackPermission: typeof import("./teamStackPermission").TeamStackPermission = null as any;
utilities.lazyLoad(exports, ["TeamStackPermission"], () => require("./teamStackPermission"));

export { WebhookArgs } from "./webhook";
export type Webhook = import("./webhook").Webhook;
export const Webhook: typeof import("./webhook").Webhook = null as any;
utilities.lazyLoad(exports, ["Webhook"], () => require("./webhook"));


// Export enums:
export * from "./types/enums";

// Export sub-modules:
import * as config from "./config";
import * as types from "./types";

export {
    config,
    types,
};

const _module = {
    version: utilities.getVersion(),
    construct: (name: string, type: string, urn: string): pulumi.Resource => {
        switch (type) {
            case "pulumiservice:index:AccessToken":
                return new AccessToken(name, <any>undefined, { urn })
            case "pulumiservice:index:DeploymentSettings":
                return new DeploymentSettings(name, <any>undefined, { urn })
            case "pulumiservice:index:OrgAccessToken":
                return new OrgAccessToken(name, <any>undefined, { urn })
            case "pulumiservice:index:StackTag":
                return new StackTag(name, <any>undefined, { urn })
            case "pulumiservice:index:Team":
                return new Team(name, <any>undefined, { urn })
            case "pulumiservice:index:TeamAccessToken":
                return new TeamAccessToken(name, <any>undefined, { urn })
            case "pulumiservice:index:TeamStackPermission":
                return new TeamStackPermission(name, <any>undefined, { urn })
            case "pulumiservice:index:Webhook":
                return new Webhook(name, <any>undefined, { urn })
            default:
                throw new Error(`unknown resource type ${type}`);
        }
    },
};
pulumi.runtime.registerResourceModule("pulumiservice", "index", _module)
pulumi.runtime.registerResourcePackage("pulumiservice", {
    version: utilities.getVersion(),
    constructProvider: (name: string, type: string, urn: string): pulumi.ProviderResource => {
        if (type !== "pulumi:providers:pulumiservice") {
            throw new Error(`unknown provider type ${type}`);
        }
        return new Provider(name, <any>undefined, { urn });
    },
});
