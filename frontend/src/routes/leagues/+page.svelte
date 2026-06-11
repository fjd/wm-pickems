<script lang="ts">
	import { api, type LeagueSummary } from '$lib/api';
	import { auth } from '$lib/auth.svelte';
	import { goto } from '$app/navigation';
	import { t } from '$lib/i18n.svelte';
	import { Users, Globe, ChevronRight, MessageSquare } from '@lucide/svelte';

	type Rank = { rank: number; total: number };

	let leagues = $state<LeagueSummary[]>([]);
	let ranks = $state<Record<string, Rank | null>>({});
	let unread = $state<Record<string, number>>({});
	let loaded = $state(false);
	let newName = $state('');
	let joinCode = $state('');
	let error = $state('');
	let busy = $state(false);

	const isGlobal = (l: LeagueSummary) => l.inviteCode === 'GLOBAL';

	// Global is the everyone-league — always pin it to the top. Other leagues
	// keep the server order (sort is stable).
	let ordered = $derived(
		[...leagues].sort((a, b) => Number(isGlobal(b)) - Number(isGlobal(a)))
	);

	async function load() {
		try {
			leagues = (await api.myLeagues()).leagues;
			leagues.forEach((l) => loadRank(l.id));
			api
				.chatUnread()
				.then((r) => (unread = r.unread))
				.catch(() => {});
		} catch {
			/* ignore */
		} finally {
			loaded = true;
		}
	}
	$effect(() => {
		load();
	});

	// My placement in a league: rank by total points (mirrors the Overall tab).
	// rows.length is the number of ranked players, i.e. the league's size.
	function loadRank(id: string) {
		api
			.leaderboard(id)
			.then(({ rows }) => {
				const i = rows.findIndex((r) => r.userId === auth.user?.id);
				ranks[id] = i >= 0 ? { rank: i + 1, total: rows.length } : null;
			})
			.catch(() => (ranks[id] = null));
	}

	async function create(e: Event) {
		e.preventDefault();
		error = '';
		busy = true;
		try {
			const r = await api.createLeague(newName);
			newName = '';
			goto(`/leagues/${r.id}`);
		} catch {
			error = t('errors.couldNotCreateLeague');
		} finally {
			busy = false;
		}
	}

	async function join(e: Event) {
		e.preventDefault();
		error = '';
		busy = true;
		try {
			const r = await api.joinLeague(joinCode);
			joinCode = '';
			goto(`/leagues/${r.id}`);
		} catch {
			error = t('errors.invalidInviteCode');
		} finally {
			busy = false;
		}
	}
</script>

<<<<<<< HEAD
<p class="kicker">{t('leagues.kicker')}</p>
<h1>{t('leagues.title')}</h1>
<p class="muted">{t('leagues.subtitle')}</p>

<section class="card">
	<h3>{t('leagues.yourLeagues')}</h3>
	{#if !loaded}
		<p class="muted">{t('common.loading')}</p>
	{:else if leagues.length === 0}
		<p class="muted">{t('leagues.noneYet')}</p>
	{:else}
		{#each leagues as l (l.id)}
			<a class="lrow" href={`/leagues/${l.id}`}>
				<span>{l.name}</span>
				{#if l.role === 'owner'}<span class="pill">{t('common.owner')}</span>{/if}
=======
<p class="kicker">Play your friends</p>
<h1>Leagues</h1>
<p class="muted sub">Private competitions — your predictions vs. your friends'.</p>

<h2 class="sec">Your leagues</h2>
{#if !loaded}
	<p class="muted pad">Loading…</p>
{:else if leagues.length === 0}
	<p class="muted pad">None yet — create one or join with a code below.</p>
{:else}
	<div class="llist">
		{#each ordered as l (l.id)}
			<a class="lrow" class:global={isGlobal(l)} href={`/leagues/${l.id}`}>
				{#if isGlobal(l)}
					<span class="gico" aria-hidden="true"><Globe size={18} /></span>
				{/if}
				<span class="lname">{l.name}</span>
				{#if l.role === 'owner'}<span class="pill">owner</span>{/if}
				{#if unread[l.id]}
					<span class="cbadge" title="Unread messages">
						<MessageSquare size={12} />
						{unread[l.id] > 99 ? '99+' : unread[l.id]}
					</span>
				{/if}
>>>>>>> upstream/main
				<span class="spacer"></span>
				<span class="standing" title="Your placement · players">
					<Users size={15} />
					{#if ranks[l.id]}
						<b class="rk">#{ranks[l.id]?.rank}</b><small>/{ranks[l.id]?.total}</small>
					{:else}
						<span class="cnt">{l.members}</span>
					{/if}
				</span>
				<ChevronRight size={18} class="cv" />
			</a>
		{/each}
	</div>
{/if}

<<<<<<< HEAD
<section class="card">
	<h3>{t('leagues.createLeague')}</h3>
	<form onsubmit={create}>
		<div class="field">
			<input class="input" placeholder={t('leagues.leagueNamePlaceholder')} bind:value={newName} required />
		</div>
		<button class="btn" disabled={busy || !newName.trim()}>{t('common.create')}</button>
	</form>
</section>

<section class="card">
	<h3>{t('leagues.joinLeague')}</h3>
	<form onsubmit={join}>
		<div class="field">
			<input
				class="input code"
				placeholder={t('leagues.inviteCodePlaceholder')}
				bind:value={joinCode}
				required
			/>
		</div>
		<button class="btn secondary" disabled={busy || !joinCode.trim()}>{t('common.join')}</button>
=======
<section class="card actions">
	<form class="action" onsubmit={create}>
		<input
			class="input"
			placeholder="New league name"
			bind:value={newName}
			required
		/>
		<button class="btn" disabled={busy || !newName.trim()}>Create</button>
	</form>
	<div class="orsep"><span>or join one</span></div>
	<form class="action" onsubmit={join}>
		<input
			class="input code"
			placeholder="INVITE CODE"
			bind:value={joinCode}
			required
		/>
		<button class="btn secondary" disabled={busy || !joinCode.trim()}>Join</button>
>>>>>>> upstream/main
	</form>
</section>

{#if error}<p class="error">{error}</p>{/if}

<style>
	h1 {
		margin: 1rem 0 0.2rem;
	}
	.sub {
		margin: 0 0 1.5rem;
	}
	.sec {
		font-size: 1.05rem;
		margin: 0 0 0.7rem;
	}
	.pad {
		padding: 0.4rem 0.2rem;
	}

	/* ---------- league list: the page's primary, comfortably tappable list --- */
	.llist {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		margin-bottom: 1.6rem;
	}
	.lrow {
		display: flex;
		align-items: center;
		gap: 0.65rem;
		min-height: 60px;
		padding: 0.85rem 0.95rem;
		background:
			linear-gradient(180deg, rgba(255, 255, 255, 0.025), transparent 40%),
			var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		color: var(--text);
		transition:
			border-color 0.15s ease,
			background 0.15s ease,
			transform 0.05s ease;
	}
	.lrow:hover {
		border-color: color-mix(in srgb, var(--accent) 45%, var(--border));
		background: color-mix(in srgb, var(--accent) 6%, var(--surface));
	}
	.lrow:active {
		transform: scale(0.992);
	}
	/* Global is special — give it a faint accent edge so it reads as the pinned,
	   everyone-league at the top. */
	.lrow.global {
		border-color: color-mix(in srgb, var(--accent) 30%, var(--border));
	}
	.gico {
		display: inline-grid;
		place-items: center;
		width: 30px;
		height: 30px;
		flex-shrink: 0;
		border-radius: var(--radius-pill);
		background: color-mix(in srgb, var(--accent) 16%, var(--surface-2));
		color: var(--accent);
	}
	.lname {
		font-weight: 600;
		font-size: 1.02rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.spacer {
		flex: 1;
	}
	.cbadge {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		flex-shrink: 0;
		padding: 0.1rem 0.45rem;
		border-radius: var(--radius-pill);
		background: var(--accent);
		color: var(--accent-fg);
		font-size: 0.72rem;
		font-weight: 800;
		font-variant-numeric: tabular-nums;
	}
	/* Combined right-hand indicator: people icon + your placement (#rank/size).
	   The /size doubles as the member count, so no separate count is shown. */
	.standing {
		display: inline-flex;
		align-items: baseline;
		gap: 0.3rem;
		flex-shrink: 0;
		color: var(--muted);
		font-variant-numeric: tabular-nums;
	}
	.standing :global(svg) {
		align-self: center;
	}
	.rk {
		color: var(--accent);
		font-weight: 700;
		font-size: 1rem;
	}
	.standing small {
		font-size: 0.78rem;
		font-weight: 600;
	}
	.cnt {
		font-size: 0.95rem;
	}
	:global(.lrow .cv) {
		color: var(--muted);
		flex-shrink: 0;
	}
	.lrow:hover :global(.cv) {
		color: var(--accent);
	}

	/* ---------- secondary actions: compact, lower-priority ------------------- */
	.actions {
		padding: 1rem;
	}
	.action {
		display: flex;
		gap: 0.55rem;
	}
	.action .input {
		flex: 1;
		min-width: 0;
	}
	.action .btn {
		width: auto;
		flex-shrink: 0;
		padding-inline: 1.3rem;
	}
	.code {
		text-transform: uppercase;
		letter-spacing: 0.2em;
		font-weight: 700;
	}
	.orsep {
		display: flex;
		align-items: center;
		gap: 0.7rem;
		margin: 0.85rem 0;
		color: var(--muted);
		font-size: 0.8rem;
	}
	.orsep::before,
	.orsep::after {
		content: '';
		flex: 1;
		height: 1px;
		background: var(--border);
	}
	.error {
		color: var(--danger);
		margin-top: 0.8rem;
	}
</style>
