package sync

import (
	"sort"
	"strconv"
	"strings"

	"github.com/pocketbase/pocketbase/core"

	"github.com/floholz/wm-pickems/internal/bracket"
)

// ResolveBracket fills knockout matches' homeTeam/awayTeam from their
// placeholder labels once the referenced results are known. This is what makes
// a knockout Tip become available (Phase 3): a Tip opens as soon as both teams
// of a matchup are resolved.
//
// Resolvable labels:
//   - "1A".."2L"      group winner / runner-up (once that group is complete)
//   - "3A/B/C/D/F"     a best-third slot (greedy interim allocation, see note)
//   - "W73" / "L101"   winner / loser of a finished knockout match
//
// NOTE: the best-third -> R32 slot mapping currently uses a greedy ranked
// allocation. FIFA uses a fixed combination table keyed by which group letters
// produced the 8 qualifying thirds; that exact table is implemented in Phase 4
// (it is also needed for the Forecast slotting) and will replace the greedy
// fill here.
func ResolveBracket(app core.App) error {
	matches, err := app.FindRecordsByFilter("matches", "id != ''", "num", 0, 0)
	if err != nil {
		return err
	}

	byNum := map[int]*core.Record{}
	for _, m := range matches {
		if n := m.GetInt("num"); n > 0 {
			byNum[n] = m
		}
	}

	first, second, thirds := groupStandings(matches)

	// Resolve the 8 R32 third-slots. With all 8 best thirds known, use FIFA's
	// official Annex C table; otherwise fall back to a deterministic greedy
	// fill (only hit while the group stage is still incomplete, when the
	// bracket can't be resolved yet anyway).
	quals := make([]string, 0, len(thirds))
	for _, st := range thirds {
		quals = append(quals, st.group)
	}
	thirdByNum := map[int]string{}
	if tbl, ok := bracket.Lookup(quals); ok {
		for _, m := range matches {
			if m.GetString("stage") != "R32" {
				continue
			}
			home, away := m.GetString("homeLabel"), m.GetString("awayLabel")
			isSlot := (strings.HasPrefix(home, "3") && strings.Contains(home, "/")) ||
				(strings.HasPrefix(away, "3") && strings.Contains(away, "/"))
			if !isSlot {
				continue
			}
			if w, ok := bracket.WinnerLetter(home, away); ok {
				thirdByNum[m.GetInt("num")] = thirdTeam[tbl[w]]
			}
		}
	} else {
		thirdQueue := make([]string, len(quals))
		copy(thirdQueue, quals)
		r32 := []*core.Record{}
		for _, m := range matches {
			if m.GetString("stage") == "R32" {
				r32 = append(r32, m)
			}
		}
		sort.Slice(r32, func(i, j int) bool {
			return r32[i].GetInt("num") < r32[j].GetInt("num")
		})
		for _, m := range r32 {
			for _, lbl := range []string{m.GetString("homeLabel"), m.GetString("awayLabel")} {
				if !strings.HasPrefix(lbl, "3") || !strings.Contains(lbl, "/") {
					continue
				}
				allowed := strings.Split(strings.TrimPrefix(lbl, "3"), "/")
				for i, g := range thirdQueue {
					if g == "" {
						continue
					}
					ok := false
					for _, a := range allowed {
						if g == a {
							ok = true
							break
						}
					}
					if ok {
						thirdByNum[m.GetInt("num")] = thirdTeam[g]
						thirdQueue[i] = ""
						break
					}
				}
			}
		}
	}

	resolve := func(label string, num int) string {
		if label == "" {
			return ""
		}
		switch label[0] {
		case '1':
			return first[label[1:]]
		case '2':
			return second[label[1:]]
		case '3':
			return thirdByNum[num]
		case 'W', 'L':
			n, err := strconv.Atoi(label[1:])
			if err != nil {
				return ""
			}
			src, ok := byNum[n]
			if !ok || src.GetString("finalizedAt") == "" {
				return ""
			}
			adv := src.GetString("advancer")
			if label[0] == 'W' {
				return adv
			}
			// loser = the side that is not the advancer
			h, a := src.GetString("homeTeam"), src.GetString("awayTeam")
			if adv == h {
				return a
			}
			if adv == a {
				return h
			}
			return ""
		}
		return ""
	}

	for _, m := range matches {
		if m.GetString("stage") == "group" {
			continue
		}
		changed := false
		num := m.GetInt("num")
		if m.GetString("homeTeam") == "" {
			if id := resolve(m.GetString("homeLabel"), num); id != "" {
				m.Set("homeTeam", id)
				changed = true
			}
		}
		if m.GetString("awayTeam") == "" {
			if id := resolve(m.GetString("awayLabel"), num); id != "" {
				m.Set("awayTeam", id)
				changed = true
			}
		}
		if changed {
			if err := app.Save(m); err != nil {
				return err
			}
		}
	}
	return nil
}

// thirdTeam maps a group letter to that group's third-placed team id; filled
// by groupStandings and read by the greedy best-third allocation above.
var thirdTeam = map[string]string{}

type standing struct {
	group string
	team  string
	pts   int
	gd    int
	gf    int
}

// groupStandings computes, from finished group matches only, the 1st/2nd team
// id per group letter (only when that group's 6 matches are all finished) plus
// the globally ranked list of third-placed teams.
func groupStandings(matches []*core.Record) (first, second map[string]string, thirds []standing) {
	first = map[string]string{}
	second = map[string]string{}

	type agg struct{ pts, gd, gf, played int }
	groups := map[string]map[string]*agg{} // letter -> teamId -> agg

	for _, m := range matches {
		if m.GetString("stage") != "group" || m.GetString("finalizedAt") == "" {
			continue
		}
		g := m.GetString("groupLetter")
		if groups[g] == nil {
			groups[g] = map[string]*agg{}
		}
		h, a := m.GetString("homeTeam"), m.GetString("awayTeam")
		hg, ag := m.GetInt("ftHome"), m.GetInt("ftAway")
		for _, id := range []string{h, a} {
			if groups[g][id] == nil {
				groups[g][id] = &agg{}
			}
		}
		ha, aa := groups[g][h], groups[g][a]
		ha.played++
		aa.played++
		ha.gf += hg
		aa.gf += ag
		ha.gd += hg - ag
		aa.gd += ag - hg
		switch {
		case hg > ag:
			ha.pts += 3
		case ag > hg:
			aa.pts += 3
		default:
			ha.pts++
			aa.pts++
		}
	}

	for g, tbl := range groups {
		order := make([]standing, 0, len(tbl))
		complete := true
		for id, v := range tbl {
			order = append(order, standing{group: g, team: id, pts: v.pts, gd: v.gd, gf: v.gf})
			if v.played < 3 { // each team plays 3 group games
				complete = false
			}
		}
		if len(tbl) < 4 {
			complete = false
		}
		sort.Slice(order, func(i, j int) bool {
			if order[i].pts != order[j].pts {
				return order[i].pts > order[j].pts
			}
			if order[i].gd != order[j].gd {
				return order[i].gd > order[j].gd
			}
			return order[i].gf > order[j].gf
		})
		if complete && len(order) >= 3 {
			first[g] = order[0].team
			second[g] = order[1].team
			thirds = append(thirds, order[2])
			thirdTeam[g] = order[2].team
		}
	}

	sort.Slice(thirds, func(i, j int) bool {
		if thirds[i].pts != thirds[j].pts {
			return thirds[i].pts > thirds[j].pts
		}
		if thirds[i].gd != thirds[j].gd {
			return thirds[i].gd > thirds[j].gd
		}
		return thirds[i].gf > thirds[j].gf
	})
	if len(thirds) > 8 {
		thirds = thirds[:8]
	}
	return first, second, thirds
}
