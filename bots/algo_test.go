package main

import (
	"context"
	"testing"
)

func testAlgo() *AlgoBrain {
	return NewAlgoBrain([]Team{
		{ID: "arg", FifaCode: "ARG"}, // 2140
		{ID: "bra", FifaCode: "BRA"}, // 2075
		{ID: "usa", FifaCode: "USA"}, // 1860
		{ID: "nzl", FifaCode: "NZL"}, // 1600
		{ID: "xx1", FifaCode: "ZZ1"}, // default 1500
		{ID: "xx2", FifaCode: "ZZ2"}, // default 1500
	})
}

func TestAlgoGroupsOrderByRating(t *testing.T) {
	a := testAlgo()
	order, thirds, _ := a.PredictGroups(context.Background(), []groupPick{
		{Letter: "A", Teams: []nameID{{ID: "nzl"}, {ID: "arg"}, {ID: "usa"}, {ID: "bra"}}},
	})
	want := []string{"arg", "bra", "usa", "nzl"}
	for i, id := range want {
		if order["A"][i] != id {
			t.Fatalf("order[A] = %v, want %v", order["A"], want)
		}
	}
	if len(thirds) != 1 || thirds[0] != "A" {
		t.Errorf("bestThirds = %v, want [A]", thirds)
	}
}

func TestAlgoWinnerHigherRated(t *testing.T) {
	a := testAlgo()
	// Stronger away team advances; equal ratings go to home.
	got, _ := a.PredictWinners(context.Background(), "R16", []matchup{
		{Num: 1, Home: nameID{ID: "nzl"}, Away: nameID{ID: "arg"}},
		{Num: 2, Home: nameID{ID: "xx1"}, Away: nameID{ID: "xx2"}},
	})
	if got[1] != "arg" {
		t.Errorf("match 1 winner = %q, want arg (higher rated)", got[1])
	}
	if got[2] != "xx1" {
		t.Errorf("match 2 winner = %q, want xx1 (home on a tie)", got[2])
	}
}

func TestAlgoTips(t *testing.T) {
	a := testAlgo()
	got, _ := a.PredictTips(context.Background(), []tipTarget{
		{MatchID: "g", Stage: "group", HomeID: "xx1", AwayID: "xx2"}, // equal
		{MatchID: "f", Stage: "group", HomeID: "arg", AwayID: "nzl"}, // big favourite home
		{MatchID: "k", Stage: "R16", HomeID: "xx1", AwayID: "xx2"},   // equal, knockout
	})
	if got["g"] != (Scoreline{Home: 1, Away: 1}) {
		t.Errorf("equal group tip = %v, want 1-1", got["g"])
	}
	if got["f"].Home <= got["f"].Away {
		t.Errorf("favourite tip = %v, want home > away", got["f"])
	}
	if got["k"].Home == got["k"].Away {
		t.Errorf("knockout tip = %v, must be decisive (no draw)", got["k"])
	}
}
