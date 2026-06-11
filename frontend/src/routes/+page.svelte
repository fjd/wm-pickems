<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { auth } from '$lib/auth.svelte';
	import { api, type LeagueSummary } from '$lib/api';
	import { t } from '$lib/i18n.svelte';
	import { stageLabel } from '$lib/i18n.svelte';
	import { tipsStore, isLocked, teamsResolved, type Match } from '$lib/tips.svelte';
	import { countdown } from '$lib/countdown.svelte';
	import { serverClock } from '$lib/serverclock.svelte';
	import { pb } from '$lib/pb';
	import Flag from '$lib/components/Flag.svelte';
	import Landing from '$lib/components/Landing.svelte';
	import {
		Telescope,
		Volleyball,
		Trophy,
		Users,
		ChevronRight,
		Check,
		Clock,
		CircleHelp
	} from '@lucide/svelte';

	type Rank = { rank: number; total: number; points: number };

	let leagues = $state<LeagueSummary[]>([]);
	let ranks = $state<Record<string, Rank | null>>({});
	let leaguesLoaded = $state(false);

	// Global is the everyone-league — pin it to the top (matches the Leagues
	// page); other leagues keep the server order (sort is stable).
	let orderedLeagues = $derived(
		[...leagues].sort(
			(a, b) =>
				Number(b.inviteCode === 'GLOBAL') - Number(a.inviteCode === 'GLOBAL')
		)
	);
	let hasForecast = $state(false);
	let forecastChecked = $state(false);

	onMount(() => {
		if (!auth.isAuthed) return;
		countdown.start();
		tipsStore.load();
		// Has the user submitted their forecast yet?
		pb.collection('forecasts')
			.getList(1, 1, { filter: `user = "${auth.user?.id}"` })
			.then((r) => (hasForecast = r.items.length > 0))
			.catch(() => {})
			.finally(() => (forecastChecked = true));
		api
			.myLeagues()
			.then((r) => {
				leagues = r.leagues;
				r.leagues.forEach((l) => loadRank(l.id));
			})
			.catch(() => {})
			.finally(() => (leaguesLoaded = true));
	});
	onDestroy(() => countdown.stop());

	// My placement in a league: rank by total points (mirrors the Overall tab).
	function loadRank(id: string) {
		api
			.leaderboard(id)
			.then((res) => {
				const rows = [...res.rows].sort((a, b) => b.total - a.total);
				const i = rows.findIndex((row) => row.userId === auth.user?.id);
				ranks[id] = i >= 0 ? { rank: i + 1, total: rows.length, points: rows[i].total } : null;
			})
			.catch(() => (ranks[id] = null));
	}

	const pad = (n: number) => String(n).padStart(2, '0');
	const finishedM = (m: Match) => m.status === 'finished' || !!m.finalizedAt;
	const byKick = (a: Match, b: Match) =>
		new Date(a.kickoff).getTime() - new Date(b.kickoff).getTime();

	function roundOf(m: Match): string {
		return m.stage === 'group' ? `${t('common.group', { letter: m.groupLetter ?? '' })} · ${m.roundLabel}` : m.roundLabel;
	}
	function fmtKick(iso: string): string {
		return new Date(iso).toLocaleString(undefined, {
			weekday: 'short',
			day: '2-digit',
			month: 'short',
			hour: '2-digit',
			minute: '2-digit'
		});
	}
	// A team slot — resolved team, or the KO placeholder label ("W73", "1A").
	function slot(id: string, label: string) {
		const tm = id ? tipsStore.team(id) : undefined;
		return { name: tm?.name ?? label ?? 'TBD', iso2: tm?.iso2 ?? '', code: tm?.fifaCode ?? '' };
	}

	let started = $derived(countdown.locked); // first kickoff has passed
	let total = $derived(tipsStore.matches.length);
	let finished = $derived(tipsStore.matches.filter(finishedM).length);
	let progress = $derived(total ? Math.round((finished / total) * 100) : 0);
	let allDone = $derived(started && total > 0 && finished === total);

	// Current phase = stage of the next match still to kick off.
	let phase = $derived.by(() => {
		const now = serverClock.now();
		const ms = [...tipsStore.matches].sort(byKick);
		const next = ms.find((m) => new Date(m.kickoff).getTime() >= now);
		return stageLabel(next?.stage ?? ms[ms.length - 1]?.stage ?? '');
	});

	// Next up = soonest match not yet kicked off.
	let nextMatch = $derived.by(() => {
		const now = serverClock.now();
		return [...tipsStore.matches].sort(byKick).find((m) => new Date(m.kickoff).getTime() >= now) ?? null;
	});
	let nextTipped = $derived(nextMatch ? !!tipsStore.tips[nextMatch.id] : false);

	// Open matches you can still tip (teams resolved, not yet locked).
	let untipped = $derived(
		tipsStore.matches.filter((m) => teamsResolved(m) && !isLocked(m) && !tipsStore.tips[m.id]).length
	);

	// Smart next moves — only what's actually still outstanding.
	type Move = { href: string; icon: typeof Telescope; titleKey: string; subKey: string; params?: Record<string, string | number> };
	let moves = $derived.by(() => {
		const out: Move[] = [];
		if (forecastChecked && !countdown.locked && !hasForecast)
			out.push({
				href: '/forecast',
				icon: Telescope,
				titleKey: 'home.moveForecastTitle',
				subKey: 'home.moveForecastSub'
			});
		if (tipsStore.loaded && untipped > 0)
			out.push({
				href: '/tips',
				icon: Volleyball,
				titleKey: untipped === 1 ? 'home.moveTipsTitleOne' : 'home.moveTipsTitleMany',
				subKey: 'home.moveTipsSub',
				params: { n: untipped }
			});
		if (leaguesLoaded && leagues.length === 0)
			out.push({
				href: '/leagues',
				icon: Trophy,
				titleKey: 'home.moveLeagueTitle',
				subKey: 'home.moveLeagueSub'
			});
		return out;
	});
	let ready = $derived(forecastChecked && tipsStore.loaded && leaguesLoaded);
	let allCaught = $derived(ready && moves.length === 0);
</script>

{#if !auth.isAuthed}
	<Landing />
{:else}
	<header>
		<p class="kicker">{t('home.kicker')}</p>
		<h1>{@html t('home.greeting', { name: auth.user?.name ?? '' })}</h1>
		<p class="muted sd">{t('home.subtitle')}</p>
	</header>

	<div class="stagger">
		<!-- ===== Tournament progress / pre-tournament countdown ===== -->
		<section class="card prog">
			{#if !countdown.ready || !tipsStore.loaded}
				<p class="muted">{t('common.loading')}</p>
			{:else if countdown.kickoff && !countdown.locked}
				<p class="kicker2">{t('home.kickoffIn')}</p>
				<div class="cd">
					<span class="u"><b class="digits">{pad(countdown.parts.days)}</b><i>{t('countdown.days')}</i></span>
					<span class="u"><b class="digits">{pad(countdown.parts.hours)}</b><i>{t('countdown.hrs')}</i></span>
					<span class="u"><b class="digits">{pad(countdown.parts.mins)}</b><i>{t('countdown.min')}</i></span>
					<span class="u"><b class="digits">{pad(countdown.parts.secs)}</b><i>{t('countdown.sec')}</i></span>
				</div>
				<p class="muted fine">The opening match kicks off {fmtKick(new Date(countdown.kickoff).toISOString())}.</p>
			{:else}
				<div class="prog-head">
					<span class="phase-lbl">{allDone ? t('home.championsCrowned') : phase}</span>
					<span class="pct digits">{progress}%</span>
				</div>
				<div class="bar"><span style="width:{progress}%"></span></div>
				<p class="muted fine">{t('home.matchesPlayed', { finished, total })}</p>
			{/if}
		</section>

		<!-- ===== Next up match ===== -->
		{#if nextMatch}
			{@const H = slot(nextMatch.homeTeam, nextMatch.homeLabel)}
			{@const A = slot(nextMatch.awayTeam, nextMatch.awayLabel)}
			<a class="card next" href="/tips">
				<div class="row">
					<h3>{t('home.nextUp')}</h3>
					<div class="spacer"></div>
					<span class="muted small">{roundOf(nextMatch)}</span>
				</div>
				<div class="nm">
					<span class="nm-team">
						<Flag iso2={H.iso2} code={H.code} />
						<span class="nm-name">{H.name}</span>
					</span>
					<span class="nm-vs">{t('common.vs')}</span>
					<span class="nm-team right">
						<span class="nm-name">{A.name}</span>
						<Flag iso2={A.iso2} code={A.code} />
					</span>
				</div>
				<div class="nm-foot">
					<span class="muted small"><Clock size={14} /> {fmtKick(nextMatch.kickoff)}</span>
					<div class="spacer"></div>
					{#if nextTipped}
						<span class="pill ok"><Check size={12} /> {t('common.tipped')}</span>
					{:else if teamsResolved(nextMatch)}
						<span class="pill act">{t('home.tipIt')}</span>
					{:else}
						<span class="pill">{t('home.teamsTbd')}</span>
					{/if}
				</div>
			</a>
		{/if}

		<!-- ===== Your next moves ===== -->
		<section class="card">
			<h3>{t('home.nextMoves')}</h3>
			{#if !ready}
				<p class="muted">{t('common.loading')}</p>
			{:else if allCaught}
				<p class="caught"><span class="ci"><Check size={18} /></span> {t('home.allCaughtUp')}</p>
			{:else}
				<div class="moves">
					{#each moves as m (m.href)}
						{@const Icon = m.icon}
						<a class="move" href={m.href}>
							<span class="mi"><Icon size={20} /></span>
							<span class="mt">
								<span class="title">{t(m.titleKey, m.params)}</span>
								<span class="muted sub">{t(m.subKey)}</span>
							</span>
							<ChevronRight size={18} class="cr" />
						</a>
					{/each}
				</div>
			{/if}
		</section>

		<!-- ===== Your leagues (with placement) ===== -->
		<section class="card">
			<div class="row">
				<h3>{t('home.yourLeagues')}</h3>
				<div class="spacer"></div>
				<a class="pill" href="/leagues">{t('common.manage')}</a>
			</div>
			{#if !leaguesLoaded}
				<p class="muted">{t('common.loading')}</p>
			{:else if leagues.length === 0}
				<p class="muted">
					{t('home.noLeagues')} <a href="/leagues">{t('home.noLeaguesLink')}</a>
				</p>
			{:else}
				{#each orderedLeagues as l (l.id)}
					<a class="lrow" href={`/leagues/${l.id}`}>
						<span class="lname">{l.name}</span>
						{#if l.role === 'owner'}<span class="pill">{t('common.owner')}</span>{/if}
						<span class="spacer"></span>
						<span class="standing" title="Your placement · players">
							<Users size={15} />
							{#if ranks[l.id]}
								<b class="rk">#{ranks[l.id]?.rank}</b><small>/{ranks[l.id]?.total}</small>
							{:else}
								<span class="cnt">{l.members}</span>
							{/if}
						</span>
					</a>
				{/each}
			{/if}
		</section>

		<!-- ===== How does it work ===== -->
		<section class="card">
			<a class="move" href="/welcome">
				<span class="mi"><CircleHelp size={20} /></span>
				<span class="mt">
					<span class="title">{t('home.howDoesItWork')}</span>
					<span class="muted sub">{t('home.howDoesItWorkSub')}</span>
				</span>
				<ChevronRight size={18} class="cr" />
			</a>
		</section>
	</div>
{/if}

<style>
	header {
		margin: 0.25rem 0 1.25rem;
	}
	h1 {
		margin: 0;
		font-size: 1.6rem;
	}
	header .muted {
		margin: 0.2rem 0 0;
	}
	.small {
		font-size: 0.85rem;
	}

	/* ---------- progress / countdown ---------- */
	.kicker2 {
		font-size: 0.7rem;
		font-weight: 700;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--accent);
		margin: 0;
	}
	.prog-head {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 0.5rem;
		margin-bottom: 0.6rem;
	}
	.phase-lbl {
		font-weight: 700;
		font-size: 1.1rem;
	}
	.pct {
		font-weight: 700;
		font-size: 1.1rem;
		color: var(--accent);
	}
	.bar {
		height: 10px;
		border-radius: var(--radius-pill);
		background: var(--surface-2);
		overflow: hidden;
	}
	.bar > span {
		display: block;
		height: 100%;
		border-radius: var(--radius-pill);
		background: linear-gradient(90deg, var(--accent), var(--accent-2));
		transition: width 0.4s ease;
	}
	.fine {
		font-size: 0.8rem;
		margin: 0.55rem 0 0;
	}
	.cd {
		display: flex;
		gap: 0.7rem;
		margin: 0.35rem 0 0.3rem;
	}
	.cd .u {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.15rem;
		min-width: 2.4ch;
	}
	.cd .u b {
		font-size: 1.85rem;
		font-weight: 700;
		line-height: 1;
		font-variant-numeric: tabular-nums;
	}
	.cd .u i {
		font-style: normal;
		font-size: 0.58rem;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--muted);
	}

	/* ---------- next match ---------- */
	/* The whole card is the link to /tips. */
	.next {
		display: block;
		color: var(--text);
		text-decoration: none;
	}
	.nm {
		display: grid;
		grid-template-columns: 1fr auto 1fr;
		align-items: center;
		gap: 0.6rem;
		padding: 0.5rem 0 0.1rem;
		color: var(--text);
	}
	.nm-team {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		min-width: 0;
	}
	.nm-team.right {
		justify-content: flex-end;
	}
	.nm-name {
		font-weight: 600;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.nm-vs {
		font-size: 0.78rem;
		color: var(--muted);
	}
	.nm-foot {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-top: 0.6rem;
	}
	.nm-foot .small {
		display: inline-flex;
		align-items: center;
		gap: 0.3rem;
	}
	.pill.act {
		color: var(--accent);
		border-color: color-mix(in srgb, var(--accent) 45%, var(--border));
	}

	/* ---------- next moves ---------- */
	.moves {
		margin-top: 0.6rem;
	}
	.move {
		display: flex;
		align-items: center;
		gap: 0.85rem;
		padding: 0.75rem 0;
		border-top: 1px solid var(--border);
		color: var(--text);
	}
	.move:first-child {
		border-top: none;
	}
	.mi {
		display: grid;
		place-items: center;
		width: 38px;
		height: 38px;
		border-radius: var(--radius-sm);
		background: var(--surface-2);
		color: var(--accent);
		flex: none;
	}
	.mt {
		display: flex;
		flex-direction: column;
	}
	.title {
		font-weight: 600;
	}
	.sub {
		font-size: 0.82rem;
	}
	:global(.move .cr) {
		margin-left: auto;
		color: var(--muted);
	}
	.caught {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		margin: 0.6rem 0 0;
	}
	.ci {
		display: grid;
		place-items: center;
		width: 32px;
		height: 32px;
		flex: none;
		border-radius: var(--radius-pill);
		background: color-mix(in srgb, var(--accent) 18%, transparent);
		color: var(--accent);
	}

	/* ---------- leagues ---------- */
	.lrow {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.7rem 0;
		border-top: 1px solid var(--border);
		color: var(--text);
	}
	.lrow:first-of-type {
		border-top: none;
	}
	.lname {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	/* Combined right-hand indicator: people icon + your placement (#rank/size).
	   The /size doubles as the member count, so no separate count is shown. */
	.standing {
		display: inline-flex;
		align-items: baseline;
		gap: 0.3rem;
		color: var(--muted);
		font-variant-numeric: tabular-nums;
	}
	.standing :global(svg) {
		align-self: center;
	}
	.rk {
		color: var(--accent);
		font-weight: 700;
	}
	.standing small {
		font-size: 0.72rem;
		font-weight: 600;
	}
	.cnt {
		font-size: 0.9rem;
	}
</style>
