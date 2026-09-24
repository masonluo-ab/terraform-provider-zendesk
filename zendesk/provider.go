package zendesk

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	client "github.com/nukosuke/go-zendesk/zendesk"
	newClient "github.com/nukosuke/terraform-provider-zendesk/zendesk/client"
)

const (
	accountVar           = "ZENDESK_ACCOUNT"
	emailVar             = "ZENDESK_EMAIL"
	tokenVar             = "ZENDESK_TOKEN"
	oauthClientIDVar     = "ZENDESK_OAUTH_CLIENT_ID"
	oauthClientSecretVar = "ZENDESK_OAUTH_CLIENT_SECRET"
	oauthScopeVar        = "ZENDESK_OAUTH_SCOPE"
)

// Provider returns provider instance for Zendesk
func Provider() *schema.Provider {
	return &schema.Provider{
		// https://developer.zendesk.com/rest_api/docs/support/introduction#security-and-authentication
		Schema: map[string]*schema.Schema{
			"account": {
				Description:  "Account name of your Zendesk instance.",
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc(accountVar, ""),
				ValidateFunc: validation.StringIsNotEmpty,
			},
			// The credentials default to nil rather than "" when their variable is unset, so that configuring
			// one kind leaves the other absent instead of set to an empty string that fails validation.
			"email": {
				Description:  "Email address of agent user who have permission to access the API. Set with `token`; conflicts with the OAuth arguments.",
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc(emailVar, nil),
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"token": {
				Description:  "[API token](https://developer.zendesk.com/rest_api/docs/support/introduction#api-token) for your Zendesk instance. Set with `email`; conflicts with the OAuth arguments.",
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc(tokenVar, nil),
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"oauth_client_id": {
				Description:  "Identifier of a confidential OAuth client to authenticate as, with the client_credentials grant. Set with `oauth_client_secret`; conflicts with `email` and `token`. Requests then act as the user who owns the client.",
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc(oauthClientIDVar, nil),
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"oauth_client_secret": {
				Description:  "Secret of the OAuth client named by `oauth_client_id`.",
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc(oauthClientSecretVar, nil),
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"oauth_scope": {
				Description:  "Space-separated scopes to request with the OAuth client, e.g. `read write`. Left out of the request when unset.",
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc(oauthScopeVar, nil),
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"zendesk_automation":                resourceZendeskAutomation(),
			"zendesk_brand":                     resourceZendeskBrand(),
			"zendesk_dynamic_content":           resourceZendeskDynamicContent(),
			"zendesk_dynamic_content_variant":   resourceZendeskDynamicContentVariant(),
			"zendesk_group":                     resourceZendeskGroup(),
			"zendesk_ticket_field":              resourceZendeskTicketField(),
			"zendesk_macro":                     resourceZendeskMacro(),
			"zendesk_view":                      resourceZendeskView(),
			"zendesk_user_field":                resourceZendeskUserField(),
			"zendesk_ticket_form":               resourceZendeskTicketForm(),
			"zendesk_trigger":                   resourceZendeskTrigger(),
			"zendesk_trigger_category":           resourceZendeskTriggerCategory(),
			"zendesk_target":                    resourceZendeskTarget(),
			"zendesk_attachment":                resourceZendeskAttachment(),
			"zendesk_oauth_client":              resourceZendeskOAuthClient(),
			"zendesk_organization":              resourceZendeskOrganization(),
			"zendesk_organization_field":        resourceZendeskOrganizationField(),
			"zendesk_sla_policy":                resourceZendeskSLAPolicy(),
			"zendesk_webhook":                   resourceZendeskWebhook(),
			"zendesk_custom_roles":              resourceZendeskCustomRoles(),
			"zendesk_custom_statuses":           resourceZendeskCustomStatuses(),
			"zendesk_group_memberships":          resourceZendeskGroupMemberships(),
			"zendesk_organization_memberships": resourceZendeskOrganizationMemberships(),
			"zendesk_users":                     resourceZendeskUsers(),
			"zendesk_tickets":                   resourceZendeskTickets(),
			"zendesk_queues":                    resourceZendeskQueues(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"zendesk_ticket_field":          dataSourceZendeskTicketField(),
			"zendesk_webhook":               dataSourceZendeskWebhook(),
			"zendesk_tags":                  dataSourceZendeskTags(),
			"zendesk_locales":               dataSourceZendeskLocales(),
			"zendesk_satisfaction_ratings": dataSourceZendeskSatisfactionRatings(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	config := Config{
		Account:           d.Get("account").(string),
		Email:             d.Get("email").(string),
		Token:             d.Get("token").(string),
		OAuthClientID:     d.Get("oauth_client_id").(string),
		OAuthClientSecret: d.Get("oauth_client_secret").(string),
		OAuthScope:        d.Get("oauth_scope").(string),
	}

	oauth, err := config.usesOAuth()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	// Create & configure Zendesk API client
	zd, err := client.NewClient(nil) // TODO: set UserAgent to terraform/version
	if err != nil {
		return nil, diag.FromErr(err)
	}

	if err = zd.SetSubdomain(config.Account); err != nil {
		return nil, diag.FromErr(err)
	}

	if oauth {
		// Exchanged once per run. The token lasts about 30 minutes and is not renewed, so a run longer than
		// that fails partway through.
		instanceURL := fmt.Sprintf("https://%s.zendesk.com", config.Account)
		httpClient := &http.Client{Timeout: 30 * time.Second}
		token, err := newClient.ExchangeClientCredentials(ctx, httpClient, instanceURL, config.OAuthClientID, config.OAuthClientSecret, config.OAuthScope)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		zd.SetCredential(client.NewBearerTokenCredential(token))
	} else {
		zd.SetCredential(client.NewAPITokenCredential(config.Email, config.Token))
	}

	newZd := &newClient.Client{
		Client: *zd,
	}
	return newZd, diags
}
