package prompt

import "strings"

// PromptProvider is an extensible provider of parts of the system prompt.
// Additional providers can be registered (for example by plugins) to append
// more context or instructions to the overall system prompt.
type PromptProvider interface {
	// Provide returns the text to include in the system prompt. Empty string
	// means nothing will be added for this provider.
	Provide() string
}

var promptProviders []PromptProvider

// RegisterPromptProvider registers a PromptProvider. Providers are iterated in
// registration order when building the system prompt.
func RegisterPromptProvider(p PromptProvider) {
	promptProviders = append(promptProviders, p)
}

// SystemPrompt builds the full system prompt by concatenating the outputs of
// all registered PromptProviders. The immutable baseProvider is always placed
// first regardless of registration order so core instructions appear first.
func SystemPrompt() string {
	parts := []string{}
	// include any baseProvider content first
	for _, p := range promptProviders {
		if _, ok := p.(baseProvider); ok {
			if s := p.Provide(); s != "" {
				parts = append(parts, s)
			}
		}
	}
	// then include all non-base providers in registration order
	for _, p := range promptProviders {
		if _, ok := p.(baseProvider); ok {
			continue
		}
		if s := p.Provide(); s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, "\n\n")
}
