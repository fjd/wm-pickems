<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { page } from '$app/stores';

	let token = $derived($page.params.token ?? '');
	let status = $state<'working' | 'done' | 'failed'>('working');
	let error = $state('');

	// Confirm immediately on landing — the token in the URL is the whole input,
	// there's nothing for the user to fill in.
	$effect(() => {
		if (!token) {
			status = 'failed';
			error = 'This verification link is incomplete.';
			return;
		}
		auth.confirmVerification(token)
			.then(() => (status = 'done'))
			.catch((err: unknown) => {
				status = 'failed';
				error =
					(err as { message?: string })?.message ??
					'This verification link is invalid or has expired.';
			});
	});
</script>

<div class="auth">
	<h1>Email verification</h1>

	{#if status === 'working'}
		<div class="card"><p class="muted wait">Verifying your email…</p></div>
	{:else if status === 'done'}
		<div class="card">
			<p class="ok">Your email is verified ✓</p>
			<p class="muted note">
				You'll now receive the notifications you've enabled — fine-tune them
				anytime in Settings.
			</p>
			<a class="btn" href={auth.isAuthed ? '/' : '/login'}>
				{auth.isAuthed ? 'Back to the app' : 'Sign in'}
			</a>
		</div>
	{:else}
		<div class="card">
			<p class="error">{error}</p>
			<p class="muted note">
				You can request a fresh link from Settings → Notifications.
			</p>
			<a class="btn secondary" href={auth.isAuthed ? '/settings' : '/login'}>
				{auth.isAuthed ? 'Open Settings' : 'Sign in'}
			</a>
		</div>
	{/if}
</div>

<style>
	.auth {
		max-width: 380px;
		margin: 12dvh auto 0;
	}
	h1 {
		margin: 0 0 1.5rem;
		font-size: 1.8rem;
	}
	.wait {
		margin: 0;
	}
	.ok {
		color: var(--success);
		font-size: 0.95rem;
		margin: 0;
	}
	.note {
		margin: 0.5rem 0 1rem;
		font-size: 0.9rem;
	}
	.btn {
		display: inline-block;
		text-decoration: none;
	}
</style>
