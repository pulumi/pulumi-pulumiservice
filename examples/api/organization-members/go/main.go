package main

import (
	api "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice/api"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		organizationName := cfg.Get("organizationName")
		if organizationName == "" {
			organizationName = "service-provider-test-org"
		}
		memberLogin := cfg.Get("memberLogin")
		if memberLogin == "" {
			memberLogin = "pulumi-bot"
		}
		memberRole := cfg.Get("memberRole")
		if memberRole == "" {
			memberRole = "member"
		}

		member, err := api.NewOrganizationMember(ctx, "member", &api.OrganizationMemberArgs{
			OrgName:   pulumi.String(organizationName),
			UserLogin: pulumi.String(memberLogin),
			Role:      pulumi.String(memberRole),
		})
		if err != nil {
			return err
		}

		ctx.Export("memberId", member.ID())
		return nil
	})
}
