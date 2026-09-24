# API document about authentication:
#   https://developer.zendesk.com/api-reference/introduction/security-and-auth/

terraform {
  required_providers {
    zendesk = {
      source  = "masonluo-ab/zendesk"
      version = ">= 1.2.0"
    }
  }
}

# Authenticate with an API token.
provider "zendesk" {
  # example.zendesk.com
  account = "example"
  email   = "john.doe@example.com"
  token   = "xxxxxxxxxx"

  # or configure from environment variables
  # if you don't want to hardcode the credentials.
  #
  # export ZENDESK_ACCOUNT="example"
  # export ZENDESK_EMAIL="john.doe@example.com"
  # export ZENDESK_TOKEN="xxxxxxxxxx"
}

# Or authenticate with a confidential OAuth client, using the client_credentials grant. The provider exchanges
# the secret for an access token once per run; the token lasts about 30 minutes and is not renewed.
#
# Requests act as the user who owns the OAuth client, so that user needs the permissions the configuration
# exercises. The client cannot be managed by the same configuration it authenticates: replacing it would
# revoke the credentials Terraform is running with.
provider "zendesk" {
  alias = "oauth"

  account             = "example"
  oauth_client_id     = "terraform"
  oauth_client_secret = "xxxxxxxxxx"
  oauth_scope         = "read write"

  # export ZENDESK_OAUTH_CLIENT_ID="terraform"
  # export ZENDESK_OAUTH_CLIENT_SECRET="xxxxxxxxxx"
  # export ZENDESK_OAUTH_SCOPE="read write"
}
