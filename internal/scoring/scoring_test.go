package scoring

import "testing"

func defaultCfg() Config {
	var c Config
	c.Match.Tendency = 3
	c.Match.Exact = 1
	c.Match.TotalGoals = 1
	c.Match.GoalDiff = 1
	c.Match.KoOtBonus = true
	c.Match.Advancer = 2
	return c
}

func TestScoreValues(t *testing.T) {
	cfg := defaultCfg()
	tests := []struct {
		name      string
		m         MatchResult
		p         TipPrediction
		wantPts   int
		wantExact int
		wantGdDev int
		wantOt    int
		wantAdv   int
		wantTend  int
	}{
		{
			name:      "group exact",
			m:         MatchResult{Stage: "group", FtH: 2, FtA: 1},
			p:         TipPrediction{FtH: 2, FtA: 1},
			wantPts:   6, // 3+1+1+1
			wantExact: 1,
			wantTend:  3,
		},
		{
			name:      "group tendency only",
			m:         MatchResult{Stage: "group", FtH: 3, FtA: 1},
			p:         TipPrediction{FtH: 1, FtA: 0},
			wantPts:   3, // tendency only
			wantGdDev: 1,
			wantTend:  3,
		},
		{
			name:      "group totally wrong",
			m:         MatchResult{Stage: "group", FtH: 1, FtA: 0},
			p:         TipPrediction{FtH: 0, FtA: 2},
			wantPts:   0,
			wantGdDev: 3,
		},
		{
			name:      "group draw exact",
			m:         MatchResult{Stage: "group", FtH: 1, FtA: 1},
			p:         TipPrediction{FtH: 1, FtA: 1},
			wantPts:   6,
			wantExact: 1,
			wantTend:  3,
		},
		{
			name: "KO exact + ET bonus + advancer",
			m: MatchResult{
				Stage: "R32", FtH: 1, FtA: 1, EtH: 2, EtA: 1, Advancer: "T1",
			},
			p: TipPrediction{
				FtH: 1, FtA: 1, EtH: 2, EtA: 1, Advancer: "T1",
			},
			wantPts:   11, // 6 (ft) + 3 (ot) + 2 (adv)
			wantExact: 1,
			wantOt:    3,
			wantAdv:   2,
			wantTend:  3,
		},
		{
			name:      "KO advancer only",
			m:         MatchResult{Stage: "SF", FtH: 2, FtA: 0, Advancer: "T2"},
			p:         TipPrediction{FtH: 0, FtA: 1, Advancer: "T2"},
			wantPts:   2,
			wantAdv:   2,
			wantGdDev: 3, // |(0-1)-(2-0)|
		},
		{
			name: "group ignores ET and advancer",
			m: MatchResult{
				Stage: "group", FtH: 1, FtA: 1, EtH: 5, EtA: 0, Advancer: "T1",
			},
			p: TipPrediction{
				FtH: 1, FtA: 1, EtH: 5, EtA: 0, Advancer: "T1",
			},
			wantPts:   6, // no ot bonus, no advancer for group
			wantExact: 1,
			wantTend:  3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := scoreValues(cfg, tc.m, tc.p)
			if r.points() != tc.wantPts {
				t.Errorf("points = %d, want %d (%+v)", r.points(), tc.wantPts, r)
			}
			if r.Exact != tc.wantExact {
				t.Errorf("exact = %d, want %d", r.Exact, tc.wantExact)
			}
			if r.GdDev != tc.wantGdDev {
				t.Errorf("gdDev = %d, want %d", r.GdDev, tc.wantGdDev)
			}
			if r.OtBonus != tc.wantOt {
				t.Errorf("otBonus = %d, want %d", r.OtBonus, tc.wantOt)
			}
			if r.Advancer != tc.wantAdv {
				t.Errorf("advancer = %d, want %d", r.Advancer, tc.wantAdv)
			}
			if r.Tendency != tc.wantTend {
				t.Errorf("tendency = %d, want %d", r.Tendency, tc.wantTend)
			}
		})
	}
}

func TestKoOtBonusDisabled(t *testing.T) {
	cfg := defaultCfg()
	cfg.Match.KoOtBonus = false
	m := MatchResult{Stage: "R32", FtH: 1, FtA: 1, EtH: 2, EtA: 1, Advancer: "T1"}
	p := TipPrediction{FtH: 1, FtA: 1, EtH: 2, EtA: 1, Advancer: "T1"}
	r := scoreValues(cfg, m, p)
	if r.OtBonus != 0 {
		t.Fatalf("otBonus = %d, want 0 when disabled", r.OtBonus)
	}
	if r.points() != 8 { // 6 ft + 0 ot + 2 adv
		t.Fatalf("points = %d, want 8", r.points())
	}
}

func TestNoOtBonusWhenDecidedIn90(t *testing.T) {
	cfg := defaultCfg()
	// Decided in regulation: ET fields zero -> no OT bonus path.
	m := MatchResult{Stage: "QF", FtH: 2, FtA: 1, Advancer: "T1"}
	p := TipPrediction{FtH: 2, FtA: 1, Advancer: "T1"}
	r := scoreValues(cfg, m, p)
	if r.OtBonus != 0 {
		t.Fatalf("otBonus = %d, want 0", r.OtBonus)
	}
	if r.points() != 8 { // 6 + advancer 2
		t.Fatalf("points = %d, want 8", r.points())
	}
}
