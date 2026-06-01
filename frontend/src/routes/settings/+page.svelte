<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { t, locale } from '$lib/i18n.svelte';
	import Avatar from '$lib/components/Avatar.svelte';
	import { SUPPORTED_LANGS, type Lang } from '$lib/translations/index';

	const MAX_AVATAR_BYTES = 5 * 1024 * 1024;

	let name = $state(auth.user?.name ?? '');
	let avatarFile = $state<File | null>(null);
	let previewUrl = $state<string | null>(null);
	let error = $state('');
	let saved = $state(false);
	let busy = $state(false);
	let fileInput: HTMLInputElement;

	let resetBusy = $state(false);
	let resetSent = $state(false);
	let resetError = $state('');

	const languages: { code: Lang; label: string }[] = [
		{ code: 'en', label: 'English' },
		{ code: 'de', label: 'Deutsch' }
	];

	async function sendReset() {
		if (!auth.user?.email) return;
		resetError = '';
		resetSent = false;
		resetBusy = true;
		try {
			await auth.requestPasswordReset(auth.user.email);
			resetSent = true;
		} catch (err: unknown) {
			resetError =
				(err as { message?: string })?.message ??
				t('errors.resetEmailFailed');
		} finally {
			resetBusy = false;
		}
	}

	$effect(() => {
		const url = previewUrl;
		return () => {
			if (url) URL.revokeObjectURL(url);
		};
	});

	function pickFile(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		if (!file.type.startsWith('image/')) {
			error = t('errors.chooseImage');
			return;
		}
		if (file.size > MAX_AVATAR_BYTES) {
			error = t('errors.imageTooLarge');
			return;
		}
		error = '';
		saved = false;
		avatarFile = file;
		previewUrl = URL.createObjectURL(file);
	}

	async function submit(e: Event) {
		e.preventDefault();
		error = '';
		saved = false;
		const trimmed = name.trim();
		if (trimmed.length < 1 || trimmed.length > 48) {
			error = t('errors.nameLength');
			return;
		}
		busy = true;
		try {
			await auth.updateProfile({ name: trimmed, avatarFile });
			avatarFile = null;
			previewUrl = null;
			if (fileInput) fileInput.value = '';
			saved = true;
		} catch (err: unknown) {
			error =
				(err as { message?: string })?.message ??
				t('errors.saveFailed');
		} finally {
			busy = false;
		}
	}
</script>

<div class="settings">
	<h1>{t('settings.title')}</h1>
	<p class="muted">{t('settings.subtitle')}</p>

	<section class="card">
		<h3>{t('settings.language')}</h3>
		<div class="lang-options">
			{#each languages as l (l.code)}
				<button
					class="btn {locale.lang === l.code ? '' : 'secondary'}"
					onclick={() => locale.set(l.code)}
				>
					{l.label}
				</button>
			{/each}
		</div>
	</section>

	<form class="card" onsubmit={submit}>
		<div class="avatar-row">
			<Avatar
				name={name || auth.user?.name || '?'}
				src={previewUrl ?? auth.user?.avatarUrl}
				size={96}
			/>
			<div>
				<button
					type="button"
					class="btn secondary"
					onclick={() => fileInput.click()}
					disabled={busy}
				>
					{t('settings.changePhoto')}
				</button>
				<p class="muted hint">{t('settings.photoHint')}</p>
			</div>
			<input
				bind:this={fileInput}
				type="file"
				accept="image/*"
				class="hidden-file"
				onchange={pickFile}
			/>
		</div>

		<div class="field">
			<label for="dn">{t('settings.displayName')}</label>
			<input
				id="dn"
				class="input"
				bind:value={name}
				maxlength="48"
				autocomplete="name"
				required
			/>
		</div>

		{#if error}<p class="error">{error}</p>{/if}
		{#if saved}<p class="ok">{t('settings.saved')}</p>{/if}

		<button class="btn" disabled={busy}>{busy ? t('common.saving') : t('settings.saveChanges')}</button>
	</form>

	<section class="card">
		<h3>{t('settings.passwordSection')}</h3>
		<p class="muted small">
			{@html t('settings.passwordResetDesc', { email: auth.user?.email ?? '' })}
		</p>
		{#if resetError}<p class="error">{resetError}</p>{/if}
		{#if resetSent}
			<p class="ok">{t('settings.resetSent')}</p>
		{/if}
		<button
			type="button"
			class="btn secondary"
			onclick={sendReset}
			disabled={resetBusy || resetSent}
		>
			{resetBusy ? t('settings.sending') : resetSent ? t('settings.sent') : t('settings.sendResetLink')}
		</button>
	</section>

	<p class="muted switch"><a href="/">{t('common.back')}</a></p>
</div>

<style>
	.settings {
		max-width: 380px;
		margin: 8dvh auto 0;
	}
	h1 {
		margin: 0;
		font-size: 1.8rem;
	}
	.muted {
		margin: 0.25rem 0 1.5rem;
	}
	.lang-options {
		display: flex;
		gap: 0.5rem;
	}
	.avatar-row {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-bottom: 1.25rem;
	}
	.hint {
		margin: 0.5rem 0 0;
		font-size: 0.8rem;
	}
	.hidden-file {
		display: none;
	}
	.ok {
		color: var(--success);
		font-size: 0.9rem;
	}
	.small {
		font-size: 0.85rem;
		margin: 0.25rem 0 0.9rem;
	}
	h3 {
		margin: 0 0 0.5rem;
		font-size: 1rem;
	}
	.switch {
		text-align: center;
		margin: 1rem 0 0;
	}
</style>
