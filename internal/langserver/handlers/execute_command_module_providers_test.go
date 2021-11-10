package handlers

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/creachadair/jrpc2/code"
	"github.com/hashicorp/terraform-ls/internal/langserver"
	"github.com/hashicorp/terraform-ls/internal/langserver/cmd"
	"github.com/hashicorp/terraform-ls/internal/state"
	"github.com/hashicorp/terraform-ls/internal/terraform/exec"
	"github.com/hashicorp/terraform-ls/internal/uri"
	"github.com/stretchr/testify/mock"
)

func TestLangServer_workspaceExecuteCommand_moduleProviders_argumentError(t *testing.T) {
	rootDir := t.TempDir()
	rootUri := uri.FromPath(rootDir)

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{
		TerraformCalls: &exec.TerraformMockCalls{
			PerWorkDir: map[string][]*mock.Call{
				rootDir: validTfMockCalls(),
			},
		},
	}))
	stop := ls.Start(t)
	defer stop()

	ls.Call(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
	    "capabilities": {},
	    "rootUri": %q,
		"processId": 12345
	}`, rootUri)})
	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})
	ls.Call(t, &langserver.CallRequest{
		Method: "textDocument/didOpen",
		ReqParams: fmt.Sprintf(`{
		"textDocument": {
			"version": 0,
			"languageId": "terraform",
			"text": "provider \"github\" {}",
			"uri": %q
		}
	}`, fmt.Sprintf("%s/main.tf", rootUri))})

	ls.CallAndExpectError(t, &langserver.CallRequest{
		Method: "workspace/executeCommand",
		ReqParams: fmt.Sprintf(`{
		"command": %q
	}`, cmd.Name("module.providers"))}, code.InvalidParams.Err())
}

func TestLangServer_workspaceExecuteCommand_moduleProviders_basic(t *testing.T) {
	rootDir := t.TempDir()
	rootUri := uri.FromPath(rootDir)
	baseDirUri := uri.FromPath(filepath.Join(rootDir, "base"))

	s, err := state.NewStateStore()
	if err != nil {
		t.Fatal(err)
	}

	modPath := t.TempDir()
	err = s.Modules.Add(modPath)
	if err != nil {
		t.Fatal(err)
	}

	// s.Modules.UpdateInstalledProviders()
	// s.Modules.UpdateMetadata()

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{
		TerraformCalls: &exec.TerraformMockCalls{
			PerWorkDir: map[string][]*mock.Call{
				rootDir: validTfMockCalls(),
			},
		},
		StateStore: s,
	}))
	stop := ls.Start(t)
	defer stop()

	ls.Call(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
	    "capabilities": {},
	    "rootUri": %q,
		"processId": 12345
	}`, rootUri)})
	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})
	ls.Call(t, &langserver.CallRequest{
		Method: "textDocument/didOpen",
		ReqParams: fmt.Sprintf(`{
		"textDocument": {
			"version": 0,
			"languageId": "terraform",
			"text": "provider \"github\" {}",
			"uri": %q
		}
	}`, fmt.Sprintf("%s/main.tf", baseDirUri))})

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method: "workspace/executeCommand",
		ReqParams: fmt.Sprintf(`{
		"command": %q,
		"arguments": ["uri=%s"]
	}`, cmd.Name("module.providers"), baseDirUri)}, fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"id": 3,
		"result": {
			"v": 0,
			"provider_requirements": {},
			"installed_providers": {}
		}
	}`))
}
