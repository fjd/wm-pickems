import { pb } from './pb';
import { auth } from './auth.svelte';
import { serverClock } from './serverclock.svelte';

export interface Team {
	id: string;
	name: string;
	iso2: string;
	fifaCode: string;
}

export interface Match {
	id: string;
	stage: string; // group | R32 | R16 | QF | SF | 3RD | FINAL
	groupLetter: string;
	roundLabel: string;
	num: number;
	kickoff: string;
	status: string;
	homeTeam: string;
	awayTeam: string;
	homeLabel: string;
	awayLabel: string;
	ftHome: number;
	ftAway: number;
	etHome: number;
	etAway: number;
	penHome: number;
	penAway: number;
	advancer: string;
	finalizedAt: string;
}

export interface Tip {
	id?: string;
	match: string;
	ftHome: number;
	ftAway: number;
	etHome: number;
	etAway: number;
	penWinner: string;
	advancer: string;
}

export interface FriendTip {
	userId: string;
	name: string;
	ftHome: number;
	ftAway: number;
	etHome: number;
	etAway: number;
	penWinner: string;
	advancer: string;
}

class TipsStore {
	teams = $state<Record<string, Team>>({});
	matches = $state<Match[]>([]);
	tips = $state<Record<string, Tip>>({}); // keyed by matchId
	scores = $state<Record<string, number>>({}); // matchId -> points (default cfg)
	tournamentGroups = $state<Record<string, string[]>>({}); // letter -> teamIds
	loaded = $state(false);

	async load() {
		const [teams, matches, mine, tgroups] = await Promise.all([
			pb.collection('teams').getFullList({ sort: 'name' }),
			pb.collection('matches').getFullList({ sort: 'kickoff' }),
			pb
				.collection('tips')
				.getFullList({ filter: `user = "${auth.user?.id}"` }),
			pb.collection('tournament_groups').getFullList({ sort: 'letter' }),
			serverClock.refresh(),
			pb
				.send('/api/tips/scores', { method: 'GET' })
				.then((r) => (this.scores = r.scores ?? {}))
				.catch(() => {})
		]);
		const gmap: Record<string, string[]> = {};
		for (const g of tgroups) gmap[g.letter] = g.teams ?? [];
		this.tournamentGroups = gmap;
		const tmap: Record<string, Team> = {};
		for (const t of teams)
			tmap[t.id] = {
				id: t.id,
				name: t.name,
				iso2: t.iso2,
				fifaCode: t.fifaCode
			};
		this.teams = tmap;
		this.matches = matches as unknown as Match[];
		const tip: Record<string, Tip> = {};
		for (const r of mine)
			tip[r.match] = {
				id: r.id,
				match: r.match,
				ftHome: r.ftHome,
				ftAway: r.ftAway,
				etHome: r.etHome,
				etAway: r.etAway,
				penWinner: r.penWinner,
				advancer: r.advancer
			};
		this.tips = tip;
		this.loaded = true;
	}

	team(id: string): Team | undefined {
		return this.teams[id];
	}

	/** Save (create or update) a tip; throws with the server message on a
	 *  rule/validation failure so the UI can show it. */
	async save(t: Tip): Promise<void> {
		const data = {
			user: auth.user?.id,
			match: t.match,
			ftHome: t.ftHome,
			ftAway: t.ftAway,
			etHome: t.etHome,
			etAway: t.etAway,
			penWinner: t.penWinner || null
		};
		let rec;
		if (t.id) {
			rec = await pb.collection('tips').update(t.id, data);
		} else {
			rec = await pb.collection('tips').create(data);
		}
		this.tips[t.match] = {
			id: rec.id,
			match: rec.match,
			ftHome: rec.ftHome,
			ftAway: rec.ftAway,
			etHome: rec.etHome,
			etAway: rec.etAway,
			penWinner: rec.penWinner,
			advancer: rec.advancer
		};
	}

	async friends(matchId: string): Promise<FriendTip[]> {
		const r = await pb.send(`/api/tips/others/${matchId}`, {
			method: 'GET'
		});
		return r.tips ?? [];
	}
}

export const tipsStore = new TipsStore();

export function isLocked(m: Match): boolean {
	return serverClock.now() >= new Date(m.kickoff).getTime();
}
export function teamsResolved(m: Match): boolean {
	return !!m.homeTeam && !!m.awayTeam;
}
