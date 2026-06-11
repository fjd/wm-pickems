<script lang="ts">
	import { pb } from '$lib/pb';
	import { serverClock } from '$lib/serverclock.svelte';
	import { api, type LeagueSummary } from '$lib/api';
	import { t, locale } from '$lib/i18n.svelte';

	let when = $state('');
	let busy = $state(false);
	let msg = $state('');

	let botCount = $state(3);
	let botLeague = $state('');
	let leagues = $state<LeagueSummary[]>([]);

	$effect(() => {
		if (serverClock.dev)
			api
				.myLeagues()
				.then((r) => (leagues = r.leagues))
				.catch(() => {});
	});

	async function genBots() {
		busy = true;
		msg = '';
		try {
			await pb.send('/api/dev/bots', {
				method: 'POST',
				body: { count: botCount, leagueId: botLeague }
			});
			location.reload();
		} catch (e: unknown) {
			msg = (e as { message?: string })?.message ?? 'Failed';
			busy = false;
		}
	}

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

	const sampleEvents = [
		{ key: 'tips_reminder', label: '⚽ Tip reminder' },
		{ key: 'results_recap', label: '🏆 Results recap' },
		{ key: 'stage_starting', label: '🏟 Stage starting' },
		{ key: 'forecast_reminder', label: '⏰ Forecast deadline' },
		{ key: 'league_lead', label: '🥇 Took the lead' },
		{ key: 'kickoff_countdown', label: '📅 Countdown to kickoff' }
	];

	async function sendSample(event: string) {
		busy = true;
		msg = '';
		try {
			const r = await pb.send<{ sent: number; total: number }>(
				'/api/dev/push/sample',
				{ method: 'POST', body: { event } }
			);
			msg = `Sent to ${r.sent}/${r.total} device(s) — watch for the notification.`;
		} catch (e: unknown) {
			msg =
				(e as { message?: string })?.message ??
				'Failed — is push enabled on this device?';
		} finally {
			busy = false;
		}
	}

	async function sendEmail(event: string) {
		busy = true;
		msg = '';
		try {
			const r = await pb.send<{ to: string; provider: string }>(
				'/api/dev/notify/email',
				{ method: 'POST', body: { event } }
			);
			msg =
				r.provider === 'log'
					? `Provider is "log" — nothing delivered. Set a mail provider to actually send.`
					: `Email sent to ${r.to} via ${r.provider} — check your inbox.`;
		} catch (e: unknown) {
			msg = (e as { message?: string })?.message ?? 'Failed to send email.';
		} finally {
			busy = false;
		}
	}

	// ---- API-Football diagnostic ----
	type ApiCheck = {
		account?: {
			subscription?: { plan?: string; active?: boolean; end?: string };
			requests?: { current?: number; limit_day?: number };
		};
		statusError?: string;
		season?: number;
		fixtures?: number;
		statusHistogram?: Record<string, number>;
		unmappedTeams?: string[];
		ourMatchesTotal?: number;
		ourMatchesMapped?: number;
		withExtraTime?: number;
		withPenalties?: number;
	};
	let apiSeason = $state(2026);
	let apiResult = $state<ApiCheck | null>(null);
	let apiBusy = $state(false);
	let apiErr = $state('');

	async function runApiCheck(season: number) {
		apiBusy = true;
		apiErr = '';
		apiResult = null;
		try {
			apiResult = await pb.send<ApiCheck>(`/api/dev/apicheck?season=${season}`, {
				method: 'GET'
			});
		} catch (e: unknown) {
			const err = e as { data?: { error?: string }; message?: string };
			apiErr = err.data?.error ?? err.message ?? 'apicheck failed';
		} finally {
			apiBusy = false;
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

<p class="kicker">{t('dev.testHarness')}</p>
<h1>{t('dev.title')}</h1>

{#if !serverClock.loaded}
	<p class="muted">…</p>
{:else if !serverClock.dev}
	<section class="card">
		<p class="muted">
			{@html t('dev.disabled')}
		</p>
	</section>
{:else}
	<section class="card">
		<div class="state">
			<span class="kicker">{t('dev.simulatedClock')}</span>
			<b class="digits"
				>{serverClock.simulated
					? new Date(serverClock.now()).toLocaleString(locale.lang)
					: t('dev.liveRealTime')}</b
			>
		</div>
	</section>

	<section class="card">
		<h3>{t('dev.advanceTo')}</h3>
		<p class="muted small">
			{@html t('dev.advanceDesc')}
		</p>
		<div class="field">
			<input class="input" type="datetime-local" bind:value={when} />
		</div>
		<button
			class="btn"
			disabled={busy || !when}
			onclick={() => advance(when)}>{t('dev.advance')}</button
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
		<h3>{t('dev.generateBots')}</h3>
		<p class="muted small">
			{t('dev.botsDesc')}
		</p>
		<div class="field">
			<label for="bc">{t('dev.howMany')}</label>
			<input
				id="bc"
				class="input"
				type="number"
				min="1"
				max="20"
				bind:value={botCount}
			/>
		</div>
		<div class="field">
			<label for="bl">{t('dev.league')}</label>
			<select id="bl" class="input" bind:value={botLeague}>
				<option value="">{t('dev.allMyLeagues')}</option>
				{#each leagues as l (l.id)}
					<option value={l.id}>{l.name}</option>
				{/each}
			</select>
		</div>
		<button class="btn" disabled={busy} onclick={genBots}>
			{t('dev.generate', { count: botCount, plural: botCount === 1 ? '' : 's' })}
		</button>
	</section>

	<section class="card">
		<h3>{t('dev.reset')}</h3>
		<p class="muted small">
			{t('dev.testPushDesc')}
		</p>
		<div class="presets">
			{#each sampleEvents as s (s.key)}
				<button class="chip" disabled={busy} onclick={() => sendSample(s.key)}
					>{s.label}</button
				>
			{/each}
		</div>
	</section>

	<section class="card">
		<h3>{t('dev.testEmails')}</h3>
		<p class="muted small">
			{t('dev.testEmailsDesc')}
		</p>
		<div class="presets">
			{#each sampleEvents as s (s.key)}
				<button class="chip" disabled={busy} onclick={() => sendEmail(s.key)}
					>{s.label}</button
				>
			{/each}
		</div>
	</section>

	<section class="card">
		<h3>API-Football check</h3>
		<p class="muted small">
			Read-only probe of the live-results API (uses <code>API_FOOTBALL_KEY</code
			>). Point at 2026 to confirm your plan covers it, or a finished season
			(2022) to exercise the results/ET/penalty path.
		</p>
		<div class="apirow">
			<input
				class="input season"
				type="number"
				bind:value={apiSeason}
				min="2018"
				max="2030"
				aria-label="Season"
			/>
			<button class="btn" disabled={apiBusy} onclick={() => runApiCheck(apiSeason)}>
				{apiBusy ? 'Checking…' : 'Run check'}
			</button>
			<button
				class="chip"
				disabled={apiBusy}
				onclick={() => {
					apiSeason = 2022;
					runApiCheck(2022);
				}}>2022 (finished)</button
			>
		</div>

		{#if apiErr}<p class="error">{apiErr}</p>{/if}

		{#if apiResult}
			{@const r = apiResult}
			{@const ok = (r.fixtures ?? 0) > 0}
			<div class="verdict" class:good={ok} class:warn={!ok}>
				{#if ok}
					✓ Reached the API — {r.fixtures} fixture(s) for season {r.season}.
				{:else}
					No fixtures for season {r.season}. Your plan likely doesn't cover it
					(stays on openfootball).
				{/if}
			</div>

			{#if r.account?.subscription}
				<p class="muted small acct">
					Plan: <b>{r.account.subscription.plan ?? '—'}</b>
					{#if r.account.requests}
						· requests today: {r.account.requests.current ?? 0}/{r.account
							.requests.limit_day ?? '—'}
					{/if}
				</p>
			{:else if r.statusError}
				<p class="muted small acct">Account status unavailable: {r.statusError}</p>
			{/if}

			<div class="kv">
				<span><i>Fixtures</i><b>{r.fixtures ?? 0}</b></span>
				<span
					><i>Our matches mapped</i><b
						class:warn={r.ourMatchesMapped !== r.ourMatchesTotal}
						>{r.ourMatchesMapped ?? 0}/{r.ourMatchesTotal ?? 0}</b
					></span
				>
				<span><i>With extra time</i><b>{r.withExtraTime ?? 0}</b></span>
				<span><i>With penalties</i><b>{r.withPenalties ?? 0}</b></span>
			</div>

			{#if r.statusHistogram && Object.keys(r.statusHistogram).length}
				<div class="hist">
					{#each Object.entries(r.statusHistogram) as [k, v] (k)}
						<span class="hchip">{k}<b>{v}</b></span>
					{/each}
				</div>
			{/if}

			{#if r.unmappedTeams?.length}
				<p class="error small">
					Unmapped teams ({r.unmappedTeams.length}): {r.unmappedTeams.join(', ')}
				</p>
			{/if}

			<details class="raw">
				<summary>Raw response</summary>
				<pre>{JSON.stringify(r, null, 2)}</pre>
			</details>
		{/if}
	</section>

	<section class="card">
		<h3>{t('dev.reset')}</h3>
		<p class="muted small">
			{t('dev.resetDesc')}
		</p>
		<button class="btn secondary" disabled={busy} onclick={reset}
			>{t('dev.resetEverything')}</button
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

	/* ---- API-Football check ---- */
	.apirow {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.5rem;
		margin-top: 0.9rem;
	}
	.season {
		width: 6rem;
		flex: none;
	}
	.apirow .btn {
		width: auto;
	}
	.verdict {
		margin-top: 0.9rem;
		padding: 0.6rem 0.8rem;
		border-radius: var(--radius-sm);
		border: 1px solid var(--border);
		font-weight: 600;
		font-size: 0.9rem;
	}
	.verdict.good {
		color: var(--success);
		border-color: color-mix(in srgb, var(--success) 45%, var(--border));
		background: color-mix(in srgb, var(--success) 8%, transparent);
	}
	.verdict.warn {
		color: var(--warning);
		border-color: color-mix(in srgb, var(--warning) 45%, var(--border));
		background: color-mix(in srgb, var(--warning) 8%, transparent);
	}
	.acct {
		margin: 0.6rem 0 0;
	}
	.kv {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.4rem 1rem;
		margin-top: 0.8rem;
	}
	.kv span {
		display: flex;
		justify-content: space-between;
		gap: 0.6rem;
		padding: 0.35rem 0;
		border-bottom: 1px solid var(--border);
	}
	.kv i {
		color: var(--muted);
		font-style: normal;
		font-size: 0.85rem;
	}
	.kv b {
		font-family: var(--font-mono);
	}
	.kv b.warn {
		color: var(--warning);
	}
	.hist {
		display: flex;
		flex-wrap: wrap;
		gap: 0.4rem;
		margin-top: 0.8rem;
	}
	.hchip {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		padding: 0.2rem 0.55rem;
		background: var(--surface-2);
		border: 1px solid var(--border);
		border-radius: var(--radius-pill);
		font-size: 0.78rem;
		color: var(--muted);
	}
	.hchip b {
		font-family: var(--font-mono);
		color: var(--text);
	}
	.raw {
		margin-top: 0.85rem;
	}
	.raw summary {
		cursor: pointer;
		color: var(--muted);
		font-size: 0.82rem;
	}
	.raw pre {
		margin: 0.6rem 0 0;
		padding: 0.7rem;
		background: var(--surface-2);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		overflow: auto;
		max-height: 320px;
		font-size: 0.78rem;
		line-height: 1.5;
	}
</style>
