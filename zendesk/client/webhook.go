package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// WebhookPatchAPI updates only the webhook fields it is given.
type WebhookPatchAPI interface {
	PatchWebhook(ctx context.Context, webhookID string, fields map[string]any) error
}

// PatchWebhook sends only the given fields, keyed by their API names. A field left out keeps its value. That is
// what an update needs, because the authentication value is write-only: Zendesk never returns it, so a full PUT
// can only resend it blank, which Zendesk rejects, or omit it, which removes the webhook's authentication.
// ref: https://developer.zendesk.com/api-reference/webhooks/webhooks-api/webhooks/#patch-webhook
func (z *Client) PatchWebhook(ctx context.Context, webhookID string, fields map[string]any) error {
	payload, err := json.Marshal(map[string]any{"webhook": fields})
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/webhooks/%s", z.baseURL, url.PathEscape(webhookID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if z.credential != nil {
		if z.credential.Bearer() {
			req.Header.Set("Authorization", "Bearer "+z.credential.Secret())
		} else {
			req.SetBasicAuth(z.credential.Email(), z.credential.Secret())
		}
	}

	resp, err := z.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("patching webhook %s: %w", webhookID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("patching webhook %s: %s: %s", webhookID, resp.Status, body)
	}

	return nil
}
