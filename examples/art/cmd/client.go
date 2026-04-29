package cmd

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	defaultAPIURL        = "https://api.qa.agrirouter.farm"
	defaultOAuthTokenURL = "https://oauth.qa.agrirouter.farm/token"
)

func getClient(ctx context.Context) (*agrirouter.Client, error) {
	apiURL := viper.GetString("ART_API_URL")
	if apiURL == "" {
		apiURL = defaultAPIURL
	}
	tokenURL := viper.GetString("ART_OAUTH_TOKEN_URL")
	if tokenURL == "" {
		tokenURL = defaultOAuthTokenURL
	}

	slog.Debug("Creating OAuth2 client using client credentials",
		slog.String("client_id", viper.GetString("AGRIROUTER_OAUTH_CLIENT_ID")),
		slog.String("token_url", tokenURL),
		slog.String("api_url", apiURL),
	)

	clientCredsConfig := clientcredentials.Config{
		ClientID:     viper.GetString("AGRIROUTER_OAUTH_CLIENT_ID"),
		ClientSecret: viper.GetString("AGRIROUTER_OAUTH_CLIENT_SECRET"),
		TokenURL:     tokenURL,
	}

	tokenSource := clientCredsConfig.TokenSource(context.Background())
	_, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	httpClient := oauth2.NewClient(ctx, tokenSource)
	client, err := agrirouter.NewClient(
		apiURL,
		agrirouter.WithHTTPClient(httpClient),
	)
	if err != nil {
		log.Fatalf("Failed to create agrirouter client: %v", err)
	}
	return client, nil
}
