<script lang="ts">
	import { api, type LeagueSummary } from '$lib/api';
	import { goto } from '$app/navigation';
	import { t } from '$lib/i18n.svelte';
	import { Users } from '@lucide/svelte';

	let leagues = $state<LeagueSummary[]>([]);
	let loaded = $state(false);
	let newName = $state('');
	let joinCode = $state('');
	let error = $state('');
	let busy = $state(false);

	async function load() {
		try {
			leagues = (await api.myLeagues()).leagues;
		} catch {
			/* ignore */
		} finally {
			loaded = true;
		}
	}
	$effect(() => {
		load();
	});

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
				<span class="spacer"></span>
				<span class="cnt"><Users size={15} /> {l.members}</span>
			</a>
		{/each}
	{/if}
</section>

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
	</form>
</section>

{#if error}<p class="error">{error}</p>{/if}

<style>
	h1 {
		margin: 1rem 0 0.2rem;
	}
	.muted {
		margin: 0 0 1rem;
	}
	.lrow {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.7rem 0;
		border-top: 1px solid var(--border);
		color: var(--text);
	}
	.lrow:first-of-type {
		border-top: none;
	}
	.cnt {
		display: inline-flex;
		align-items: center;
		gap: 0.3rem;
		color: var(--muted);
		font-size: 0.9rem;
	}
	.code {
		text-transform: uppercase;
		letter-spacing: 0.2em;
		font-weight: 700;
	}
</style>
