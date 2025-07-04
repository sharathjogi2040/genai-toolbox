// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package alloydbainl

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/googleapis/genai-toolbox/internal/server/mcp/jsonrpc"
	"github.com/googleapis/genai-toolbox/tests"
)

var (
	AlloyDBAINLSourceKind = "alloydb-postgres"
	AlloyDBAINLToolKind   = "alloydb-ai-nl"
	AlloyDBAINLProject    = os.Getenv("ALLOYDB_AI_NL_PROJECT")
	AlloyDBAINLRegion     = os.Getenv("ALLOYDB_AI_NL_REGION")
	AlloyDBAINLCluster    = os.Getenv("ALLOYDB_AI_NL_CLUSTER")
	AlloyDBAINLInstance   = os.Getenv("ALLOYDB_AI_NL_INSTANCE")
	AlloyDBAINLDatabase   = os.Getenv("ALLOYDB_AI_NL_DATABASE")
	AlloyDBAINLUser       = os.Getenv("ALLOYDB_AI_NL_USER")
	AlloyDBAINLPass       = os.Getenv("ALLOYDB_AI_NL_PASS")
)

func getAlloyDBAINLVars(t *testing.T) map[string]any {
	switch "" {
	case AlloyDBAINLProject:
		t.Fatal("'ALLOYDB_AI_NL_PROJECT' not set")
	case AlloyDBAINLRegion:
		t.Fatal("'ALLOYDB_AI_NL_REGION' not set")
	case AlloyDBAINLCluster:
		t.Fatal("'ALLOYDB_AI_NL_CLUSTER' not set")
	case AlloyDBAINLInstance:
		t.Fatal("'ALLOYDB_AI_NL_INSTANCE' not set")
	case AlloyDBAINLDatabase:
		t.Fatal("'ALLOYDB_AI_NL_DATABASE' not set")
	case AlloyDBAINLUser:
		t.Fatal("'ALLOYDB_AI_NL_USER' not set")
	case AlloyDBAINLPass:
		t.Fatal("'ALLOYDB_AI_NL_PASS' not set")
	}
	return map[string]any{
		"kind":     AlloyDBAINLSourceKind,
		"project":  AlloyDBAINLProject,
		"cluster":  AlloyDBAINLCluster,
		"instance": AlloyDBAINLInstance,
		"region":   AlloyDBAINLRegion,
		"database": AlloyDBAINLDatabase,
		"user":     AlloyDBAINLUser,
		"password": AlloyDBAINLPass,
	}
}

func TestAlloyDBAINLToolEndpoints(t *testing.T) {
	sourceConfig := getAlloyDBAINLVars(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var args []string

	// Write config into a file and pass it to command
	toolsFile := getAINLToolsConfig(sourceConfig)

	cmd, cleanup, err := tests.StartCmd(ctx, toolsFile, args...)
	if err != nil {
		t.Fatalf("command initialization returned an error: %s", err)
	}
	defer cleanup()

	waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	out, err := cmd.WaitForString(waitCtx, regexp.MustCompile(`Server ready to serve`))
	if err != nil {
		t.Logf("toolbox command logs: \n%s", out)
		t.Fatalf("toolbox didn't start successfully: %s", err)
	}

	runAINLToolGetTest(t)
	runAINLToolInvokeTest(t)
	runAINLMCPToolCallMethod(t)
}

func runAINLToolGetTest(t *testing.T) {
	// Test tool get endpoint
	tcs := []struct {
		name string
		api  string
		want map[string]any
	}{
		{
			name: "get my-simple-tool",
			api:  "http://127.0.0.1:5000/api/tool/my-simple-tool/",
			want: map[string]any{
				"my-simple-tool": map[string]any{
					"description": "Simple tool to test end to end functionality.",
					"parameters": []any{
						map[string]any{
							"name":        "question",
							"type":        "string",
							"required":    true,
							"description": "The natural language question to ask.",
							"authSources": []any{},
						},
					},
					"authRequired": []any{},
				},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(tc.api)
			if err != nil {
				t.Fatalf("error when sending a request: %s", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				t.Fatalf("response status code is not 200")
			}

			var body map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&body)
			if err != nil {
				t.Fatalf("error parsing response body")
			}

			got, ok := body["tools"]
			if !ok {
				t.Fatalf("unable to find tools in response body")
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func runAINLToolInvokeTest(t *testing.T) {
	// Get ID token
	idToken, err := tests.GetGoogleIdToken(tests.ClientId)
	if err != nil {
		t.Fatalf("error getting Google ID token: %s", err)
	}

	// Test tool invoke endpoint
	invokeTcs := []struct {
		name          string
		api           string
		requestHeader map[string]string
		requestBody   io.Reader
		want          string
		isErr         bool
	}{
		{
			name:          "invoke my-simple-tool",
			api:           "http://127.0.0.1:5000/api/tool/my-simple-tool/invoke",
			requestHeader: map[string]string{},
			requestBody:   bytes.NewBuffer([]byte(`{"question": "return 1"}`)),
			want:          "[{\"execute_nl_query\":{\"?column?\":1}}]",
			isErr:         false,
		},
		{
			name:          "Invoke my-tool without parameters",
			api:           "http://127.0.0.1:5000/api/tool/my-tool/invoke",
			requestHeader: map[string]string{},
			requestBody:   bytes.NewBuffer([]byte(`{}`)),
			isErr:         true,
		},
		{
			name:          "Invoke my-auth-tool with auth token",
			api:           "http://127.0.0.1:5000/api/tool/my-auth-tool/invoke",
			requestHeader: map[string]string{"my-google-auth_token": idToken},
			requestBody:   bytes.NewBuffer([]byte(`{"question": "can you show me the name of this user?"}`)),
			want:          "[{\"execute_nl_query\":{\"name\":\"Alice\"}}]",
			isErr:         false,
		},
		{
			name:          "Invoke my-auth-tool with invalid auth token",
			api:           "http://127.0.0.1:5000/api/tool/my-auth-tool/invoke",
			requestHeader: map[string]string{"my-google-auth_token": "INVALID_TOKEN"},
			requestBody:   bytes.NewBuffer([]byte(`{"question": "return 1"}`)),
			isErr:         true,
		},
		{
			name:          "Invoke my-auth-tool without auth token",
			api:           "http://127.0.0.1:5000/api/tool/my-auth-tool/invoke",
			requestHeader: map[string]string{},
			requestBody:   bytes.NewBuffer([]byte(`{"question": "return 1"}`)),
			isErr:         true,
		},
		{
			name:          "Invoke my-auth-required-tool with auth token",
			api:           "http://127.0.0.1:5000/api/tool/my-auth-required-tool/invoke",
			requestHeader: map[string]string{"my-google-auth_token": idToken},
			requestBody:   bytes.NewBuffer([]byte(`{"question": "return 1"}`)),
			isErr:         false,
			want:          "[{\"execute_nl_query\":{\"?column?\":1}}]",
		},
		{
			name:          "Invoke my-auth-required-tool with invalid auth token",
			api:           "http://127.0.0.1:5000/api/tool/my-auth-required-tool/invoke",
			requestHeader: map[string]string{"my-google-auth_token": "INVALID_TOKEN"},
			requestBody:   bytes.NewBuffer([]byte(`{"question": "return 1"}`)),
			isErr:         true,
		},
		{
			name:          "Invoke my-auth-required-tool without auth token",
			api:           "http://127.0.0.1:5000/api/tool/my-auth-tool/invoke",
			requestHeader: map[string]string{},
			requestBody:   bytes.NewBuffer([]byte(`{"question": "return 1"}`)),
			isErr:         true,
		},
	}
	for _, tc := range invokeTcs {
		t.Run(tc.name, func(t *testing.T) {
			// Send Tool invocation request
			req, err := http.NewRequest(http.MethodPost, tc.api, tc.requestBody)
			if err != nil {
				t.Fatalf("unable to create request: %s", err)
			}
			req.Header.Add("Content-type", "application/json")
			for k, v := range tc.requestHeader {
				req.Header.Add(k, v)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("unable to send request: %s", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				if tc.isErr == true {
					return
				}
				bodyBytes, _ := io.ReadAll(resp.Body)
				t.Fatalf("response status code is not 200, got %d: %s", resp.StatusCode, string(bodyBytes))
			}

			// Check response body
			var body map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&body)
			if err != nil {
				t.Fatalf("error parsing response body")
			}
			got, ok := body["result"].(string)
			if !ok {
				t.Fatalf("unable to find result in response body")
			}

			if got != tc.want {
				t.Fatalf("unexpected value: got %q, want %q", got, tc.want)
			}
		})
	}

}

func getAINLToolsConfig(sourceConfig map[string]any) map[string]any {
	// Write config into a file and pass it to command
	toolsFile := map[string]any{
		"sources": map[string]any{
			"my-instance": sourceConfig,
		},
		"authServices": map[string]any{
			"my-google-auth": map[string]any{
				"kind":     "google",
				"clientId": tests.ClientId,
			},
		},
		"tools": map[string]any{
			"my-simple-tool": map[string]any{
				"kind":        AlloyDBAINLToolKind,
				"source":      "my-instance",
				"description": "Simple tool to test end to end functionality.",
				"nlConfig":    "my_nl_config",
			},
			"my-auth-tool": map[string]any{
				"kind":        AlloyDBAINLToolKind,
				"source":      "my-instance",
				"description": "Tool to test authenticated parameters.",
				"nlConfig":    "my_nl_config",
				"nlConfigParameters": []map[string]any{
					{
						"name":        "email",
						"type":        "string",
						"description": "user email",
						"authServices": []map[string]string{
							{
								"name":  "my-google-auth",
								"field": "email",
							},
						},
					},
				},
			},
			"my-auth-required-tool": map[string]any{
				"kind":        AlloyDBAINLToolKind,
				"source":      "my-instance",
				"description": "Tool to test auth required invocation.",
				"nlConfig":    "my_nl_config",
				"authRequired": []string{
					"my-google-auth",
				},
			},
		},
	}

	return toolsFile
}

func runAINLMCPToolCallMethod(t *testing.T) {
	sessionId := tests.RunInitialize(t, "2024-11-05")
	header := map[string]string{}
	if sessionId != "" {
		header["Mcp-Session-Id"] = sessionId
	}

	// Test tool invoke endpoint
	invokeTcs := []struct {
		name          string
		api           string
		requestBody   jsonrpc.JSONRPCRequest
		requestHeader map[string]string
		want          string
	}{
		{
			name:          "MCP Invoke my-simple-tool",
			api:           "http://127.0.0.1:5000/mcp",
			requestHeader: map[string]string{},
			requestBody: jsonrpc.JSONRPCRequest{
				Jsonrpc: "2.0",
				Id:      "my-simple-tool",
				Request: jsonrpc.Request{
					Method: "tools/call",
				},
				Params: map[string]any{
					"name": "my-simple-tool",
					"arguments": map[string]any{
						"question": "return 1",
					},
				},
			},
			want: `{"jsonrpc":"2.0","id":"my-simple-tool","result":{"content":[{"type":"text","text":"{\"execute_nl_query\":{\"?column?\":1}}"}]}}`,
		},
		{
			name:          "MCP Invoke invalid tool",
			api:           "http://127.0.0.1:5000/mcp",
			requestHeader: map[string]string{},
			requestBody: jsonrpc.JSONRPCRequest{
				Jsonrpc: "2.0",
				Id:      "invalid-tool",
				Request: jsonrpc.Request{
					Method: "tools/call",
				},
				Params: map[string]any{
					"name":      "foo",
					"arguments": map[string]any{},
				},
			},
			want: `{"jsonrpc":"2.0","id":"invalid-tool","error":{"code":-32602,"message":"invalid tool name: tool with name \"foo\" does not exist"}}`,
		},
		{
			name:          "MCP Invoke my-auth-tool without parameters",
			api:           "http://127.0.0.1:5000/mcp",
			requestHeader: map[string]string{},
			requestBody: jsonrpc.JSONRPCRequest{
				Jsonrpc: "2.0",
				Id:      "invoke-without-parameter",
				Request: jsonrpc.Request{
					Method: "tools/call",
				},
				Params: map[string]any{
					"name":      "my-auth-tool",
					"arguments": map[string]any{},
				},
			},
			want: `{"jsonrpc":"2.0","id":"invoke-without-parameter","error":{"code":-32602,"message":"provided parameters were invalid: parameter question is required"}}`,
		},
	}
	for _, tc := range invokeTcs {
		t.Run(tc.name, func(t *testing.T) {
			reqMarshal, err := json.Marshal(tc.requestBody)
			if err != nil {
				t.Fatalf("unexpected error during marshaling of request body")
			}
			// Send Tool invocation request
			req, err := http.NewRequest(http.MethodPost, tc.api, bytes.NewBuffer(reqMarshal))
			if err != nil {
				t.Fatalf("unable to create request: %s", err)
			}
			req.Header.Add("Content-type", "application/json")
			for k, v := range header {
				req.Header.Add(k, v)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("unable to send request: %s", err)
			}
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("unable to read request body: %s", err)
			}
			defer resp.Body.Close()
			got := string(bytes.TrimSpace(respBody))

			// Remove `\` and `"` for string comparison
			got = strings.ReplaceAll(got, "\\", "")
			want := strings.ReplaceAll(tc.want, "\\", "")
			got = strings.ReplaceAll(got, "\"", "")
			want = strings.ReplaceAll(want, "\"", "")

			if !strings.Contains(got, want) {
				t.Fatalf("Expected substring not found:\ngot:  %q\nwant: %q (to be contained within got)", got, want)
			}
		})
	}
}
