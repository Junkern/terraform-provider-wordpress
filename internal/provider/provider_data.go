package provider

import (
	"fmt"

	"terraform-provider-wordpress/internal/wpapi"
	"terraform-provider-wordpress/internal/wpappauth"
)

func appClientForProviderData(input any) (*wpapi.Client, error) {
	data, ok := input.(*providerData)
	if !ok {
		return nil, fmt.Errorf("expected provider data type *providerData, got: %T", input)
	}

	if data.AppClient == nil {
		return nil, fmt.Errorf("this resource requires a configured host and app_auth credentials")
	}

	return data.AppClient, nil
}

func publicClientForProviderData(input any) (*wpapi.Client, error) {
	if _, ok := input.(*providerData); !ok {
		return nil, fmt.Errorf("expected provider data type *providerData, got: %T", input)
	}

	return &wpapi.Client{}, nil
}

func userClientForProviderData(input any) (*wpappauth.Service, error) {
	data, ok := input.(*providerData)
	if !ok {
		return nil, fmt.Errorf("expected provider data type *providerData, got: %T", input)
	}

	if data.UserClient == nil {
		return nil, fmt.Errorf("this resource requires a configured host and user_auth credentials")
	}

	return data.UserClient, nil
}
