// Package oauth wraps Google's OAuth2 web-server flow: building the
// consent-screen redirect URL, exchanging an auth code for a token, and
// fetching the signed-in user's basic profile (email, name, picture).
package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleUserInfo is the subset of Google's userinfo response this app uses.
// Full response also includes given_name/family_name/locale, which we don't need.
type GoogleUserInfo struct {
	ID            string `json:"id"` // stable Google subject ID
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// GoogleClient wraps an oauth2.Config configured for Google's web-server flow.
type GoogleClient struct {
	config *oauth2.Config
}

// NewGoogleClient builds a GoogleClient from the app's registered OAuth
// credentials. redirectURL must exactly match one of the "Authorized
// redirect URIs" configured in the Google Cloud Console for this client ID.
func NewGoogleClient(clientID, clientSecret, redirectURL string) *GoogleClient {
	return &GoogleClient{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

// AuthCodeURL returns the URL to redirect the user's browser to in order
// to start Google's consent flow. state is an opaque, unguessable value
// the caller generates (via utils.RandomToken) and verifies on callback
// against a short-lived cookie to prevent CSRF — see AuthHandler.GoogleLogin
// and AuthHandler.GoogleCallback.
func (g *GoogleClient) AuthCodeURL(state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// Exchange trades a one-time authorization code (received on the OAuth
// callback) for an access token.
func (g *GoogleClient) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging oauth code: %w", err)
	}
	return token, nil
}

// FetchUserInfo calls Google's userinfo endpoint with the given token to
// retrieve the signed-in user's email, name, and picture.
func (g *GoogleClient) FetchUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := g.config.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("fetching google userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading google userinfo response: %w", err)
	}

	var info GoogleUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parsing google userinfo response: %w", err)
	}

	return &info, nil
}
