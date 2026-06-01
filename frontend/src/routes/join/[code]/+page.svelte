<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { auth } from '$lib/auth.svelte';
	import { t } from '$lib/i18n.svelte';

	let code = $derived($page.params.code ?? '');
	let leagueName = $state('');
	let phase = $state<'loading' | 'invite' | 'joining' | 'invalid' | 'error'>(
		'loading'
	);

	$effect(() => {
		const c = code;
		if (!c) {
			phase = 'invalid';
			return;
		}
		let cancelled = false;
		(async () => {
			try {
				const lg = await api.invitePreview(c);
				if (cancelled) return;
				leagueName = lg.name;
				if (auth.isAuthed) {
					phase = 'joining';
					const r = await api.joinLeague(c);
					if (!cancelled) goto(`/leagues/${r.id}`);
				} else {
					phase = 'invite';
				}
			} catch {
				if (!cancelled) phase = phase === 'joining' ? 'error' : 'invalid';
			}
		})();
		return () => {
			cancelled = true;
		};
	});
</script>

<div class="auth">
	<h1>{t('auth.wmTips')}</h1>
	<p class="muted">{t('auth.tagline')}</p>

	<div class="card">
		{#if phase === 'loading'}
			<p class="muted">{t('join.checkingInvite')}</p>
		{:else if phase === 'joining'}
			<p class="muted">{@html t('join.joining', { name: leagueName })}</p>
		{:else if phase === 'invite'}
			<p class="kicker">{t('join.invited')}</p>
			<h2 class="lname">{leagueName}</h2>
			<p class="muted">
				{t('join.inviteDesc')}
			</p>
			<a class="btn" href={`/register?invite=${encodeURIComponent(code)}`}>
				{t('auth.createAccount')}
			</a>
			<a
				class="btn secondary"
				href={`/login?invite=${encodeURIComponent(code)}`}
			>
				{t('auth.signIn')}
			</a>
		{:else if phase === 'error'}
			<p class="error">{t('join.joinFailed')}</p>
			<a class="btn secondary" href="/leagues">{t('join.goToLeagues')}</a>
		{:else}
			<p class="error">{t('join.invalidInvite')}</p>
			<a class="btn secondary" href="/">{t('join.goHome')}</a>
		{/if}
	</div>
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
	.lname {
		margin: 0.1rem 0 0.6rem;
		font-size: 1.7rem;
	}
	.card .btn + .btn {
		margin-top: 0.6rem;
	}
</style>
