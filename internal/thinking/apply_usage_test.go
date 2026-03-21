package thinking_test

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/thinking"
)

// TestThinkingLevelForUsage_ProcessedBody tests extraction from request bodies
// that have been processed by ApplyThinking. The body already contains the
// final thinking configuration (including suffix-converted values if applicable).
func TestThinkingLevelForUsage_ProcessedBody(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		model       string
		toFormat    string
		providerKey string
		want        string
	}{
		// Claude format - budget_tokens
		{
			name:        "claude budget_tokens enabled",
			body:        `{"thinking":{"type":"enabled","budget_tokens":8192}}`,
			model:       "claude-sonnet-4-5",
			toFormat:    "claude",
			providerKey: "claude",
			want:        "8192",
		},
		{
			name:        "claude disabled",
			body:        `{"thinking":{"type":"disabled"}}`,
			model:       "claude-sonnet-4-5",
			toFormat:    "claude",
			providerKey: "claude",
			want:        "none",
		},
		{
			name:        "claude budget_tokens -1 (auto)",
			body:        `{"thinking":{"budget_tokens":-1}}`,
			model:       "claude-sonnet-4-5",
			toFormat:    "claude",
			providerKey: "claude",
			want:        "auto",
		},
		{
			name:        "claude adaptive effort",
			body:        `{"thinking":{"type":"adaptive"},"output_config":{"effort":"high"}}`,
			model:       "claude-sonnet-4-5",
			toFormat:    "claude",
			providerKey: "claude",
			want:        "high",
		},
		// Gemini format
		{
			name:        "gemini thinkingLevel",
			body:        `{"generationConfig":{"thinkingConfig":{"thinkingLevel":"high"}}}`,
			model:       "gemini-2.5-pro",
			toFormat:    "gemini",
			providerKey: "gemini",
			want:        "high",
		},
		{
			name:        "gemini thinkingBudget",
			body:        `{"generationConfig":{"thinkingConfig":{"thinkingBudget":8192}}}`,
			model:       "gemini-2.5-pro",
			toFormat:    "gemini",
			providerKey: "gemini",
			want:        "8192",
		},
		// Gemini CLI format
		{
			name:        "gemini-cli thinkingLevel with prefix",
			body:        `{"request":{"generationConfig":{"thinkingConfig":{"thinkingLevel":"medium"}}}}`,
			model:       "gemini-2.5-pro",
			toFormat:    "gemini-cli",
			providerKey: "gemini-cli",
			want:        "medium",
		},
		// OpenAI format
		{
			name:        "openai reasoning_effort high",
			body:        `{"reasoning_effort":"high"}`,
			model:       "gpt-5",
			toFormat:    "openai",
			providerKey: "openai",
			want:        "high",
		},
		{
			name:        "openai reasoning_effort medium",
			body:        `{"reasoning_effort":"medium"}`,
			model:       "gpt-5",
			toFormat:    "openai",
			providerKey: "openai",
			want:        "medium",
		},
		{
			name:        "openai reasoning_effort low",
			body:        `{"reasoning_effort":"low"}`,
			model:       "gpt-5",
			toFormat:    "openai",
			providerKey: "openai",
			want:        "low",
		},
		// Codex / OpenAI Response format
		{
			name:        "codex reasoning.effort",
			body:        `{"reasoning":{"effort":"high"}}`,
			model:       "gpt-5-codex",
			toFormat:    "codex",
			providerKey: "codex",
			want:        "high",
		},
		{
			name:        "openai-response format",
			body:        `{"reasoning":{"effort":"medium"}}`,
			model:       "gpt-5",
			toFormat:    "openai-response",
			providerKey: "codex",
			want:        "medium",
		},
		// Antigravity format
		{
			name:        "antigravity format",
			body:        `{"request":{"generationConfig":{"thinkingConfig":{"thinkingLevel":"high"}}}}`,
			model:       "gemini-2.5-pro",
			toFormat:    "antigravity",
			providerKey: "antigravity",
			want:        "high",
		},
		// Kimi format (OpenAI compatible)
		{
			name:        "kimi format",
			body:        `{"reasoning_effort":"high"}`,
			model:       "kimi-model",
			toFormat:    "kimi",
			providerKey: "kimi",
			want:        "high",
		},
		// iFlow format - GLM enable_thinking
		{
			name:        "iflow enable_thinking true",
			body:        `{"chat_template_kwargs":{"enable_thinking":true}}`,
			model:       "glm-4-plus",
			toFormat:    "iflow",
			providerKey: "iflow",
			want:        "1",
		},
		{
			name:        "iflow enable_thinking false",
			body:        `{"chat_template_kwargs":{"enable_thinking":false}}`,
			model:       "glm-4-plus",
			toFormat:    "iflow",
			providerKey: "iflow",
			want:        "none",
		},
		// iFlow format - MiniMax reasoning_split
		{
			name:        "iflow reasoning_split true",
			body:        `{"reasoning_split":true}`,
			model:       "minimax-text-01",
			toFormat:    "iflow",
			providerKey: "iflow",
			want:        "1",
		},
		{
			name:        "iflow reasoning_split false",
			body:        `{"reasoning_split":false}`,
			model:       "minimax-text-01",
			toFormat:    "iflow",
			providerKey: "iflow",
			want:        "none",
		},
		// Gemini snake_case format (Google Python SDK)
		{
			name:        "gemini thinking_level snake_case",
			body:        `{"generationConfig":{"thinkingConfig":{"thinking_level":"high"}}}`,
			model:       "gemini-2.5-pro",
			toFormat:    "gemini",
			providerKey: "gemini",
			want:        "high",
		},
		{
			name:        "gemini thinking_budget snake_case",
			body:        `{"generationConfig":{"thinkingConfig":{"thinking_budget":4096}}}`,
			model:       "gemini-2.5-pro",
			toFormat:    "gemini",
			providerKey: "gemini",
			want:        "4096",
		},
		// Level normalization
		{
			name:        "level uppercase normalized to lowercase",
			body:        `{"reasoning_effort":"HIGH"}`,
			model:       "gpt-5",
			toFormat:    "openai",
			providerKey: "openai",
			want:        "high",
		},
		{
			name:        "level mixed case normalized to lowercase",
			body:        `{"reasoning_effort":"Medium"}`,
			model:       "gpt-5",
			toFormat:    "openai",
			providerKey: "openai",
			want:        "medium",
		},
		// No thinking config
		{
			name:        "no thinking config returns none",
			body:        `{"messages":[{"role":"user","content":"hello"}]}`,
			model:       "gpt-5",
			toFormat:    "openai",
			providerKey: "openai",
			want:        "none",
		},
		{
			name:        "empty body returns none",
			body:        `{}`,
			model:       "gpt-5",
			toFormat:    "openai",
			providerKey: "openai",
			want:        "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := thinking.ThinkingLevelForUsage([]byte(tt.body), tt.model, tt.toFormat, tt.providerKey)
			if got != tt.want {
				t.Fatalf("ThinkingLevelForUsage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestThinkingLevelForUsage_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		model       string
		toFormat    string
		providerKey string
		want        string
	}{
		{
			name:        "invalid JSON body returns error",
			body:        `{invalid}`,
			model:       "gpt-5",
			toFormat:    "openai",
			providerKey: "openai",
			want:        "error",
		},
		{
			name:        "nil body returns none",
			body:        ``,
			model:       "gemini-2.5-pro",
			toFormat:    "gemini",
			providerKey: "gemini",
			want:        "none",
		},
		{
			name:        "case insensitive format",
			body:        `{"reasoning_effort":"high"}`,
			model:       "gpt-5",
			toFormat:    "OpenAI",
			providerKey: "openai",
			want:        "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := thinking.ThinkingLevelForUsage([]byte(tt.body), tt.model, tt.toFormat, tt.providerKey)
			if got != tt.want {
				t.Fatalf("ThinkingLevelForUsage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestThinkingLevelForUsage_ModelNotSupported(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		model       string
		toFormat    string
		providerKey string
		want        string
	}{
		{
			name:        "unknown model treated as user defined returns parsed value",
			body:        `{"reasoning_effort":"high"}`,
			model:       "unknown-model-xyz",
			toFormat:    "openai",
			providerKey: "openai",
			want:        "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := thinking.ThinkingLevelForUsage([]byte(tt.body), tt.model, tt.toFormat, tt.providerKey)
			if got != tt.want {
				t.Fatalf("ThinkingLevelForUsage() = %q, want %q", got, tt.want)
			}
		})
	}
}
