import { pb } from './pb';
import { auth } from './auth.svelte';
import type { Team } from './tips.svelte';

export interface KOMatch {
	num: number;
	stage: string;
	round: string;
	homeLabel: string;
	awayLabel: string;
}
export interface ThirdSlot {
	matchNum: number;
	allowed: string[];
}
export interface GroupDef {
	letter: string;
	teams: string[];
}

/** Stable key for a KO match: its number, or the stage for the number-less
 *  Final / third-place matches. */
export function koKey(m: { num: number; stage: string }): string {
	return m.num > 0 ? String(m.num) : m.stage;
}

class ForecastStore {
	loaded = $state(false);
	locked = $state(false);
	tournamentStart = $state<string>('');
	teams = $state<Record<string, Team>>({});
	groups = $state<GroupDef[]>([]);
	knockout = $state<KOMatch[]>([]);
	thirdSlots = $state<ThirdSlot[]>([]);

	// Editable forecast.
	recId: string | undefined;
	groupOrder = $state<Record<string, string[]>>({}); // letter -> [id x4]
	thirds = $state<Record<string, string>>({}); // matchNum -> teamId
	bracket = $state<Record<string, string>>({}); // koKey -> winner teamId

	// Actual results, for post-stage correctness indicators.
	results = $state<
		{
			stage: string;
			groupLetter: string;
			num: number;
			homeTeam: string;
			awayTeam: string;
			ftHome: number;
			ftAway: number;
			advancer: string;
			finished: boolean;
		}[]
	>([]);

	async load() {
		const [structure, teams, mine, matches] = await Promise.all([
			pb.send('/api/forecast/structure', { method: 'GET' }),
			pb.collection('teams').getFullList({ sort: 'name' }),
			pb
				.collection('forecasts')
				.getFullList({ filter: `user = "${auth.user?.id}"` }),
			pb.collection('matches').getFullList({ sort: 'kickoff' })
		]);
		this.results = (matches as unknown[]).map((m) => {
			const r = m as Record<string, unknown>;
			return {
				stage: r.stage as string,
				groupLetter: r.groupLetter as string,
				num: r.num as number,
				homeTeam: r.homeTeam as string,
				awayTeam: r.awayTeam as string,
				ftHome: r.ftHome as number,
				ftAway: r.ftAway as number,
				advancer: r.advancer as string,
				finished:
					r.status === 'finished' || !!(r.finalizedAt as string)
			};
		});
		const tmap: Record<string, Team> = {};
		for (const t of teams)
			tmap[t.id] = {
				id: t.id,
				name: t.name,
				iso2: t.iso2,
				fifaCode: t.fifaCode
			};
		this.teams = tmap;
		this.groups = structure.groups;
		this.knockout = structure.knockout;
		this.thirdSlots = structure.thirdSlots ?? [];
		this.tournamentStart = structure.tournamentStart;
		this.locked = structure.locked;

		const f = mine[0];
		this.recId = f?.id;
		// Default group order = the group's team list until the user reorders.
		const order: Record<string, string[]> = {};
		for (const g of this.groups)
			order[g.letter] = f?.groupOrder?.[g.letter]?.length
				? [...f.groupOrder[g.letter]]
				: [...g.teams];
		this.groupOrder = order;
		this.thirds = f?.thirdQualifiers ?? {};
		this.bracket = f?.bracket ?? {};
		this.loaded = true;
	}

	team(id: string) {
		return this.teams[id];
	}

	/** True once every group match is finished. */
	get groupStageDone(): boolean {
		const g = this.results.filter((r) => r.stage === 'group');
		return g.length > 0 && g.every((r) => r.finished);
	}

	// Standings (pts, gd, gf) for one group's finished matches.
	private standing(letter: string) {
		const t: Record<
			string,
			{ id: string; pts: number; gd: number; gf: number; p: number }
		> = {};
		for (const id of this.groups.find((x) => x.letter === letter)?.teams ??
			[])
			t[id] = { id, pts: 0, gd: 0, gf: 0, p: 0 };
		for (const m of this.results) {
			if (m.stage !== 'group' || m.groupLetter !== letter || !m.finished)
				continue;
			const H = t[m.homeTeam],
				A = t[m.awayTeam];
			if (!H || !A) continue;
			H.p++;
			A.p++;
			H.gf += m.ftHome;
			A.gf += m.ftAway;
			H.gd += m.ftHome - m.ftAway;
			A.gd += m.ftAway - m.ftHome;
			if (m.ftHome > m.ftAway) H.pts += 3;
			else if (m.ftHome < m.ftAway) A.pts += 3;
			else {
				H.pts++;
				A.pts++;
			}
		}
		return Object.values(t);
	}

	/** Actual final 1st→4th of a group, or null until it's complete. */
	actualOrder(letter: string): string[] | null {
		const rows = this.standing(letter);
		if (rows.length < 4 || rows.some((r) => r.p < 3)) return null;
		rows.sort((a, b) => b.pts - a.pts || b.gd - a.gd || b.gf - a.gf);
		return rows.map((r) => r.id);
	}

	/** The 8 teams that actually qualify as best thirds, or null until the
	 *  whole group stage is done. */
	actualBestThirds(): Set<string> | null {
		if (!this.groupStageDone) return null;
		const thirds: { id: string; pts: number; gd: number; gf: number }[] =
			[];
		for (const g of this.groups) {
			const rows = this.standing(g.letter).sort(
				(a, b) => b.pts - a.pts || b.gd - a.gd || b.gf - a.gf
			);
			if (rows[2]) thirds.push(rows[2]);
		}
		thirds.sort((a, b) => b.pts - a.pts || b.gd - a.gd || b.gf - a.gf);
		return new Set(thirds.slice(0, 8).map((t) => t.id));
	}

	/** Actual advancer of a knockout match number, '' if not finished. */
	advancerOf(num: number): string {
		const m = this.results.find((r) => r.num === num);
		return m && m.finished ? m.advancer : '';
	}

	move(letter: string, idx: number, dir: -1 | 1) {
		const arr = [...this.groupOrder[letter]];
		const j = idx + dir;
		if (j < 0 || j >= arr.length) return;
		[arr[idx], arr[j]] = [arr[j], arr[idx]];
		this.groupOrder[letter] = arr;
	}

	/** Resolve a placeholder label ("1A","2B","3A/B/..","W74","L101") to a
	 *  team id given the current predictions, or '' if undecidable. */
	resolve(label: string, forMatchNum: number, seen = new Set<number>()): string {
		if (!label) return '';
		const c = label[0];
		if (c === '1' || c === '2') {
			const letter = label.slice(1);
			return this.groupOrder[letter]?.[c === '1' ? 0 : 1] ?? '';
		}
		if (c === '3') return this.thirdAssignment()[forMatchNum] ?? '';
		if (c === 'W' || c === 'L') {
			const n = parseInt(label.slice(1), 10);
			if (seen.has(n)) return '';
			seen.add(n);
			const w = this.bracket[String(n)] ?? '';
			if (c === 'W') return w;
			const src = this.knockout.find((m) => m.num === n);
			if (!src || !w) return '';
			const h = this.resolve(src.homeLabel, n, seen);
			const a = this.resolve(src.awayLabel, n, seen);
			return w === h ? a : w === a ? h : '';
		}
		return '';
	}

	sides(m: KOMatch): [string, string] {
		return [
			this.resolve(m.homeLabel, m.num),
			this.resolve(m.awayLabel, m.num)
		];
	}

	pick(m: KOMatch, teamId: string) {
		if (!teamId) return;
		this.bracket[koKey(m)] = teamId;
	}

	readonly maxThirds = 8;

	/** The predicted 3rd-placed team of a group (from the current order). */
	groupThird(letter: string): string {
		return this.groupOrder[letter]?.[2] ?? '';
	}

	/** Letters the user ticked to advance as a best third. */
	get chosenThirdLetters(): string[] {
		return Object.keys(this.thirds);
	}

	toggleThird(letter: string) {
		if (this.thirds[letter]) {
			delete this.thirds[letter];
			this.thirds = { ...this.thirds };
		} else if (this.chosenThirdLetters.length < this.maxThirds) {
			this.thirds = { ...this.thirds, [letter]: this.groupThird(letter) };
		}
	}

	/** Slot the chosen thirds into the 8 R32 third-slots. Greedy can dead-end
	 *  (a slot left empty though a full assignment exists), so this does a
	 *  deterministic backtracking perfect matching — slots in match order,
	 *  letters tried in alphabetical order. Mirrors the Go scorer exactly so
	 *  Forecast knockout scoring agrees. */
	thirdAssignment(): Record<number, string> {
		const slots = [...this.thirdSlots].sort(
			(a, b) => a.matchNum - b.matchNum
		);
		const chosen = this.chosenThirdLetters.sort();
		const assign: (string | null)[] = new Array(slots.length).fill(null);

		const solve = (i: number): boolean => {
			if (i === slots.length) return true;
			for (const letter of chosen) {
				if (assign.includes(letter)) continue;
				if (!slots[i].allowed.includes(letter)) continue;
				assign[i] = letter;
				if (solve(i + 1)) return true;
				assign[i] = null;
			}
			return false;
		};
		solve(0);

		const out: Record<number, string> = {};
		slots.forEach((s, i) => {
			if (assign[i]) out[s.matchNum] = this.groupThird(assign[i] as string);
		});
		return out;
	}

	async save() {
		// Persist thirds as {groupLetter: currentThirdTeamId} so the value
		// stays correct even if the group order changed after ticking.
		const thirdQualifiers: Record<string, string> = {};
		for (const letter of this.chosenThirdLetters)
			thirdQualifiers[letter] = this.groupThird(letter);
		const data = {
			user: auth.user?.id,
			groupOrder: this.groupOrder,
			thirdQualifiers,
			bracket: this.bracket
		};
		const rec = this.recId
			? await pb.collection('forecasts').update(this.recId, data)
			: await pb.collection('forecasts').create(data);
		this.recId = rec.id;
	}
}

export const forecastStore = new ForecastStore();
