package thinking_test

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/thinking"
)

func TestThinkingLevelForUsage_SuffixPriority(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		model      string
		fromFormat string
		want       string
	}{
		{
			name:       "suffix budget takes priority over body",
			body:       `{"thinking":{"budget_tokens":1024}}`,
			model:      "claude-sonnet-4-5(8192)",
			fromFormat: "claude",
			want:       "8192",
		},
		{
			name:       "suffix level takes priority over body",
			body:       `{"reasoning_effort":"low"}`,
			model:      "gpt-5(high)",
			fromFormat: "openai",
			want:       "high",
		},
		{
			name:       "suffix none",
			body:       `{"thinking":{"budget_tokens":1024}}`,
			model:      "claude-sonnet-4-5(none)",
			fromFormat: "claude",
			want:       "none",
		},
		{
			name:       "suffix auto",
			body:       `{}`,
			model:      "gemini-2.5-pro(auto)",
			fromFormat: "gemini",
			want:       "auto",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := thinking.ThinkingLevelForUsage([]byte(tt.body), tt.model, tt.fromFormat)
			if got != tt.want {
				t.Fatalf("ThinkingLevelForUsage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestThinkingLevelForUsage_BodyConfig(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		model      string
		fromFormat string
		want       string
	}{
		// OpenAI format
		{
			name:       "openai reasoning_effort high",
			body:       `{"reasoning_effort":"high"}`,
			model:      "gpt-5",
			fromFormat: "openai",
			want:       "high",
		},
		{
			name:       "openai reasoning_effort medium",
			body:       `{"reasoning_effort":"medium"}`,
			model:      "gpt-5",
			fromFormat: "openai",
			want:       "medium",
		},
		{
			name:       "openai reasoning_effort low",
			body:       `{"reasoning_effort":"low"}`,
			model:      "gpt-5",
			fromFormat: "openai",
			want:       "low",
		},
		// Codex format
		{
			name:       "codex reasoning.effort",
			body:       `{"reasoning":{"effort":"high"}}`,
			model:      "gpt-5-codex",
			fromFormat: "codex",
			want:       "high",
		},
		// Claude format
		{
			name:       "claude budget_tokens",
			body:       `{"thinking":{"type":"enabled","budget_tokens":2048}}`,
			model:      "claude-sonnet-4-5",
			fromFormat: "claude",
			want:       "2048",
		},
		{
			name:       "claude budget_tokens auto (-1)",
			body:       `{"thinking":{"budget_tokens":-1}}`,
			model:      "claude-sonnet-4-5",
			fromFormat: "claude",
			want:       "auto",
		},
		{
			name:       "claude adaptive output_config.effort",
			body:       `{"thinking":{"type":"adaptive"},"output_config":{"effort":"high"}}`,
			model:      "claude-sonnet-4-5",
			fromFormat: "claude",
			want:       "high",
		},
		// Gemini format
		{
			name:       "gemini thinkingLevel",
			body:       `{"generationConfig":{"thinkingConfig":{"thinkingLevel":"high"}}}`,
			model:      "gemini-2.5-pro",
			fromFormat: "gemini",
			want:       "high",
		},
		{
			name:       "gemini thinkingBudget",
			body:       `{"generationConfig":{"thinkingConfig":{"thinkingBudget":1024}}}`,
			model:      "gemini-2.5-pro",
			fromFormat: "gemini",
			want:       "1024",
		},
		// Gemini CLI format
		{
			name:       "gemini-cli thinkingLevel with prefix",
			body:       `{"request":{"generationConfig":{"thinkingConfig":{"thinkingLevel":"medium"}}}}`,
			model:      "gemini-2.5-pro",
			fromFormat: "gemini-cli",
			want:       "medium",
		},
		// No config
		{
			name:       "no thinking config",
			body:       `{"messages":[{"role":"user","content":"hello"}]}`,
			model:      "gpt-5",
			fromFormat: "openai",
			want:       "",
		},
		{
			name:       "empty body",
			body:       `{}`,
			model:      "gpt-5",
			fromFormat: "openai",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := thinking.ThinkingLevelForUsage([]byte(tt.body), tt.model, tt.fromFormat)
			if got != tt.want {
				t.Fatalf("ThinkingLevelForUsage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestThinkingLevelForUsage_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		model      string
		fromFormat string
		want       string
	}{
		{
			name:       "invalid JSON body",
			body:       `{invalid}`,
			model:      "gpt-5",
			fromFormat: "openai",
			want:       "",
		},
		{
			name:       "nil body with suffix",
			body:       ``,
			model:      "gemini-2.5-pro(4096)",
			fromFormat: "gemini",
			want:       "4096",
		},
		{
			name:       "empty model name",
			body:       `{"reasoning_effort":"high"}`,
			model:      "",
			fromFormat: "openai",
			want:       "high",
		},
		{
			name:       "case insensitive format",
			body:       `{"reasoning_effort":"high"}`,
			model:      "gpt-5",
			fromFormat: "OpenAI",
			want:       "high",
		},
		{
			name:       "whitespace in model",
			body:       `{}`,
			model:      "  gpt-5(high)  ",
			fromFormat: "openai",
			want:       "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := thinking.ThinkingLevelForUsage([]byte(tt.body), tt.model, tt.fromFormat)
			if got != tt.want {
				t.Fatalf("ThinkingLevelForUsage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestThinkingLevelForUsage_MultipleFormats(t *testing.T) {
	// Test openai-response format (alias for codex)
	tests := []struct {
		name       string
		body       string
		model      string
		fromFormat string
		want       string
	}{
		{
			name:       "openai-response format",
			body:       `{"reasoning":{"effort":"medium"}}`,
			model:      "gpt-5",
			fromFormat: "openai-response",
			want:       "medium",
		},
		{
			name:       "antigravity format",
			body:       `{"request":{"generationConfig":{"thinkingConfig":{"thinkingLevel":"high"}}}}`,
			model:      "gemini-2.5-pro",
			fromFormat: "antigravity",
			want:       "high",
		},
		{
			name:       "kimi format (openai compatible)",
			body:       `{"reasoning_effort":"high"}`,
			model:      "kimi-model",
			fromFormat: "kimi",
			want:       "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := thinking.ThinkingLevelForUsage([]byte(tt.body), tt.model, tt.fromFormat)
			if got != tt.want {
				t.Fatalf("ThinkingLevelForUsage() = %q, want %q", got, tt.want)
			}
		})
	}
}
