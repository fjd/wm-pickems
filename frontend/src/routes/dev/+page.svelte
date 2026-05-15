<script lang="ts">
	import { pb } from '$lib/pb';
	import { serverClock } from '$lib/serverclock.svelte';

	let when = $state('');
	let busy = $state(false);
	let msg = $state('');

	$effect(() => {
		serverClock.refresh();
	});

	// Seed the input from the current sim time (or now).
	$effect(() => {
		if (!when) {
			const base = serverClock.simTime
				? new Date(serverClock.simTime)
				: new Date(serverClock.now());
			when = base.toISOString().slice(0, 16);
		}
	});

	const presets: { label: string; ts: string }[] = [
		{ label: 'Opening match', ts: '2026-06-11T20:00' },
		{ label: 'Group MD2 live', ts: '2026-06-15T21:30' },
		{ label: 'After groups', ts: '2026-06-25T06:00' },
		{ label: 'After R32', ts: '2026-07-04T06:00' },
		{ label: 'After QF', ts: '2026-07-12T06:00' },
		{ label: 'After final', ts: '2026-07-20T00:00' }
	];

	async function advance(ts: string) {
		busy = true;
		msg = '';
		try {
			await pb.send('/api/dev/advance', {
				method: 'POST',
				body: { timestamp: ts }
			});
			location.reload(); // re-pull all stores against the new clock
		} catch (e: unknown) {
			msg = (e as { message?: string })?.message ?? 'Failed';
			busy = false;
		}
	}

	async function reset() {
		busy = true;
		msg = '';
		try {
			await pb.send('/api/dev/reset', { method: 'POST', body: {} });
			location.reload();
		} catch (e: unknown) {
			msg = (e as { message?: string })?.message ?? 'Failed';
			busy = false;
		}
	}
</script>

<p class="kicker">Test harness</p>
<h1>Dev tools</h1>

{#if !serverClock.loaded}
	<p class="muted">…</p>
{:else if !serverClock.dev}
	<section class="card">
		<p class="muted">
			Disabled. Start the server with <code>WMP_DEV=1</code> to simulate the
			tournament.
		</p>
	</section>
{:else}
	<section class="card">
		<div class="state">
			<span class="kicker">Simulated clock</span>
			<b class="digits"
				>{serverClock.simulated
					? new Date(serverClock.now()).toLocaleString()
					: 'live (real time)'}</b
			>
		</div>
	</section>

	<section class="card">
		<h3>Advance to</h3>
		<p class="muted small">
			Matches before this moment are simulated (finished, or <b>live</b> if
			mid-match); later ones reset. Locks, friends'-tips and the Forecast
			deadline follow this clock.
		</p>
		<div class="field">
			<input class="input" type="datetime-local" bind:value={when} />
		</div>
		<button
			class="btn"
			disabled={busy || !when}
			onclick={() => advance(when)}>Advance</button
		>

		<div class="presets">
			{#each presets as p (p.ts)}
				<button
					class="chip"
					disabled={busy}
					onclick={() => advance(p.ts)}>{p.label}</button
				>
			{/each}
		</div>
	</section>

	<section class="card">
		<h3>Reset</h3>
		<p class="muted small">
			Clear all results and the simulated clock (back to real time).
		</p>
		<button class="btn secondary" disabled={busy} onclick={reset}
			>Reset everything</button
		>
	</section>

	{#if msg}<p class="error">{msg}</p>{/if}
{/if}

<style>
	h1 {
		margin: 0.1rem 0 1rem;
	}
	.small {
		font-size: 0.85rem;
	}
	.state {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
	}
	.state b {
		font-size: 1.2rem;
	}
	.presets {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin-top: 0.9rem;
	}
	.chip {
		padding: 0.5rem 0.8rem;
		background: var(--surface-2);
		border: 1px solid var(--border);
		border-radius: var(--radius-pill);
		color: var(--text);
		font:
			700 0.78rem var(--font);
		text-transform: uppercase;
		letter-spacing: 0.04em;
		cursor: pointer;
	}
	.chip:hover {
		border-color: var(--accent);
	}
	code {
		font-family: var(--font-mono);
		color: var(--accent);
	}
</style>
