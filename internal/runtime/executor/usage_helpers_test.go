package executor

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v6/sdk/translator"
)

func TestNewUsageReporterUsesRequestedModelSuffix(t *testing.T) {
	reporter := newUsageReporter(context.Background(), "gemini", "gemini-2.5-pro", cliproxyexecutor.Options{SourceFormat: "gemini"}, "gemini-2.5-pro(8192)", nil)
	if reporter.thinkingLevel != "8192" {
		t.Fatalf("thinking level = %q, want %q", reporter.thinkingLevel, "8192")
	}
}

func TestResolveThinkingLevelPrefersEffectiveModelOverMetadata(t *testing.T) {
	opts := cliproxyexecutor.Options{
		SourceFormat: "gemini",
		Metadata: map[string]any{
			cliproxyexecutor.RequestedModelMetadataKey: "gemini-2.5-pro(2048)",
		},
	}

	got := resolveThinkingLevel(context.Background(), opts, "gemini-2.5-pro(1024)")
	if got != "1024" {
		t.Fatalf("thinking level = %q, want %q", got, "1024")
	}
}

func TestResolveThinkingLevelFallsBackToMetadataRequestedModel(t *testing.T) {
	opts := cliproxyexecutor.Options{
		SourceFormat: "claude",
		Metadata: map[string]any{
			cliproxyexecutor.RequestedModelMetadataKey: "claude-sonnet-4-5(4096)",
		},
	}

	got := resolveThinkingLevel(context.Background(), opts, "")
	if got != "4096" {
		t.Fatalf("thinking level = %q, want %q", got, "4096")
	}
}

func TestResolveThinkingLevelPrefersOriginalRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(recorder)
	ginCtx.Set("REQUEST_BODY_OVERRIDE", `{"reasoning_effort":"low"}`)
	ctx := context.WithValue(context.Background(), "gin", ginCtx)
	originalRequest := []byte(`{"reasoning_effort":"high"}`)

	got := resolveThinkingLevel(ctx, cliproxyexecutor.Options{SourceFormat: "openai", OriginalRequest: originalRequest}, "gpt-5")
	if got != "high" {
		t.Fatalf("thinking level = %q, want %q", got, "high")
	}
}

func TestResolveThinkingLevel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		requestedModel  string
		sourceFormat    string
		originalRequest []byte
		override        any
		want            string
	}{
		{
			name:            "openai chat completions",
			requestedModel:  "gpt-5",
			sourceFormat:    "openai",
			originalRequest: []byte(`{"reasoning_effort":"high"}`),
			want:            "high",
		},
		{
			name:            "openai responses",
			requestedModel:  "gpt-5",
			sourceFormat:    "openai-response",
			originalRequest: []byte(`{"reasoning":{"effort":"medium"}}`),
			want:            "medium",
		},
		{
			name:            "codex",
			requestedModel:  "gpt-5-codex",
			sourceFormat:    "codex",
			originalRequest: []byte(`{"reasoning":{"effort":"low"}}`),
			want:            "low",
		},
		{
			name:            "claude",
			requestedModel:  "claude-sonnet-4-5",
			sourceFormat:    "claude",
			originalRequest: []byte(`{"thinking":{"type":"enabled","budget_tokens":2048}}`),
			want:            "2048",
		},
		{
			name:            "gemini",
			requestedModel:  "gemini-2.5-pro",
			sourceFormat:    "gemini",
			originalRequest: []byte(`{"generationConfig":{"thinkingConfig":{"thinkingLevel":"high"}}}`),
			want:            "high",
		},
		{
			name:            "gemini cli",
			requestedModel:  "gemini-2.5-pro",
			sourceFormat:    "gemini-cli",
			originalRequest: []byte(`{"request":{"generationConfig":{"thinkingConfig":{"thinkingBudget":1024}}}}`),
			want:            "1024",
		},
		{
			name:           "request body override takes precedence",
			requestedModel: "gpt-5",
			sourceFormat:   "openai",
			override:       `{"reasoning_effort":"high"}`,
			want:           "high",
		},
		{
			name:            "no thinking config",
			requestedModel:  "gpt-5",
			sourceFormat:    "openai",
			originalRequest: []byte(`{"messages":[{"role":"user","content":"hi"}]}`),
			want:            "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.override != nil {
				recorder := httptest.NewRecorder()
				ginCtx, _ := gin.CreateTestContext(recorder)
				ginCtx.Set("REQUEST_BODY_OVERRIDE", tt.override)
				ctx = context.WithValue(ctx, "gin", ginCtx)
			}
			if got := resolveThinkingLevel(ctx, cliproxyexecutor.Options{SourceFormat: sdktranslator.FromString(tt.sourceFormat), OriginalRequest: tt.originalRequest}, tt.requestedModel); got != tt.want {
				t.Fatalf("thinking level = %q, want %q", got, tt.want)
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
