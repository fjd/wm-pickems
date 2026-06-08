<script lang="ts">
	// Announcement that the notifications feature exists, nudging the user to
	// install (if needed) and turn on push on this device. Closing hides it for
	// the session but it returns on the next visit, so it can't be lost before
	// it's read — only ticking "Don't show again" persists the dismissal
	// (localStorage). Once push is enabled it never shows again regardless.
	import { push } from '$lib/push.svelte';
	import { pwa } from '$lib/pwa.svelte';
	import { t } from '$lib/i18n.svelte';
	import { BellRing, X } from '@lucide/svelte';

	const KEY = 'notify-announce-v1';

	let dismissed = $state(true);
	let dontShowAgain = $state(false);
	if (typeof localStorage !== 'undefined') {
		try {
			dismissed = localStorage.getItem(KEY) === '1';
		} catch {
			dismissed = false;
		}
	}

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

	// Auto-dismiss once push is enabled on this device.
	$effect(() => {
		if (push.subscribed) close();
	});

	// Show only after the subscription check settled, and only if not already
	// subscribed/dismissed.
	let open = $derived(push.ready && !dismissed && !push.subscribed);

	async function enable() {
		await push.enable();
		if (push.subscribed) close();
	}
</script>

{#if open}
	<button
		type="button"
		class="backdrop"
		aria-label="Close"
		onclick={close}
	></button>
	<div class="sheet" role="dialog" aria-label={t('notify.title')}>
		<button class="x" aria-label={t('common.dismiss')} onclick={close}><X size={16} /></button>

		<div class="icon"><BellRing size={22} /></div>
		<h3>{t('notify.title')}</h3>
		<p class="body">
			{@html t('notify.body')}
		</p>

		{#if push.supported && !push.blocked}
			<button class="btn" onclick={enable} disabled={push.busy}>
				{push.busy ? t('common.saving') : t('notify.enablePush')}
			</button>
		{:else if !pwa.installed}
			<button class="btn" onclick={() => pwa.install()}>
				{t('notify.installForPush')}
			</button>
			<p class="hint muted">
				{t('notify.iphoneHint')}
			</p>
		{:else if push.blocked}
			<p class="hint muted">
				{t('notify.pushBlocked')}
			</p>
		{/if}

		{#if push.error}<p class="err">{push.error}</p>{/if}

		<label class="dsa">
			<input type="checkbox" bind:checked={dontShowAgain} />
			<span>{t('notify.dontShowAgain')}</span>
		</label>

		<div class="foot">
			<a href="/settings" onclick={close}>{t('notify.fineTuneSettings')}</a>
			<button class="later" onclick={close}>{t('notify.maybeLater')}</button>
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
	.hint {
		margin: 0.7rem 0 0;
		font-size: 0.82rem;
		line-height: 1.45;
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
