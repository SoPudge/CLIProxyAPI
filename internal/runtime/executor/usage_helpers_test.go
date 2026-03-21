package executor

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewUsageReporterDefaultsToNone(t *testing.T) {
	reporter := newUsageReporter(context.Background(), "gemini", "gemini-2.5-pro", nil)
	thinkingLevel := reporter.CaptureThinkingLevel(nil, "gemini-2.5-pro", "gemini", "gemini")
	if thinkingLevel != "none" {
		t.Fatalf("thinking level = %q, want %q", thinkingLevel, "none")
	}
}

func TestCaptureThinkingLevel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		model    string
		provider string
		toFormat string
		body     []byte
		want     string
	}{
		{
			name:     "openai chat completions",
			model:    "gpt-5",
			provider: "openai",
			toFormat: "openai",
			body:     []byte(`{"reasoning_effort":"high"}`),
			want:     "high",
		},
		{
			name:     "openai responses",
			model:    "gpt-5",
			provider: "openai",
			toFormat: "openai-response",
			body:     []byte(`{"reasoning":{"effort":"medium"}}`),
			want:     "medium",
		},
		{
			name:     "codex",
			model:    "gpt-5-codex",
			provider: "codex",
			toFormat: "codex",
			body:     []byte(`{"reasoning":{"effort":"low"}}`),
			want:     "low",
		},
		{
			name:     "claude",
			model:    "claude-sonnet-4-5",
			provider: "claude",
			toFormat: "claude",
			body:     []byte(`{"thinking":{"type":"enabled","budget_tokens":2048}}`),
			want:     "2048",
		},
		{
			name:     "gemini",
			model:    "gemini-2.5-pro",
			provider: "gemini",
			toFormat: "gemini",
			body:     []byte(`{"generationConfig":{"thinkingConfig":{"thinkingLevel":"high"}}}`),
			want:     "high",
		},
		{
			name:     "gemini cli",
			model:    "gemini-2.5-pro",
			provider: "gemini-cli",
			toFormat: "gemini-cli",
			body:     []byte(`{"request":{"generationConfig":{"thinkingConfig":{"thinkingBudget":1024}}}}`),
			want:     "1024",
		},
		{
			name:     "no thinking config returns none",
			model:    "gpt-5",
			provider: "openai",
			toFormat: "openai",
			body:     []byte(`{"messages":[{"role":"user","content":"hi"}]}`),
			want:     "none",
		},
		{
			name:     "empty body returns none",
			model:    "gpt-5",
			provider: "openai",
			toFormat: "openai",
			body:     nil,
			want:     "none",
		},
		{
			name:     "processed body with budget from suffix",
			model:    "gemini-2.5-pro",
			provider: "gemini",
			toFormat: "gemini",
			body:     []byte(`{"generationConfig":{"thinkingConfig":{"thinkingBudget":8192}}}`),
			want:     "8192",
		},
		{
			name:     "claude disabled",
			model:    "claude-sonnet-4-5",
			provider: "claude",
			toFormat: "claude",
			body:     []byte(`{"thinking":{"type":"disabled"}}`),
			want:     "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := newUsageReporter(context.Background(), tt.provider, tt.model, nil)
			thinkingLevel := reporter.CaptureThinkingLevel(tt.body, tt.model, tt.provider, tt.toFormat)
			if thinkingLevel != tt.want {
				t.Fatalf("thinking level = %q, want %q", thinkingLevel, tt.want)
			}
		})
	}
}

func TestParseOpenAIUsageChatCompletions(t *testing.T) {
	data := []byte(`{"usage":{"prompt_tokens":1,"completion_tokens":2,"total_tokens":3,"prompt_tokens_details":{"cached_tokens":4},"completion_tokens_details":{"reasoning_tokens":5}}}`)
	detail := parseOpenAIUsage(data)
	if detail.InputTokens != 1 {
		t.Fatalf("input tokens = %d, want %d", detail.InputTokens, 1)
	}
	if detail.OutputTokens != 2 {
		t.Fatalf("output tokens = %d, want %d", detail.OutputTokens, 2)
	}
	if detail.TotalTokens != 3 {
		t.Fatalf("total tokens = %d, want %d", detail.TotalTokens, 3)
	}
	if detail.CachedTokens != 4 {
		t.Fatalf("cached tokens = %d, want %d", detail.CachedTokens, 4)
	}
	if detail.ReasoningTokens != 5 {
		t.Fatalf("reasoning tokens = %d, want %d", detail.ReasoningTokens, 5)
	}
}

func TestParseOpenAIUsageResponses(t *testing.T) {
	data := []byte(`{"usage":{"input_tokens":10,"output_tokens":20,"total_tokens":30,"input_tokens_details":{"cached_tokens":7},"output_tokens_details":{"reasoning_tokens":9}}}`)
	detail := parseOpenAIUsage(data)
	if detail.InputTokens != 10 {
		t.Fatalf("input tokens = %d, want %d", detail.InputTokens, 10)
	}
	if detail.OutputTokens != 20 {
		t.Fatalf("output tokens = %d, want %d", detail.OutputTokens, 20)
	}
	if detail.TotalTokens != 30 {
		t.Fatalf("total tokens = %d, want %d", detail.TotalTokens, 30)
	}
	if detail.CachedTokens != 7 {
		t.Fatalf("cached tokens = %d, want %d", detail.CachedTokens, 7)
	}
	if detail.ReasoningTokens != 9 {
		t.Fatalf("reasoning tokens = %d, want %d", detail.ReasoningTokens, 9)
	}
}
