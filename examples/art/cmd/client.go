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

func getClient(ctx context.Context) (*agrirouter.Client, error) {
	slog.Debug("Creating OAuth2 client using client credentials",
		slog.String("client_id", viper.GetString("AGRIROUTER_OAUTH_CLIENT_ID")),
		slog.String("token_url", "https://oauth.qa.agrirouter.farm/token"),
	)

	clientCredsConfig := clientcredentials.Config{
		ClientID:     viper.GetString("AGRIROUTER_OAUTH_CLIENT_ID"),
		ClientSecret: viper.GetString("AGRIROUTER_OAUTH_CLIENT_SECRET"),
		TokenURL:     "https://oauth.qa.agrirouter.farm/token",
	}

	tokenSource := clientCredsConfig.TokenSource(context.Background())
	_, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	httpClient := oauth2.NewClient(ctx, tokenSource)
	client, err := agrirouter.NewClient(
		"https://api.qa.agrirouter.farm",
		agrirouter.WithHTTPClient(httpClient),
	)
	if err != nil {
		log.Fatalf("Failed to create agrirouter client: %v", err)
	}
	return client, nil
}
