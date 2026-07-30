resource role "pulumiservice:api:Role" {
  orgName = "my-org"
  name = "platform-team-role"
  uxPurpose = "set"
  resourceType = "global"
  details = {
    __type = "PermissionDescriptorAllow"
    permissions = ["organization:read_usage"]
  }
}
