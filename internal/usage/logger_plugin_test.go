package usage

import (
	"context"
	"testing"
	"time"

	coreusage "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/usage"
)

func TestRecordIncludesThinkingLevelInSnapshot(t *testing.T) {
	stats := NewRequestStatistics()
	timestamp := time.Date(2026, time.March, 18, 10, 30, 0, 0, time.UTC)

	stats.Record(context.Background(), coreusage.Record{
		Provider:    "openai",
		Model:       "gpt-5",
		APIKey:      "api-key-1",
		Source:      "source-1",
		AuthIndex:   "1",
		RequestedAt: timestamp,
		Detail: coreusage.Detail{
			InputTokens:  10,
			OutputTokens: 5,
			TotalTokens:  15,
		},
		Thinking: coreusage.ThinkingDetail{ThinkingLevel: "high"},
	})

	snapshot := stats.Snapshot()
	apiSnapshot, ok := snapshot.APIs["api-key-1"]
	if !ok {
		t.Fatalf("snapshot missing api entry")
	}
	modelSnapshot, ok := apiSnapshot.Models["gpt-5"]
	if !ok {
		t.Fatalf("snapshot missing model entry")
	}
	if len(modelSnapshot.Details) != 1 {
		t.Fatalf("details length = %d, want 1", len(modelSnapshot.Details))
	}
	if got := modelSnapshot.Details[0].ThinkingLevel; got != "high" {
		t.Fatalf("thinking level = %q, want %q", got, "high")
	}
}

func TestMergeSnapshotKeepsDistinctThinkingLevels(t *testing.T) {
	stats := NewRequestStatistics()
	timestamp := time.Date(2026, time.March, 18, 11, 0, 0, 0, time.UTC)

	baseDetail := RequestDetail{
		Timestamp: timestamp,
		Source:    "source-1",
		AuthIndex: "1",
		Tokens: TokenStats{
			InputTokens:  10,
			OutputTokens: 5,
			TotalTokens:  15,
		},
	}

	result := stats.MergeSnapshot(StatisticsSnapshot{
		APIs: map[string]APISnapshot{
			"api-key-1": {
				Models: map[string]ModelSnapshot{
					"gpt-5": {
						Details: []RequestDetail{
							func() RequestDetail {
								detail := baseDetail
								detail.ThinkingLevel = "low"
								return detail
							}(),
							func() RequestDetail {
								detail := baseDetail
								detail.ThinkingLevel = "high"
								return detail
							}(),
						},
					},
				},
			},
		},
	})
	if result.Added != 2 {
		t.Fatalf("added = %d, want 2", result.Added)
	}
	if result.Skipped != 0 {
		t.Fatalf("skipped = %d, want 0", result.Skipped)
	}

	snapshot := stats.Snapshot()
	details := snapshot.APIs["api-key-1"].Models["gpt-5"].Details
	if len(details) != 2 {
		t.Fatalf("details length = %d, want 2", len(details))
	}
	if details[0].ThinkingLevel == details[1].ThinkingLevel {
		t.Fatalf("thinking levels should differ, got %q and %q", details[0].ThinkingLevel, details[1].ThinkingLevel)
	}

	result = stats.MergeSnapshot(StatisticsSnapshot{
		APIs: map[string]APISnapshot{
			"api-key-1": {
				Models: map[string]ModelSnapshot{
					"gpt-5": {
						Details: []RequestDetail{
							func() RequestDetail {
								detail := baseDetail
								detail.ThinkingLevel = "low"
								return detail
							}(),
						},
					},
				},
			},
		},
	})
	if result.Added != 0 {
		t.Fatalf("added on duplicate merge = %d, want 0", result.Added)
	}
	if result.Skipped != 1 {
		t.Fatalf("skipped on duplicate merge = %d, want 1", result.Skipped)
	}
}
