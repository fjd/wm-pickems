<script lang="ts">
	// Nudges email/password users to verify their address — unverified accounts
	// receive no email notifications (deliverability protection: we never send
	// to unconfirmed, possibly mistyped addresses). Mirrors NotifyAnnounce:
	// closing hides it for the session but it returns on the next visit; only
	// "Don't show again" persists the dismissal. Never shows once verified.
	import { auth } from '$lib/auth.svelte';
	import { MailCheck, X } from '@lucide/svelte';

	const KEY = 'verify-announce-v1';

	let dismissed = $state(true);
	let dontShowAgain = $state(false);
	if (typeof localStorage !== 'undefined') {
		try {
			dismissed = localStorage.getItem(KEY) === '1';
		} catch {
			dismissed = false;
		}
	}

	let busy = $state(false);
	let sent = $state(false);
	let error = $state('');

	function close() {
		if (dontShowAgain) {
			try {
				localStorage.setItem(KEY, '1');
			} catch {
				/* ignore (private mode) */
			}
		}
		dismissed = true;
	}

	// Auto-dismiss once the account gets verified (e.g. in another tab).
	$effect(() => {
		if (auth.user?.verified) dismissed = true;
	});

	let open = $derived(!!auth.user && !auth.user.verified && !dismissed);

	async function send() {
		error = '';
		busy = true;
		try {
			await auth.requestVerification();
			sent = true;
		} catch (err: unknown) {
			error =
				(err as { message?: string })?.message ??
				'Could not send the verification email.';
		} finally {
			busy = false;
		}
	}
</script>

{#if open}
	<button
		type="button"
		class="backdrop"
		aria-label="Close"
		onclick={close}
	></button>
	<div class="sheet" role="dialog" aria-label="Verify your email">
		<button class="x" aria-label="Dismiss" onclick={close}><X size={16} /></button>

		<div class="icon"><MailCheck size={22} /></div>
		<h3>Verify your email 📬</h3>
		<p class="body">
			Kickoff reminders, matchday recaps and league alerts only go to
			<strong>verified</strong> addresses. Confirm
			<strong>{auth.user?.email}</strong> to keep receiving them — check your
			inbox for the link, or send a fresh one below.
		</p>

		{#if sent}
			<p class="ok">Verification email sent — check your inbox.</p>
		{:else}
			<button class="btn" onclick={send} disabled={busy}>
				{busy ? 'Sending…' : 'Send verification email'}
			</button>
		{/if}

		{#if error}<p class="err">{error}</p>{/if}

		<label class="dsa">
			<input type="checkbox" bind:checked={dontShowAgain} />
			<span>Don't show this again</span>
		</label>

		<div class="foot">
			<a href="/settings" onclick={close}>Manage in Settings</a>
			<button class="later" onclick={close}>Maybe later</button>
		</div>
	</div>
{/if}

<style>
	.backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.55);
		border: none;
		padding: 0;
		z-index: 70;
		cursor: pointer;
	}
	.sheet {
		position: fixed;
		inset: auto 0.75rem calc(var(--nav-h, 0px) + 0.75rem) 0.75rem;
		z-index: 71;
		max-width: 420px;
		margin: 0 auto;
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 1.25rem 1.25rem 1.1rem;
		box-shadow: var(--shadow-pop);
	}
	@media (min-width: 640px) {
		.sheet {
			inset: 50% auto auto 50%;
			transform: translate(-50%, -50%);
		}
	}
	.icon {
		display: inline-grid;
		place-items: center;
		width: 44px;
		height: 44px;
		border-radius: 999px;
		background: color-mix(in srgb, var(--accent) 18%, var(--surface-2));
		color: var(--accent);
		margin-bottom: 0.6rem;
	}
	h3 {
		margin: 0 0 0.5rem;
		font-size: 1.2rem;
	}
	.body {
		margin: 0 0 1.1rem;
		font-size: 0.92rem;
		line-height: 1.55;
		color: var(--muted);
	}
	.body strong {
		color: var(--text);
	}
	.btn {
		width: 100%;
	}
	.ok {
		margin: 0;
		font-size: 0.92rem;
		color: var(--success);
	}
	.err {
		margin: 0.6rem 0 0;
		font-size: 0.82rem;
		color: var(--danger);
	}
	.dsa {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-top: 1rem;
		font-size: 0.85rem;
		color: var(--muted);
		cursor: pointer;
	}
	.dsa input {
		width: 16px;
		height: 16px;
		accent-color: var(--accent);
		cursor: pointer;
	}
	.foot {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-top: 0.85rem;
	}
	.foot a {
		font-size: 0.85rem;
		color: var(--muted);
	}
	.later {
		background: transparent;
		border: none;
		color: var(--muted);
		font-size: 0.85rem;
		cursor: pointer;
		padding: 0.3rem 0;
	}
	.later:hover {
		color: var(--text);
	}
	.x {
		position: absolute;
		top: 0.6rem;
		right: 0.6rem;
		display: inline-grid;
		place-items: center;
		width: 32px;
		height: 32px;
		border-radius: 999px;
		background: transparent;
		color: var(--muted);
		border: none;
		cursor: pointer;
	}
	.x:hover {
		color: var(--text);
		background: var(--surface-2);
	}
</style>
