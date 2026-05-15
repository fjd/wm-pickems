<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { goto } from '$app/navigation';

	let identity = $state('');
	let password = $state('');
	let error = $state('');
	let busy = $state(false);

	async function submit(e: Event) {
		e.preventDefault();
		error = '';
		busy = true;
		try {
			await auth.login(identity, password);
			goto('/');
		} catch {
			error = 'Invalid email or password.';
		} finally {
			busy = false;
		}
	}
</script>

<div class="auth">
	<h1>WM Tips</h1>
	<p class="muted">Predict the World Cup. Beat your friends.</p>

	<form class="card" onsubmit={submit}>
		<div class="field">
			<label for="id">Email</label>
			<input
				id="id"
				class="input"
				type="email"
				bind:value={identity}
				autocomplete="email"
				required
			/>
		</div>
		<div class="field">
			<label for="pw">Password</label>
			<input
				id="pw"
				class="input"
				type="password"
				bind:value={password}
				autocomplete="current-password"
				required
			/>
		</div>
		{#if error}<p class="error">{error}</p>{/if}
		<button class="btn" disabled={busy}>{busy ? 'Signing in…' : 'Sign in'}</button>
		<p class="muted switch">
			No account? <a href="/register">Create one</a>
		</p>
	</form>
</div>

<style>
	.auth {
		max-width: 380px;
		margin: 12dvh auto 0;
	}
	h1 {
		margin: 0;
		font-size: 2rem;
	}
	.muted {
		margin: 0.25rem 0 1.5rem;
	}
	.switch {
		text-align: center;
		margin: 1rem 0 0;
	}
</style>
