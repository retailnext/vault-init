# vault-init
When a new hashicorp vault cluster starts, it needs to be initialized. The code handles the initialization and some tasks after the initialization

## Vault initialization for vault managed by Terraform Cloud

`vault-init` initializes the vault in the given address and
saves the output to a gcp/aws secret or file. **Currently it does
NOT handle unseal process and it assumes that auto unseal is
implemented already (usually through KMS).**
After the intialization, with the initial root token, `vault-init` 
can perform the following tasks:
 - Set up policies; in order for the authentication to work properly,
  policies need to be set. Typically, `admin` policy can be set
  through this task.
 - Set up `jwt` type auth for oidc; oidc configuration and the initial
  role can be set up. Typically, `admin` role is set up with the policy
  created in the previous "policy task". For example, the role of
  terraform agent and workspace for vault ACL can be set up through
  this task. Refer to https://developer.hashicorp.com/terraform/cloud-docs/workspaces/dynamic-provider-credentials/vault-configuration for details
- Set up `gcp` type auth; refer to https://developer.hashicorp.com/vault/docs/auth/gcp
