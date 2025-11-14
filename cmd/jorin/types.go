package main

import "github.com/dave1010/jorin/internal/types"

// Replace type aliases with direct imports: re-export types from internal/types
// so cmd package references the internal types explicitly.

// Use the internal types directly where needed in other files. This file will
// be removed once all references are updated.

// Deprecated: keep short compatibility shim. Prefer importing internal/types directly.

type Policy = types.Policy
