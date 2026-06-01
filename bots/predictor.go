package main

import "context"

// Predictor is the "brain" interface shared by every bot kind. The Claude brain
// (brain.go) and the algorithmic brain (algo.go) both implement it, so main.go
// can pick one by BOT_KIND and the rest of the flow is identical.
type Predictor interface {
	// PredictGroups orders each group 1st..4th (team ids), returns the 8 chosen
	// group letters whose third-placed team is expected to advance, and a
	// per-group rationale (letter -> text; empty for brains that don't explain).
	PredictGroups(ctx context.Context, groups []groupPick) (map[string][]string, []string, map[string]string, error)
	// PredictWinners picks the advancing team id for each resolved knockout
	// matchup (by match number), plus a per-match rationale (num -> text).
	PredictWinners(ctx context.Context, stageLabel string, ms []matchup) (map[int]string, map[int]string, error)
	// PredictTips returns an outcome distribution per match (keyed by match id);
	// the shared selectTip turns each into a concrete scoreline.
	PredictTips(ctx context.Context, targets []tipTarget) (map[string]TipOutcome, error)
}
