package rule_ecs

import (
	"testing"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/evaluator"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
)

func Test_DataProvider_EmptyRuleField(t *testing.T) {
	provider := NewDataProvider()

	event := &beat.Event{Fields: map[string]any{}} // DOES NOT have 'rule' key

	err := provider.EnrichEvent(event, fetching.ResourceMetadata{})
	assert.NoError(t, err, "should NOT fail when there is no 'rule' key")

	_, ok := event.Fields["rule"]
	require.False(t, ok, "event should NOT have 'rule' key")
}

func Test_DataProvider_UpdateRuleECSFields(t *testing.T) {
	tests := []struct {
		name           string
		ruleID         string
		ruleUUID       string // for negative cases only
		ruleReferences string
		ruleReference  string // for negative cases only
		wantID         string
		wantUUID       string
		wantReference  string
		wantReferences string
	}{
		{
			name: "got nothing, did nothing",
		},
		{
			name:     "id to uuid",
			ruleID:   "abc",
			wantID:   "abc",
			wantUUID: "abc",
		},
		{
			name:           "references to reference",
			ruleReferences: "ref",
			wantReference:  "ref",
			wantReferences: "ref",
		},
		{
			name:           "does not work backwards (uuid -/-> id, reference -/-> references)",
			ruleUUID:       "nope",
			wantID:         "",
			ruleReference:  "nope",
			wantReferences: "",
		},
		{
			name:           "all fields at once",
			ruleID:         "a",
			ruleReferences: "b",
			wantID:         "a",
			wantUUID:       "a",
			wantReference:  "b",
			wantReferences: "b",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewDataProvider()

			event := &beat.Event{
				Fields: map[string]any{
					"rule": evaluator.Rule{
						Id:         tc.ruleID,
						References: tc.ruleReferences,
					},
				},
			}

			err := provider.EnrichEvent(event, fetching.ResourceMetadata{})
			require.NoError(t, err)

			rule, ok := event.Fields["rule"].(evaluator.Rule)
			require.True(t, ok, "unexpected error unpacking event.Fields['rule']")

			assert.Equal(t, tc.wantID, rule.Id)
			assert.Equal(t, tc.wantUUID, rule.UUID)
			assert.Equal(t, rule.Id, rule.UUID)

			assert.Equal(t, tc.wantReference, rule.Reference)
			assert.Equal(t, tc.wantReferences, rule.References)
			assert.Equal(t, rule.References, rule.Reference)
		})
	}
}
