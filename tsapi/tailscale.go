package tsapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/clientcredentials"
)

var BASE_URL = "https://api.tailscale.com/api/v2/"

type TSApi struct {
	httpClient  *http.Client
	tailnetName string
}

type getDevicesResponse struct {
	Devices []Device `json:"devices"`
}

type Device struct {
	Addresses []string `json:"addresses"`
	Hostname  string   `json:"hostname"`
	Name      string   `json:"name"`
	NodeId    string   `json:"nodeId"`
}

func NewTSClient(tailnetName string) *TSApi {
	var oauthConfig = &clientcredentials.Config{
		ClientID:     viper.GetString("tailscale.oauth-client-id"),
		ClientSecret: viper.GetString("tailscale.oauth-client-secret"),
		TokenURL:     buildPath("/oauth/token"),
	}

	client := oauthConfig.Client(context.Background())
	tsapi := &TSApi{httpClient: client, tailnetName: tailnetName}

	return tsapi
}

func (ts *TSApi) Devices() ([]Device, error) {
	// Make the request
	url := buildPath("/tailnet", ts.tailnetName, "/devices")
	log.Trace().Str("url", url).Msg("Sending GET /devices")
	resp, err := ts.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("cannot get tailnet devices: %w", err)
	}

	// Read the body into a slice
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read devices response: %w", err)
	}

	// Parse the body to JSON
	var deviceResponse getDevicesResponse
	log.Trace().Int("bytes", len(body)).Msg("Parsing response")
	err = json.Unmarshal(body, &deviceResponse)
	if err != nil {
		return nil, fmt.Errorf("cannot parse devices json: %w", err)
	}

	return deviceResponse.Devices, nil
}

func buildPath(elem ...string) string {
	result, err := url.JoinPath(BASE_URL, elem...)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	return result
}
