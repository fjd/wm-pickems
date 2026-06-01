<script lang="ts">
	import { tipsStore, type Match } from '$lib/tips.svelte';
	import TipCard from '$lib/components/TipCard.svelte';
	import GroupStandings from '$lib/components/GroupStandings.svelte';
	import { bestThirds } from '$lib/standings';
	import { collapseOnScroll } from '$lib/actions';
	import { serverClock } from '$lib/serverclock.svelte';
	import { t, stageLabel, locale } from '$lib/i18n.svelte';
	import { LocateFixed } from '@lucide/svelte';
	import { tick } from 'svelte';

	let tab = $state<'all' | 'group' | 'ko'>('all');

	let openId = $state('');

	let thirdsAdv = $derived.by(() => {
		const by: Record<string, Match[]> = {};
		for (const m of tipsStore.matches)
			if (m.stage === 'group') (by[m.groupLetter] ||= []).push(m);
		return bestThirds(Object.values(by), tipsStore.tips);
	});

	$effect(() => {
		if (!tipsStore.loaded) tipsStore.load().catch(() => {});
	});

	let filtered = $derived(
		tipsStore.matches.filter((m) => {
			if (tab === 'group') return m.stage === 'group';
			if (tab === 'ko') return m.stage !== 'group';
			return true;
		})
	);

	let nowId = $derived.by(() => {
		const now = serverClock.now();
		const next = filtered.find(
			(m) => new Date(m.kickoff).getTime() >= now
		);
		return (next ?? filtered[filtered.length - 1])?.id ?? '';
	});

	function goNow() {
		document
			.getElementById(`day-${nowDayIndex}`)
			?.scrollIntoView({ behavior: 'smooth', block: 'start' });
	}

	let days = $derived.by(() => {
		const byKickoff = (a: Match, b: Match) =>
			new Date(a.kickoff).getTime() - new Date(b.kickoff).getTime();
		if (tab === 'group') {
			const byGroup: Record<string, Match[]> = {};
			for (const m of filtered) (byGroup[m.groupLetter] ||= []).push(m);
			return Object.keys(byGroup)
				.sort()
				.map(
					(l) =>
						[t('common.group', { letter: l }), byGroup[l].sort(byKickoff)] as [string, Match[]]
				);
		}
		if (tab === 'ko') {
			const stageOrder = ['R32', 'R16', 'QF', 'SF', '3RD', 'FINAL'];
			const byStage: Record<string, Match[]> = {};
			for (const m of filtered) (byStage[m.stage] ||= []).push(m);
			return stageOrder
				.filter((s) => byStage[s])
				.map(
					(s) =>
						[stageLabel(s), byStage[s].sort(byKickoff)] as [
							string,
							Match[]
						]
				);
		}
		return Object.entries(
			filtered.reduce<Record<string, Match[]>>((acc, m) => {
				const d = new Date(m.kickoff).toLocaleDateString(locale.lang, {
					weekday: 'long',
					day: 'numeric',
					month: 'long'
				});
				(acc[d] ||= []).push(m);
				return acc;
			}, {})
		);
	});

	let nowDayIndex = $derived(
		days.findIndex(([, ms]) => ms.some((m) => m.id === nowId))
	);

	let didAutoScroll = false;
	$effect(() => {
		if (didAutoScroll || !tipsStore.loaded) return;
		const idx = nowDayIndex;
		if (idx < 0) return;
		didAutoScroll = true;
		if (idx === 0) return;
		tick().then(() =>
			document
				.getElementById(`day-${idx}`)
				?.scrollIntoView({ block: 'start' })
		);
	});
</script>

<div class="stickyhead" use:collapseOnScroll>
	<p class="kicker">{t('tips.kicker')}</p>
	<div class="sh-expand">
		<div class="sh-inner">
			<h1>{t('tips.title')}</h1>
			<p class="muted desc">{t('tips.description')}</p>
		</div>
	</div>
	<div class="tabs">
		<button class:active={tab === 'all'} onclick={() => (tab = 'all')}>{t('common.all')}</button>
		<button class:active={tab === 'group'} onclick={() => (tab = 'group')}
			>{t('common.groups')}</button
		>
		<button class:active={tab === 'ko'} onclick={() => (tab = 'ko')}
			>{t('common.knockout')}</button
		>
	</div>
</div>

{#if !tipsStore.loaded}
	<p class="muted">{t('tips.loadingFixtures')}</p>
{:else if filtered.length === 0}
	<p class="muted">{t('tips.nothingHere')}</p>
{:else}
	{#each days as [day, ms], i (day)}
		<h3 class="day" id={`day-${i}`}>{day}</h3>
		{#each ms as m (m.id)}
			<div class="match">
				<TipCard
					match={m}
					open={openId === m.id}
					onToggle={() => (openId = openId === m.id ? '' : m.id)}
				/>
			</div>
		{/each}
		{#if tab === 'group'}
			<GroupStandings matches={ms} bestThirds={thirdsAdv} />
		{/if}
	{/each}
	<div class="fabpad"></div>
{/if}

{#if tipsStore.loaded && nowId}
	<button class="fab" onclick={goNow} aria-label={t('tips.scrollToNext')}>
		<LocateFixed size={18} /> {t('common.now')}
	</button>
{/if}

<style>
	.stickyhead {
		position: sticky;
		top: var(--topbar-h);
		z-index: 20;
		margin: 0 -1rem;
		padding: 0.6rem 1rem 0.75rem;
		background: color-mix(in srgb, var(--bg) 86%, transparent);
		backdrop-filter: blur(12px) saturate(1.3);
		border-bottom: 1px solid var(--border);
	}
	.stickyhead h1 {
		margin: 0.1rem 0 0;
	}
	.stickyhead .desc {
		margin: 0.3rem 0 0;
		font-size: 0.9rem;
	}
	@media (min-width: 900px) {
		.stickyhead {
			top: 0;
			margin: 0 -2rem;
			padding: 0.75rem 2rem 0.85rem;
		}
	}
	.tabs {
		display: flex;
		gap: 0.4rem;
		margin: 0.75rem 0 0;
		z-index: 10;
	}
	.tabs button {
		flex: 1;
		padding: 0.5rem;
		background: var(--surface-2);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		color: var(--muted);
		font-weight: 600;
		font-size: 0.85rem;
	}
	.tabs button.active {
		color: var(--accent-fg);
		background: var(--accent);
		border-color: var(--accent);
	}
	.day {
		margin: 1.3rem 0 0.6rem;
		font-size: 0.95rem;
		color: var(--muted);
		scroll-margin-top: 150px;
	}
	@media (min-width: 900px) {
		.day {
			scroll-margin-top: 96px;
		}
	}
	.match + .match {
		margin-top: 6px;
	}
	.fabpad {
		height: 4rem;
	}
	.fab {
		position: fixed;
		right: 1rem;
		bottom: calc(var(--nav-h) + 1rem);
		z-index: 40;
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
		padding: 0.7rem 1rem;
		border: none;
		border-radius: var(--radius-pill);
		background: var(--accent);
		color: var(--accent-fg);
		font:
			800 0.8rem var(--font);
		letter-spacing: 0.06em;
		text-transform: uppercase;
		cursor: pointer;
		box-shadow: var(--shadow-pop);
		transition:
			transform 0.12s ease,
			box-shadow 0.2s ease;
	}
	.fab:hover {
		transform: translateY(-2px);
		box-shadow: var(--glow);
	}
	@media (min-width: 900px) {
		.fab {
			bottom: 1.5rem;
			right: 1.5rem;
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.fab {
			transition: none;
		}
	}
</style>
