package main

import (
	"context"
	"math"
	"sort"
)

// AlgoBrain is a deterministic, API-free predictor: a rating-based ("Elo-lite")
// model. Each team gets a strength rating (from an embedded table keyed by FIFA
// code, with a neutral default for unknowns); group order, the bracket, and
// scorelines all follow from those ratings. Same inputs → same predictions, so
// it's a stable "house" baseline for the others to beat.
type AlgoBrain struct {
	rating map[string]int // teamId -> rating
}

func NewAlgoBrain(teams []Team) *AlgoBrain {
	r := make(map[string]int, len(teams))
	for _, t := range teams {
		r[t.ID] = ratingFor(t.FifaCode)
	}
	return &AlgoBrain{rating: r}
}

func (a *AlgoBrain) rat(id string) int {
	if v, ok := a.rating[id]; ok {
		return v
	}
	return defaultRating
}

// PredictGroups: order each group by rating (desc), and pick the 8 groups whose
// third-placed team is the strongest. Ties break by id/letter for determinism.
func (a *AlgoBrain) PredictGroups(_ context.Context, groups []groupPick) (map[string][]string, []string, error) {
	order := make(map[string][]string, len(groups))
	type third struct {
		letter string
		r      int
	}
	var thirds []third
	for _, g := range groups {
		ids := make([]string, len(g.Teams))
		for i, t := range g.Teams {
			ids[i] = t.ID
		}
		sort.SliceStable(ids, func(i, j int) bool {
			if ri, rj := a.rat(ids[i]), a.rat(ids[j]); ri != rj {
				return ri > rj
			}
			return ids[i] < ids[j]
		})
		order[g.Letter] = ids
		if len(ids) >= 3 {
			thirds = append(thirds, third{letter: g.Letter, r: a.rat(ids[2])})
		}
	}
	sort.SliceStable(thirds, func(i, j int) bool {
		if thirds[i].r != thirds[j].r {
			return thirds[i].r > thirds[j].r
		}
		return thirds[i].letter < thirds[j].letter
	})
	best := make([]string, 0, 8)
	for i := 0; i < len(thirds) && i < 8; i++ {
		best = append(best, thirds[i].letter)
	}
	return order, best, nil
}

// PredictWinners: the higher-rated team advances; a tie goes to the home side.
func (a *AlgoBrain) PredictWinners(_ context.Context, _ string, ms []matchup) (map[int]string, error) {
	out := make(map[int]string, len(ms))
	for _, m := range ms {
		if a.rat(m.Away.ID) > a.rat(m.Home.ID) {
			out[m.Num] = m.Away.ID
		} else {
			out[m.Num] = m.Home.ID
		}
	}
	return out, nil
}

// PredictTips: expected goals from the rating gap. Group games may draw;
// knockouts are coerced to a decisive favourite (the server needs a winner).
func (a *AlgoBrain) PredictTips(_ context.Context, targets []tipTarget) (map[string]Scoreline, error) {
	out := make(map[string]Scoreline, len(targets))
	for _, t := range targets {
		rh, ra := a.rat(t.HomeID), a.rat(t.AwayID)
		h, av := expectedGoals(rh, ra), expectedGoals(ra, rh)
		if t.Stage != "group" && h == av {
			if rh >= ra {
				h++
			} else {
				av++
			}
		}
		out[t.MatchID] = Scoreline{Home: h, Away: av}
	}
	return out, nil
}

const (
	algoBase     = 1.4   // expected goals for an evenly-matched side
	algoStep     = 120.0 // rating points ≈ one extra/fewer goal
	algoMaxGoals = 6
)

// expectedGoals maps a rating gap to a goal tally: round(base ± gap/step),
// clamped. Equal teams → 1, a ~120-point edge → 2 vs 0, a big gap → 3+.
func expectedGoals(forRating, oppRating int) int {
	g := int(math.Round(algoBase + float64(forRating-oppRating)/algoStep))
	return max(0, min(g, algoMaxGoals))
}

// defaultRating is used for any team whose FIFA code isn't in the table below
// (e.g. a late or unexpected qualifier) — a mid/low baseline.
const defaultRating = 1500

// ratings is an approximate strength table for likely WC2026 nations, keyed by
// FIFA 3-letter code. Values are rough Elo-style numbers (stronger = higher);
// they only need to rank teams sensibly relative to each other. Edit freely —
// this single table is what gives the algo bot its "opinion".
var ratings = map[string]int{
	// Top tier
	"ARG": 2140, "FRA": 2120, "ESP": 2110, "ENG": 2080, "BRA": 2075,
	"POR": 2050, "NED": 2030, "BEL": 2000, "GER": 2010, "CRO": 1980,
	// Strong
	"ITA": 1985, "URU": 1960, "COL": 1945, "MAR": 1940, "USA": 1860,
	"MEX": 1855, "SUI": 1900, "DEN": 1905, "JPN": 1900, "SEN": 1915,
	"KOR": 1850, "SRB": 1880, "POL": 1860, "AUT": 1875, "ECU": 1865,
	// Mid
	"UKR": 1820, "AUS": 1790, "WAL": 1800, "TUR": 1830, "CAN": 1810,
	"NGA": 1825, "EGY": 1800, "IRN": 1815, "PER": 1780, "CHI": 1790,
	"CIV": 1795, "TUN": 1760, "GHA": 1760, "CMR": 1790, "RSA": 1740,
	// Lower
	"QAT": 1680, "KSA": 1670, "PAN": 1670, "CRC": 1700, "JAM": 1680,
	"PAR": 1730, "VEN": 1730, "NZL": 1600, "UZB": 1680, "IRQ": 1650,
	"JOR": 1640, "OMA": 1610, "UAE": 1600, "BOL": 1620,
}

func ratingFor(fifaCode string) int {
	if v, ok := ratings[fifaCode]; ok {
		return v
	}
	return defaultRating
}
