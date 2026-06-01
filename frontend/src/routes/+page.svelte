<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { api, type LeagueSummary } from '$lib/api';
	import { t } from '$lib/i18n.svelte';
	import { Telescope, Volleyball, Trophy, Users, ChevronRight } from '@lucide/svelte';

	let leagues = $state<LeagueSummary[]>([]);
	let loaded = $state(false);

	$effect(() => {
		if (!auth.isAuthed) return;
		api
			.myLeagues()
			.then((r) => (leagues = r.leagues))
			.catch(() => {})
			.finally(() => (loaded = true));
	});

	const moves = [
		{
			href: '/forecast',
			icon: Telescope,
			titleKey: 'home.moveForecastTitle',
			subKey: 'home.moveForecastSub'
		},
		{
			href: '/tips',
			icon: Volleyball,
			titleKey: 'home.moveTipsTitle',
			subKey: 'home.moveTipsSub'
		},
		{
			href: '/leagues',
			icon: Trophy,
			titleKey: 'home.moveLeagueTitle',
			subKey: 'home.moveLeagueSub'
		}
	];
</script>

<header>
	<p class="kicker">{t('home.kicker')}</p>
	<h1>{@html t('home.greeting', { name: auth.user?.name ?? '' })}</h1>
	<p class="muted sd">{t('home.subtitle')}</p>
</header>

<div class="stagger">
<section class="card">
	<h3>{t('home.nextMoves')}</h3>
	<div class="moves">
		{#each moves as m (m.href)}
			{@const Icon = m.icon}
			<a class="move" href={m.href}>
				<span class="mi"><Icon size={20} /></span>
				<span class="mt">
					<span class="title">{t(m.titleKey)}</span>
					<span class="muted sub">{t(m.subKey)}</span>
				</span>
				<ChevronRight size={18} class="cr" />
			</a>
		{/each}
	</div>
</section>

<section class="card">
	<div class="row">
		<h3>{t('home.yourLeagues')}</h3>
		<div class="spacer"></div>
		<a class="pill" href="/leagues">{t('common.manage')}</a>
	</div>
	{#if !loaded}
		<p class="muted">{t('common.loading')}</p>
	{:else if leagues.length === 0}
		<p class="muted">
			{t('home.noLeagues')} <a href="/leagues">{t('home.noLeaguesLink')}</a>
		</p>
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
</div>

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
</style>
