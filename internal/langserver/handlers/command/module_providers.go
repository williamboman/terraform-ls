package command

import (
	"context"
	"fmt"

	"github.com/creachadair/jrpc2/code"
	"github.com/hashicorp/terraform-ls/internal/langserver/cmd"
	"github.com/hashicorp/terraform-ls/internal/uri"
)

const moduleProvidersVersion = 0

type moduleProvidersResponse struct {
	FormatVersion   int              `json:"v"`
	ModuleProviders []moduleProvider `json:"module_providers"`
}

type moduleProvider struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

func ModuleProvidersHandler(ctx context.Context, args cmd.CommandArgs) (interface{}, error) {
	response := moduleProvidersResponse{
		FormatVersion:   moduleProvidersVersion,
		ModuleProviders: make([]moduleProvider, 0),
	}

	modUri, ok := args.GetString("uri")
	if !ok || modUri == "" {
		return response, fmt.Errorf("%w: expected module uri argument to be set", code.InvalidParams.Err())
	}

	if !uri.IsURIValid(modUri) {
		return response, fmt.Errorf("URI %q is not valid", modUri)
	}

	_, err := uri.PathFromURI(modUri)
	if err != nil {
		return response, err
	}

	response.ModuleProviders = []moduleProvider{
		{
			Name:    "hashcorp/aws",
			Version: "3.64.1",
		},
		{
			Name:    "hashicorp/google",
			Version: "3.90.1",
		},
	}

	return response, nil
}
