<!-- A themed confirmation modal built on the native <dialog> element, so it
     gets the platform's focus trap, top-layer stacking, Esc-to-close and a real
     ::backdrop for free. Drives off a bindable `open`; the parent runs the
     action in `onconfirm` (showing `busy` while it's in flight) and closes the
     dialog by setting `open = false`. Cancel / Esc / backdrop call `oncancel`. -->
<script lang="ts">
	let {
		open = $bindable(false),
		title = '',
		message = '',
		confirmLabel = 'Confirm',
		cancelLabel = 'Cancel',
		danger = false,
		busy = false,
		onconfirm,
		oncancel
	}: {
		open?: boolean;
		title?: string;
		message?: string;
		confirmLabel?: string;
		cancelLabel?: string;
		danger?: boolean;
		busy?: boolean;
		onconfirm?: () => void;
		oncancel?: () => void;
	} = $props();

	let dialog = $state<HTMLDialogElement | null>(null);

	$effect(() => {
		const d = dialog;
		if (!d) return;
		if (open && !d.open) d.showModal();
		else if (!open && d.open) d.close();
	});

	function cancel() {
		if (busy) return;
		open = false;
		oncancel?.();
	}
</script>

<dialog
	bind:this={dialog}
	class="cd"
	oncancel={(e) => {
		e.preventDefault();
		cancel();
	}}
	onclick={(e) => {
		if (e.target === dialog) cancel();
	}}
>
	<div class="cd-inner">
		{#if title}<h3 class="cd-title">{title}</h3>{/if}
		{#if message}<p class="cd-msg">{message}</p>{/if}
		<div class="cd-actions">
			<button class="btn secondary" onclick={cancel} disabled={busy}>
				{cancelLabel}
			</button>
			<button
				class="btn cd-confirm"
				class:danger
				onclick={() => onconfirm?.()}
				disabled={busy}
			>
				{confirmLabel}
			</button>
		</div>
	</div>
</dialog>

<style>
	.cd {
		border: none;
		padding: 0;
		background: transparent;
		max-width: min(92vw, 26rem);
		color: inherit;
	}
	.cd::backdrop {
		background: rgba(0, 0, 0, 0.55);
		backdrop-filter: blur(2px);
	}
	.cd-inner {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 1.25rem 1.3rem;
		box-shadow: 0 18px 50px rgba(0, 0, 0, 0.45);
	}
	.cd-title {
		margin: 0 0 0.4rem;
		font-size: 1.05rem;
		font-weight: 700;
	}
	.cd-msg {
		margin: 0 0 1.15rem;
		color: var(--muted);
		line-height: 1.45;
	}
	.cd-actions {
		display: flex;
		gap: 0.6rem;
		justify-content: flex-end;
	}
	.cd-actions .btn {
		width: auto;
	}
	.cd-confirm.danger {
		background: var(--danger);
		color: #fff;
		border-color: transparent;
	}

	/* Entrance: gentle scale/fade once the dialog is in the top layer. */
	dialog[open].cd {
		animation: cd-pop 0.14s ease-out;
	}
	@keyframes cd-pop {
		from {
			opacity: 0;
			transform: translateY(6px) scale(0.97);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		dialog[open].cd {
			animation: none;
		}
	}
</style>
