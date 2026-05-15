<script lang="ts">
	import { page } from '$app/stores';
	import { ForecastStore, koKey, type KOMatch } from '$lib/forecast.svelte';
	import Flag from '$lib/components/Flag.svelte';
	import { collapseOnScroll } from '$lib/actions';
	import { Check, CircleCheck, X, Trophy, ArrowLeft } from '@lucide/svelte';

	const fs = new ForecastStore();
	let section = $state<'groups' | 'thirds' | 'bracket'>('groups');
	let err = $state('');

	$effect(() => {
		const uid = $page.params.userId;
		if (uid) fs.loadView(uid).catch((e) => (err = e?.message ?? 'Not allowed'));
	});

	const ord = (n: number) =>
		n === 1 ? '1st' : n === 2 ? '2nd' : n === 3 ? '3rd' : `${n}th`;
	const tname = (id: string) => fs.team(id)?.name ?? '';

	const stages = ['R32', 'R16', 'QF', 'SF', '3RD', 'FINAL'];
	const stageName: Record<string, string> = {
		R32: 'Round of 32',
		R16: 'Round of 16',
		QF: 'Quarter-finals',
		SF: 'Semi-finals',
		'3RD': 'Third place',
		FINAL: 'Final'
	};
	let byStage = $derived(
		stages.map((s) => ({
			stage: s,
			matches: fs.knockout.filter((m) => m.stage === s)
		}))
	);
	let finalMatch = $derived(fs.knockout.find((m) => m.stage === 'FINAL'));
	let champion = $derived(finalMatch ? fs.bracket[koKey(finalMatch)] : '');
	let actualThirds = $derived(fs.actualBestThirds());

	function side(m: KOMatch, s: 'home' | 'away') {
		const [h, a] = fs.sides(m);
		const id = s === 'home' ? h : a;
		if (id) return { id, name: tname(id), team: fs.team(id) };
		return {
			id: '',
			name: s === 'home' ? m.homeLabel : m.awayLabel,
			team: undefined
		};
	}
</script>

<button class="muted back" type="button" onclick={() => history.back()}>
	<ArrowLeft size={15} /> Back
</button>

<div class="stickyhead" use:collapseOnScroll>
	<p class="kicker">Forecast</p>
	<div class="sh-expand">
		<div class="sh-inner">
			<h1>{fs.viewName || '…'}</h1>
			<p class="muted desc">Read-only — your friend's tournament call.</p>
		</div>
	</div>
	{#if fs.loaded}
		<div class="seg">
			<button class:on={section === 'groups'} onclick={() => (section = 'groups')}>Groups</button>
			<button class:on={section === 'thirds'} onclick={() => (section = 'thirds')}>Best thirds</button>
			<button class:on={section === 'bracket'} onclick={() => (section = 'bracket')}>Bracket</button>
		</div>
	{/if}
</div>

{#if err}
	<p class="error">{err}</p>
{:else if !fs.loaded}
	<p class="muted">Loading…</p>
{:else if section === 'groups'}
	{#each fs.groups as g (g.letter)}
		<section class="card grp">
			<h3>Group {g.letter}</h3>
			{#each fs.groupOrder[g.letter] as id, i (id)}
				{@const ao = fs.actualOrder(g.letter)}
				{@const apos = ao ? ao.indexOf(id) + 1 : 0}
				{@const exact = ao ? ao[i] === id : null}
				{@const advanced =
					ao &&
					(apos <= 2 ||
						(apos === 3 && (actualThirds?.has(id) ?? false)))}
				{@const scoredAdv =
					advanced && (i < 2 || (i === 2 && !!fs.thirds[g.letter]))}
				{@const state =
					exact === null
						? 'pending'
						: exact
							? 'ok'
							: scoredAdv
								? 'half'
								: 'miss'}
				<div class="trow" class:rwin={state === 'ok'} class:rhalf={state === 'half'} class:rmiss={state === 'miss'}>
					<span class="pos">{i + 1}</span>
					<Flag iso2={fs.team(id)?.iso2 ?? ''} code={fs.team(id)?.fifaCode ?? ''} />
					<span class="nm">{tname(id)}</span>
					<span class="tag">
						{#if state === 'ok'}<span class="ind ok"><Check size={15} /></span>
						{:else if state === 'half'}<span class="apos half">finished {ord(apos)}</span><span class="ind half"><CircleCheck size={15} /></span>
						{:else if state === 'miss'}<span class="apos">finished {ord(apos)}</span><span class="ind no"><X size={15} /></span>
						{:else if i < 2}<span class="pill ok">advances</span>
						{:else if i === 2}<span class="pill">3rd</span>{/if}
					</span>
				</div>
			{/each}
		</section>
	{/each}
{:else if section === 'thirds'}
	<section class="card tlist">
		{#each fs.groups as g (g.letter)}
			{@const tid = fs.groupThird(g.letter)}
			{@const on = !!fs.thirds[g.letter]}
			{@const adv = actualThirds ? actualThirds.has(tid) : null}
			{#if on}
				<div class="trow">
					<span class="gl">{g.letter}</span>
					<Flag iso2={fs.team(tid)?.iso2 ?? ''} code={fs.team(tid)?.fifaCode ?? ''} />
					<span class="nm">{tname(tid) || '—'}</span>
					<span class="spacer"></span>
					{#if adv === true}<span class="ind ok"><Check size={15} /></span>
					{:else if adv === false}<span class="ind no"><X size={15} /></span>{/if}
				</div>
			{/if}
		{/each}
		{#if Object.keys(fs.thirds).length === 0}
			<p class="muted small">No best-third picks.</p>
		{/if}
	</section>
{:else}
	{#if champion}
		<div class="card champ">
			<Trophy size={20} />
			<span class="lbl">Predicted champion</span>
			<Flag iso2={fs.team(champion)?.iso2 ?? ''} code={fs.team(champion)?.fifaCode ?? ''} size={26} />
			<b>{tname(champion)}</b>
		</div>
	{/if}
	{#each byStage as col (col.stage)}
		<h3 class="rname">{stageName[col.stage]}</h3>
		{#each col.matches as m (koKey(m))}
			{@const H = side(m, 'home')}
			{@const A = side(m, 'away')}
			{@const w = fs.bracket[koKey(m)]}
			{@const actAdv =
				m.num > 0
					? fs.advancerOf(m.num)
					: (fs.results.find((r) => r.stage === m.stage && r.finished)
							?.advancer ?? '')}
			{@const bok = actAdv ? w === actAdv : null}
			<div class="bm card" class:rwin={bok === true} class:rmiss={bok === false}>
				<div class="bteam" class:win={w && w === H.id}>
					{#if H.team}<Flag iso2={H.team.iso2} code={H.team.fifaCode} />{/if}
					<span class="bn" class:ph={!H.id}>{H.name}</span>
				</div>
				<span class="vs">vs</span>
				<div class="bteam right" class:win={w && w === A.id}>
					<span class="bn" class:ph={!A.id}>{A.name}</span>
					{#if A.team}<Flag iso2={A.team.iso2} code={A.team.fifaCode} />{/if}
				</div>
				{#if bok === true}<span class="ind ok"><Check size={15} /></span>
				{:else if bok === false}<span class="ind no"><X size={15} /></span>{/if}
			</div>
		{/each}
	{/each}
{/if}

<style>
	.back {
		display: inline-flex;
		align-items: center;
		gap: 0.3rem;
		margin: 0.25rem 0 0.5rem;
		background: none;
		border: none;
		padding: 0;
		font: inherit;
		color: var(--muted);
		cursor: pointer;
	}
	h1 {
		margin: 0.1rem 0 0;
	}
	.desc {
		margin: 0.3rem 0 0;
		font-size: 0.9rem;
	}
	.seg {
		display: flex;
		gap: 0.4rem;
		margin-top: 0.75rem;
	}
	.seg button {
		flex: 1;
		padding: 0.55rem 0.5rem;
		background: var(--surface-2);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		color: var(--muted);
		font-weight: 700;
		font-size: 0.78rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	.seg button.on {
		background: var(--accent);
		color: var(--accent-fg);
		border-color: var(--accent);
	}
	.grp h3,
	.rname {
		margin: 0 0 0.6rem;
	}
	.rname {
		font-family: var(--font-display);
		text-transform: uppercase;
		color: var(--muted);
		margin: 1.4rem 0 0.6rem;
	}
	.trow {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.5rem 0;
		border-top: 1px solid var(--border);
	}
	.trow:nth-child(2) {
		border-top: none;
	}
	.pos {
		width: 1.2rem;
		text-align: center;
		font-weight: 800;
		color: var(--muted);
	}
	.nm {
		flex: 1;
		font-weight: 600;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.tag {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
	}
	.gl {
		display: grid;
		place-items: center;
		width: 24px;
		height: 24px;
		border-radius: 6px;
		background: var(--surface-2);
		font-family: var(--font-display);
		font-size: 0.85rem;
		color: var(--muted);
	}
	.pill.ok {
		color: var(--accent);
		border-color: color-mix(in srgb, var(--accent) 45%, var(--border));
	}
	.ind {
		display: inline-grid;
		place-items: center;
	}
	.ind.ok {
		color: var(--success);
	}
	.ind.no {
		color: var(--danger);
	}
	.ind.half {
		color: var(--gold);
	}
	.apos {
		font-size: 0.72rem;
		font-weight: 700;
		text-transform: uppercase;
		color: var(--muted);
	}
	.apos.half {
		color: var(--gold);
	}
	.trow.rwin {
		border-color: color-mix(in srgb, var(--success) 45%, var(--border));
	}
	.trow.rhalf {
		border-color: color-mix(in srgb, var(--gold) 45%, var(--border));
	}
	.trow.rmiss {
		border-color: color-mix(in srgb, var(--danger) 40%, var(--border));
	}
	.champ {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		color: var(--gold);
		border-color: color-mix(in srgb, var(--gold) 45%, var(--border));
		text-shadow: 0 0 14px color-mix(in srgb, var(--gold) 55%, transparent);
	}
	.champ .lbl {
		text-transform: uppercase;
		letter-spacing: 0.14em;
		font-size: 0.78rem;
		font-weight: 700;
	}
	.champ b {
		font-family: var(--font-display);
		font-size: 1.15rem;
	}
	.bm {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.6rem 0.8rem;
	}
	.bm + .bm {
		margin-top: 0.5rem;
	}
	.bteam {
		flex: 1;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		min-width: 0;
	}
	.bteam.right {
		justify-content: flex-end;
	}
	.bteam.win .bn {
		color: var(--accent);
		font-weight: 700;
	}
	.bn {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		font-weight: 600;
		font-size: 0.9rem;
	}
	.bn.ph {
		color: var(--muted);
		font-weight: 500;
	}
	.vs {
		color: var(--muted);
		font-size: 0.8rem;
	}
	.small {
		font-size: 0.85rem;
	}
</style>
