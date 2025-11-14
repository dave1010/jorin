package main

import "github.com/dave1010/jorin/internal/types"

// Type aliases to internal/types so cmd package can continue using the original
// type names. This preserves minimal changes to other files during the package
// split.

type Message = types.Message
type ToolCall = types.ToolCall
type Tool = types.Tool
type ToolFunction = types.ToolFunction
type ChatRequest = types.ChatRequest
type Choice = types.Choice
type ChatResponse = types.ChatResponse

type Policy = types.Policy
