// SPDX-License-Identifier: MPL-2.0

package model

import "github.com/hashicorp/terraform-plugin-framework/types"

// RecommendationSegmentStats is a pattern's per-segment volume/drop-rate
// breakdown, keyed by the segment's LogQL selector in Recommendation.Segments.
type RecommendationSegmentStats struct {
	Volume              int64 `json:"volume"`
	IngestedLines       int64 `json:"ingested_lines"`
	QueriedLines        int64 `json:"queried_lines"`
	RecommendedDropRate int64 `json:"recommended_drop_rate"`
	Locked              bool  `json:"locked"`
}

// Recommendation mirrors one entry from GET /adaptive-logs/recommendations:
// a system-detected log pattern with a suggested drop rate, and its
// per-segment breakdown (the read-only counterpart to "segment overrides").
type Recommendation struct {
	Tokens              []string                              `json:"tokens"`
	Locked              bool                                  `json:"locked"`
	ConfiguredDropRate  int64                                 `json:"configured_drop_rate"`
	Volume              int64                                 `json:"volume"`
	IngestedLines       int64                                 `json:"ingested_lines"`
	QueriedLines        int64                                 `json:"queried_lines"`
	RecommendedDropRate int64                                 `json:"recommended_drop_rate"`
	Superseded          bool                                  `json:"superseded"`
	Levels              []string                              `json:"levels"`
	Segments            map[string]RecommendationSegmentStats `json:"segments"`
}

func (r Recommendation) ToTF() RecommendationTF {
	var tokens []types.String
	for _, t := range r.Tokens {
		tokens = append(tokens, types.StringValue(t))
	}
	var levels []types.String
	for _, l := range r.Levels {
		levels = append(levels, types.StringValue(l))
	}

	var segments map[string]RecommendationSegmentStatsTF
	if len(r.Segments) > 0 {
		segments = make(map[string]RecommendationSegmentStatsTF, len(r.Segments))
		for selector, stats := range r.Segments {
			segments[selector] = RecommendationSegmentStatsTF{
				Volume:              types.Int64Value(stats.Volume),
				IngestedLines:       types.Int64Value(stats.IngestedLines),
				QueriedLines:        types.Int64Value(stats.QueriedLines),
				RecommendedDropRate: types.Int64Value(stats.RecommendedDropRate),
				Locked:              types.BoolValue(stats.Locked),
			}
		}
	}

	return RecommendationTF{
		Tokens:              tokens,
		Locked:              types.BoolValue(r.Locked),
		ConfiguredDropRate:  types.Int64Value(r.ConfiguredDropRate),
		Volume:              types.Int64Value(r.Volume),
		IngestedLines:       types.Int64Value(r.IngestedLines),
		QueriedLines:        types.Int64Value(r.QueriedLines),
		RecommendedDropRate: types.Int64Value(r.RecommendedDropRate),
		Superseded:          types.BoolValue(r.Superseded),
		Levels:              levels,
		Segments:            segments,
	}
}

// RecommendationSegmentStatsTF is the Terraform-side representation of RecommendationSegmentStats.
type RecommendationSegmentStatsTF struct {
	Volume              types.Int64 `tfsdk:"volume"`
	IngestedLines       types.Int64 `tfsdk:"ingested_lines"`
	QueriedLines        types.Int64 `tfsdk:"queried_lines"`
	RecommendedDropRate types.Int64 `tfsdk:"recommended_drop_rate"`
	Locked              types.Bool  `tfsdk:"locked"`
}

// RecommendationTF is the Terraform-side representation of a Recommendation.
type RecommendationTF struct {
	Tokens              []types.String                          `tfsdk:"tokens"`
	Locked              types.Bool                              `tfsdk:"locked"`
	ConfiguredDropRate  types.Int64                             `tfsdk:"configured_drop_rate"`
	Volume              types.Int64                             `tfsdk:"volume"`
	IngestedLines       types.Int64                             `tfsdk:"ingested_lines"`
	QueriedLines        types.Int64                             `tfsdk:"queried_lines"`
	RecommendedDropRate types.Int64                             `tfsdk:"recommended_drop_rate"`
	Superseded          types.Bool                              `tfsdk:"superseded"`
	Levels              []types.String                          `tfsdk:"levels"`
	Segments            map[string]RecommendationSegmentStatsTF `tfsdk:"segments"`
}
