import type { Match, Tip } from './tips.svelte';

export type StandRow = { id: string; pts: number; gf: number; ga: number; pld: number };

// cmp ranks a group: points, then goal difference, then goals for, then id (a
// stable, deterministic last resort — not a real FIFA tiebreak).
function cmp(x: StandRow, y: StandRow): number {
	return (
		y.pts - x.pts ||
		y.gf - y.ga - (x.gf - x.ga) ||
		y.gf - x.gf ||
		x.id.localeCompare(y.id)
	);
}

// groupTable projects a group's standings: played matches count as-is, the rest
// use the user's saved pick (untipped + unplayed contribute nothing). All teams
// that appear in the group's fixtures are seeded so the table shows in full.
export function groupTable(matches: Match[], tips: Record<string, Tip>): StandRow[] {
	const tbl: Record<string, StandRow> = {};
	const ensure = (id: string) => (tbl[id] ||= { id, pts: 0, gf: 0, ga: 0, pld: 0 });
	for (const m of matches) {
		if (m.homeTeam) ensure(m.homeTeam);
		if (m.awayTeam) ensure(m.awayTeam);
	}
	for (const m of matches) {
		if (!m.homeTeam || !m.awayTeam) continue;
		const played = m.status === 'finished' || !!m.finalizedAt;
		let h: number, a: number;
		if (played) {
			h = m.ftHome;
			a = m.ftAway;
		} else {
			const t = tips[m.id];
			if (!t) continue;
			h = t.ftHome;
			a = t.ftAway;
		}
		const H = ensure(m.homeTeam);
		const A = ensure(m.awayTeam);
		H.pld++;
		A.pld++;
		H.gf += h;
		H.ga += a;
		A.gf += a;
		A.ga += h;
		if (h > a) H.pts += 3;
		else if (a > h) A.pts += 3;
		else {
			H.pts++;
			A.pts++;
		}
	}
	return Object.values(tbl).sort(cmp);
}

// groupComplete: every fixture has a result or a saved tip.
export function groupComplete(matches: Match[], tips: Record<string, Tip>): boolean {
	return matches.every(
		(m) =>
			!m.homeTeam ||
			!m.awayTeam ||
			m.status === 'finished' ||
			!!m.finalizedAt ||
			!!tips[m.id]
	);
}

// bestThirds returns the ids of the projected best `n` third-placed teams across
// all groups — but only once EVERY group is complete, since the comparison is
// top-n-of-all and is meaningless while some groups are unsettled. Returns an
// empty set until then.
export function bestThirds(
	groups: Match[][],
	tips: Record<string, Tip>,
	n = 8
): Set<string> {
	const thirds: StandRow[] = [];
	for (const g of groups) {
		if (!groupComplete(g, tips)) return new Set();
		const t = groupTable(g, tips);
		if (t.length >= 3) thirds.push(t[2]);
	}
	thirds.sort(cmp);
	return new Set(thirds.slice(0, n).map((r) => r.id));
}
