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

package v20241105

import (
	"github.com/googleapis/genai-toolbox/internal/server/mcp/jsonrpc"
	"github.com/googleapis/genai-toolbox/internal/tools"
)

// SERVER_NAME is the server name used in Implementation.
const SERVER_NAME = "Toolbox"

// PROTOCOL_VERSION is the version of the MCP protocol in this package.
const PROTOCOL_VERSION = "2024-11-05"

// methods that are supported.
const (
	TOOLS_LIST = "tools/list"
	TOOLS_CALL = "tools/call"
)

/* Empty result */

// EmptyResult represents a response that indicates success but carries no data.
type EmptyResult jsonrpc.Result

/* Pagination */

// Cursor is an opaque token used to represent a cursor for pagination.
type Cursor string

type PaginatedRequest struct {
	jsonrpc.Request
	Params struct {
		// An opaque token representing the current pagination position.
		// If provided, the server should return results starting after this cursor.
		Cursor Cursor `json:"cursor,omitempty"`
	} `json:"params,omitempty"`
}

type PaginatedResult struct {
	jsonrpc.Result
	// An opaque token representing the pagination position after the last returned result.
	// If present, there may be more results available.
	NextCursor Cursor `json:"nextCursor,omitempty"`
}

/* Tools */

// Sent from the client to request a list of tools the server has.
type ListToolsRequest struct {
	PaginatedRequest
}

// The server's response to a tools/list request from the client.
type ListToolsResult struct {
	PaginatedResult
	Tools []tools.McpManifest `json:"tools"`
}

// Used by the client to invoke a tool provided by the server.
type CallToolRequest struct {
	jsonrpc.Request
	Params struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments,omitempty"`
	} `json:"params,omitempty"`
}

// The sender or recipient of messages and data in a conversation.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Base for objects that include optional annotations for the client.
// The client can use annotations to inform how objects are used or displayed
type Annotated struct {
	Annotations *struct {
		// Describes who the intended customer of this object or data is.
		// It can include multiple entries to indicate content useful for multiple
		// audiences (e.g., `["user", "assistant"]`).
		Audience []Role `json:"audience,omitempty"`
		// Describes how important this data is for operating the server.
		//
		// A value of 1 means "most important," and indicates that the data is
		// effectively required, while 0 means "least important," and indicates that
		// the data is entirely optional.
		//
		// @TJS-type number
		// @minimum 0
		// @maximum 1
		Priority float64 `json:"priority,omitempty"`
	} `json:"annotations,omitempty"`
}

// TextContent represents text provided to or from an LLM.
type TextContent struct {
	Annotated
	Type string `json:"type"`
	// The text content of the message.
	Text string `json:"text"`
}

// The server's response to a tool call.
//
// Any errors that originate from the tool SHOULD be reported inside the result
// object, with `isError` set to true, _not_ as an MCP protocol-level error
// response. Otherwise, the LLM would not be able to see that an error occurred
// and self-correct.
//
// However, any errors in _finding_ the tool, an error indicating that the
// server does not support tool calls, or any other exceptional conditions,
// should be reported as an MCP error response.
type CallToolResult struct {
	jsonrpc.Result
	// Could be either a TextContent, ImageContent, or EmbeddedResources
	// For Toolbox, we will only be sending TextContent
	Content []TextContent `json:"content"`
	// Whether the tool call ended in an error.
	// If not set, this is assumed to be false (the call was successful).
	IsError bool `json:"isError,omitempty"`
}
