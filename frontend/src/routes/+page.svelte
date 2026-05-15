<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { api, type LeagueSummary } from '$lib/api';

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
</script>

<header>
	<div class="row">
		<div>
			<h1>Hi, {auth.user?.name} 👋</h1>
			<p class="muted">World Cup 2026 · 11 Jun – 19 Jul</p>
		</div>
		<div class="spacer"></div>
		<button class="btn ghost logout" onclick={() => auth.logout()}>Log out</button>
	</div>
</header>

<section class="card">
	<h3>Your next moves</h3>
	<ul class="todo">
		<li><a href="/forecast">🔮 Fill in your tournament Forecast</a> <span class="muted">— before the opening match</span></li>
		<li><a href="/tips">🎯 Tip the upcoming matches</a> <span class="muted">— editable until kickoff</span></li>
		<li><a href="/leagues">🏆 Create or join a League</a> <span class="muted">— play against friends</span></li>
	</ul>
</section>

<section class="card">
	<div class="row">
		<h3>Your leagues</h3>
		<div class="spacer"></div>
		<a class="pill" href="/leagues">Manage</a>
	</div>
	{#if !loaded}
		<p class="muted">Loading…</p>
	{:else if leagues.length === 0}
		<p class="muted">You're not in a league yet. <a href="/leagues">Create or join one →</a></p>
	{:else}
		{#each leagues as l (l.id)}
			<a class="lrow" href={`/leagues/${l.id}`}>
				<span>{l.name}</span>
				<span class="spacer"></span>
				<span class="pill">{l.members} 👥</span>
			</a>
		{/each}
	{/if}
</section>

<style>
	header {
		margin: 1rem 0 1.25rem;
	}
	h1 {
		margin: 0;
		font-size: 1.6rem;
	}
	.muted {
		margin: 0.2rem 0 0;
	}
	.logout {
		width: auto;
	}
	.todo {
		margin: 0.5rem 0 0;
		padding-left: 1.1rem;
		line-height: 1.9;
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
</style>
