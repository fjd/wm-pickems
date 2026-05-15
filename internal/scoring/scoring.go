// Package scoring computes match (Tip) and tournament (Forecast) points from
// a per-League scoring config, recomputes on every result change, and builds
// League leaderboards with the agreed tiebreakers.
//
// Scale is tiny (friends app: a handful of users, 104 matches), so every
// result change triggers a full, idempotent recompute — simplest and always
// correct.
package scoring

import (
	"encoding/json"
	"sort"
	"strconv"

	"github.com/pocketbase/pocketbase/core"
)

// ---- Config ----

type Config struct {
	Match struct {
		Tendency   int  `json:"tendency"`
		Exact      int  `json:"exact"`
		TotalGoals int  `json:"totalGoals"`
		GoalDiff   int  `json:"goalDiff"`
		KoOtBonus  bool `json:"koOtBonus"`
		Advancer   int  `json:"advancer"`
	} `json:"match"`
	Forecast struct {
		GroupPosition     int            `json:"groupPosition"`
		PerfectGroupBonus int            `json:"perfectGroupBonus"`
		ThirdQualifier    int            `json:"thirdQualifier"`
		Round             map[string]int `json:"round"`
	} `json:"forecast"`
}

func loadConfig(rec *core.Record) Config {
	var c Config
	_ = json.Unmarshal([]byte(rec.GetString("config")), &c)
	return c
}

// configsInUse returns every scoring config referenced by a League plus the
// default, so per-(user,match,config) scores cover all Leagues.
func configsInUse(app core.App) (map[string]Config, string, error) {
	out := map[string]Config{}
	def, err := app.FindFirstRecordByFilter("scoring_configs", "isDefault = true")
	if err != nil {
		return nil, "", err
	}
	out[def.Id] = loadConfig(def)
	leagues, err := app.FindRecordsByFilter("leagues", "id != ''", "", 0, 0)
	if err != nil {
		return nil, "", err
	}
	for _, l := range leagues {
		cid := l.GetString("scoringConfig")
		if _, done := out[cid]; cid == "" || done {
			continue
		}
		if cr, err := app.FindRecordById("scoring_configs", cid); err == nil {
			out[cid] = loadConfig(cr)
		}
	}
	return out, def.Id, nil
}

func sign(n int) int {
	if n > 0 {
		return 1
	}
	if n < 0 {
		return -1
	}
	return 0
}

// ---- Match (Tip) scoring ----

type tipComponents struct {
	Tendency   int `json:"tendency"`
	Exact      int `json:"exact"`
	TotalGoals int `json:"totalGoals"`
	GoalDiff   int `json:"goalDiff"`
	OtBonus    int `json:"otBonus"`
	Advancer   int `json:"advancer"`
	GdDev      int `json:"gdDev"` // |predicted GD - actual GD| (tiebreaker only)
}

func (c tipComponents) points() int {
	return c.Tendency + c.Exact + c.TotalGoals + c.GoalDiff + c.OtBonus + c.Advancer
}

func scoreTip(cfg Config, match, tip *core.Record) tipComponents {
	var r tipComponents
	aH, aA := match.GetInt("ftHome"), match.GetInt("ftAway")
	pH, pA := tip.GetInt("ftHome"), tip.GetInt("ftAway")

	if sign(pH-pA) == sign(aH-aA) {
		r.Tendency = cfg.Match.Tendency
	}
	if pH == aH && pA == aA {
		r.Exact = cfg.Match.Exact
	}
	if pH+pA == aH+aA {
		r.TotalGoals = cfg.Match.TotalGoals
	}
	if pH-pA == aH-aA {
		r.GoalDiff = cfg.Match.GoalDiff
	}
	d := (pH - pA) - (aH - aA)
	if d < 0 {
		d = -d
	}
	r.GdDev = d

	if match.GetString("stage") == "group" {
		return r
	}

	// Knockout extras.
	if cfg.Match.KoOtBonus && aH == aA && pH == pA {
		// Game went to extra time and the user predicted a 90' draw.
		meH, meA := match.GetInt("etHome"), match.GetInt("etAway")
		teH, teA := tip.GetInt("etHome"), tip.GetInt("etAway")
		if meH != 0 || meA != 0 { // ET result present
			if teH == meH && teA == meA {
				r.OtBonus += cfg.Match.Exact
			}
			if teH+teA == meH+meA {
				r.OtBonus += cfg.Match.TotalGoals
			}
			if teH-teA == meH-meA {
				r.OtBonus += cfg.Match.GoalDiff
			}
		}
	}
	if a := match.GetString("advancer"); a != "" && a == tip.GetString("advancer") {
		r.Advancer = cfg.Match.Advancer
	}
	return r
}

// ---- Group standings (final, from finalized group matches) ----

type teamAgg struct {
	id                 string
	pts, gd, gf, games int
}

// finalGroups returns, for each fully-finished group, the ordered team ids
// (1st..4th) and collects the 12 third-placed teams for the best-third rank.
func finalGroups(app core.App) (order map[string][]string, thirds []teamAgg) {
	order = map[string][]string{}
	ms, _ := app.FindRecordsByFilter("matches",
		"stage = 'group' && finalizedAt != ''", "", 0, 0)
	groups := map[string]map[string]*teamAgg{}
	for _, m := range ms {
		g := m.GetString("groupLetter")
		if groups[g] == nil {
			groups[g] = map[string]*teamAgg{}
		}
		h, a := m.GetString("homeTeam"), m.GetString("awayTeam")
		hg, ag := m.GetInt("ftHome"), m.GetInt("ftAway")
		for _, id := range []string{h, a} {
			if groups[g][id] == nil {
				groups[g][id] = &teamAgg{id: id}
			}
		}
		H, A := groups[g][h], groups[g][a]
		H.games++
		A.games++
		H.gf += hg
		A.gf += ag
		H.gd += hg - ag
		A.gd += ag - hg
		switch {
		case hg > ag:
			H.pts += 3
		case ag > hg:
			A.pts += 3
		default:
			H.pts++
			A.pts++
		}
	}
	for g, tbl := range groups {
		if len(tbl) < 4 {
			continue
		}
		arr := make([]teamAgg, 0, 4)
		complete := true
		for _, v := range tbl {
			arr = append(arr, *v)
			if v.games < 3 {
				complete = false
			}
		}
		if !complete {
			continue
		}
		sortAggs(arr)
		ids := make([]string, len(arr))
		for i, v := range arr {
			ids[i] = v.id
		}
		order[g] = ids
		thirds = append(thirds, arr[2])
	}
	return order, thirds
}

func sortAggs(a []teamAgg) {
	sort.Slice(a, func(i, j int) bool {
		if a[i].pts != a[j].pts {
			return a[i].pts > a[j].pts
		}
		if a[i].gd != a[j].gd {
			return a[i].gd > a[j].gd
		}
		return a[i].gf > a[j].gf
	})
}

func bestThirdSet(thirds []teamAgg) map[string]bool {
	sortAggs(thirds)
	set := map[string]bool{}
	for i, t := range thirds {
		if i >= 8 {
			break
		}
		set[t.id] = true
	}
	return set
}

// ---- Forecast scoring ----

// actualRoundTeams maps stage -> set(teamId) of teams that actually reached
// that round, plus the actual champion.
func actualRoundTeams(app core.App) (map[string]map[string]bool, string) {
	res := map[string]map[string]bool{}
	champion := ""
	ms, _ := app.FindRecordsByFilter("matches", "stage != 'group'", "num", 0, 0)
	for _, m := range ms {
		st := m.GetString("stage")
		if res[st] == nil {
			res[st] = map[string]bool{}
		}
		for _, f := range []string{"homeTeam", "awayTeam"} {
			if id := m.GetString(f); id != "" {
				res[st][id] = true
			}
		}
		if st == "FINAL" && m.GetString("finalizedAt") != "" {
			champion = m.GetString("advancer")
		}
	}
	return res, champion
}

type fcResolver struct {
	order   map[string][]string
	thirds  map[string]string
	bracket map[string]string
	ko      map[int]*core.Record
}

func (r *fcResolver) resolve(label string, forNum int, seen map[int]bool) string {
	if label == "" {
		return ""
	}
	switch label[0] {
	case '1', '2':
		idx := 0
		if label[0] == '2' {
			idx = 1
		}
		o := r.order[label[1:]]
		if len(o) > idx {
			return o[idx]
		}
		return ""
	case '3':
		return r.thirds[strconv.Itoa(forNum)]
	case 'W', 'L':
		n, _ := strconv.Atoi(label[1:])
		if seen[n] {
			return ""
		}
		seen[n] = true
		w := r.bracket[strconv.Itoa(n)]
		if label[0] == 'W' {
			return w
		}
		src := r.ko[n]
		if src == nil || w == "" {
			return ""
		}
		h := r.resolve(src.GetString("homeLabel"), n, seen)
		a := r.resolve(src.GetString("awayLabel"), n, seen)
		if w == h {
			return a
		}
		if w == a {
			return h
		}
		return ""
	}
	return ""
}

func koStableKey(m *core.Record) string {
	if n := m.GetInt("num"); n > 0 {
		return strconv.Itoa(n)
	}
	return m.GetString("stage")
}

type fcBreakdown struct {
	Groups   int `json:"groups"`
	Thirds   int `json:"thirds"`
	Knockout int `json:"knockout"`
	Champion int `json:"champion"`
}

func (b fcBreakdown) total() int { return b.Groups + b.Thirds + b.Knockout + b.Champion }

func scoreForecast(app core.App, cfg Config, fc *core.Record) (fcBreakdown, int) {
	var b fcBreakdown

	var order map[string][]string
	_ = fc.UnmarshalJSONField("groupOrder", &order)
	var thirds map[string]string
	_ = fc.UnmarshalJSONField("thirdQualifiers", &thirds)
	var bracket map[string]string
	_ = fc.UnmarshalJSONField("bracket", &bracket)

	actualOrder, thirdAggs := finalGroups(app)
	for g, actual := range actualOrder {
		pred := order[g]
		allCorrect := len(pred) == 4
		for i := 0; i < 4 && i < len(actual); i++ {
			if i < len(pred) && pred[i] == actual[i] {
				b.Groups += cfg.Forecast.GroupPosition
			} else {
				allCorrect = false
			}
		}
		if allCorrect {
			b.Groups += cfg.Forecast.PerfectGroupBonus
		}
	}

	if len(thirdAggs) >= 12 { // all groups done -> best-8 fixed
		best := bestThirdSet(thirdAggs)
		for _, tid := range thirds {
			if best[tid] {
				b.Thirds += cfg.Forecast.ThirdQualifier
			}
		}
	}

	actualRounds, actualChamp := actualRoundTeams(app)
	koList, _ := app.FindRecordsByFilter("matches", "stage != 'group'", "num", 0, 0)
	koByNum := map[int]*core.Record{}
	for _, m := range koList {
		if n := m.GetInt("num"); n > 0 {
			koByNum[n] = m
		}
	}
	r := &fcResolver{order: order, thirds: thirds, bracket: bracket, ko: koByNum}

	for _, m := range koList {
		st := m.GetString("stage")
		w := cfg.Forecast.Round[st]
		if w == 0 {
			continue
		}
		predHome := r.resolve(m.GetString("homeLabel"), m.GetInt("num"), map[int]bool{})
		predAway := r.resolve(m.GetString("awayLabel"), m.GetInt("num"), map[int]bool{})
		for _, pid := range []string{predHome, predAway} {
			if pid != "" && actualRounds[st] != nil && actualRounds[st][pid] {
				b.Knockout += w
			}
		}
	}

	if actualChamp != "" {
		var champKey string
		for _, m := range koList {
			if m.GetString("stage") == "FINAL" {
				champKey = koStableKey(m)
			}
		}
		if champKey != "" && bracket[champKey] == actualChamp {
			b.Champion += cfg.Forecast.Round["CHAMPION"]
		}
	}

	return b, b.total()
}
