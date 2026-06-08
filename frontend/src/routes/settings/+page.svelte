<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { t, locale } from '$lib/i18n.svelte';
	import { push } from '$lib/push.svelte';
	import { goto } from '$app/navigation';
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

	// Notification preferences. Each event defaults to ON when no pref is stored.
	const NOTIFY_EVENTS = [
		{
			key: 'kickoff_countdown',
			labelKey: 'settings.notifyKickoffCountdown',
			hintKey: 'settings.notifyKickoffCountdownHint'
		},
		{
			key: 'stage_starting',
			labelKey: 'settings.notifyStageStarting',
			hintKey: 'settings.notifyStageStartingHint'
		},
		{
			key: 'tips_reminder',
			labelKey: 'settings.notifyTipsReminder',
			hintKey: 'settings.notifyTipsReminderHint'
		},
		{
			key: 'forecast_reminder',
			labelKey: 'settings.notifyForecastReminder',
			hintKey: 'settings.notifyForecastReminderHint'
		},
		{
			key: 'results_recap',
			labelKey: 'settings.notifyResultsRecap',
			hintKey: 'settings.notifyResultsRecapHint'
		},
		{
			key: 'league_lead',
			labelKey: 'settings.notifyLeagueLead',
			hintKey: 'settings.notifyLeagueLeadHint'
		}
	];

	type Channel = 'email' | 'push';
	let prefs = $state<Record<string, { email?: boolean; push?: boolean }>>({
		...(auth.user?.notifyPrefs ?? {})
	});
	let notifyBusy = $state(false);
	let notifyError = $state('');

	// Absent pref defaults to ON (matches the backend default-on semantics).
	const isOn = (key: string, ch: Channel) => prefs[key]?.[ch] !== false;

	let testMsg = $state('');
	let testBusy = $state(false);
	async function sendTest() {
		testMsg = '';
		testBusy = true;
		try {
			const { sent, total } = await push.test();
			testMsg =
				sent > 0
					? t('settings.testSent', { sent, total })
					: t('settings.testNoDevice', { total });
		} catch (e: unknown) {
			testMsg = (e as { message?: string })?.message ?? 'Test failed.';
		} finally {
			testBusy = false;
		}
	}

	async function toggleNotify(key: string, ch: Channel) {
		const next = {
			...prefs,
			[key]: { ...prefs[key], [ch]: !isOn(key, ch) }
		};
		const prev = prefs;
		prefs = next;
		notifyError = '';
		notifyBusy = true;
		try {
			await auth.updateNotifyPrefs(next);
		} catch (err: unknown) {
			prefs = prev; // revert on failure
			notifyError =
				(err as { message?: string })?.message ??
				t('settings.couldNotSaveNotify');
		} finally {
			notifyBusy = false;
		}
	}

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

	<section class="card">
		<h3>{t('settings.notifications')}</h3>
		<p class="muted small">
			{@html t('settings.notifyDesc', { email: auth.user?.email ?? '' })}
		</p>

		<div class="push-device">
			{#if !push.supported}
				<p class="muted small">
					{t('settings.pushNotSupported')}
				</p>
			{:else if push.blocked}
				<p class="muted small">
					{t('settings.pushBlocked')}
				</p>
			{:else if push.subscribed}
				<div class="push-row">
					<span class="ok small">{t('settings.pushEnabledLabel')}</span>
					<div class="push-actions">
						<button
							type="button"
							class="btn secondary tiny"
							onclick={sendTest}
							disabled={testBusy}
						>
							{testBusy ? t('common.saving') : t('settings.sendTest')}
						</button>
						<button
							type="button"
							class="btn secondary tiny"
							onclick={() => push.disable()}
							disabled={push.busy}
						>
							{push.busy ? t('common.saving') : t('settings.disable')}
						</button>
					</div>
				</div>
				{#if testMsg}<p class="muted small">{testMsg}</p>{/if}
			{:else}
				<button
					type="button"
					class="btn secondary"
					onclick={() => push.enable()}
					disabled={push.busy}
				>
					{push.busy ? t('common.saving') : t('settings.enablePushDevice')}
				</button>
			{/if}
			{#if push.error}<p class="error small">{push.error}</p>{/if}
		</div>

		{#if notifyError}<p class="error">{notifyError}</p>{/if}
		<ul class="notify-list">
			<li class="notify-row notify-head">
				<span></span>
				<span class="col-label">{t('settings.emailCol')}</span>
				<span class="col-label">{t('settings.pushCol')}</span>
			</li>
			{#each NOTIFY_EVENTS as ev (ev.key)}
				<li class="notify-row">
					<div class="notify-text">
						<span class="notify-label">{t(ev.labelKey)}</span>
						<span class="muted notify-hint">{t(ev.hintKey)}</span>
					</div>
					{#each ['email', 'push'] as const as ch}
						<button
							type="button"
							role="switch"
							aria-checked={isOn(ev.key, ch)}
							aria-label={`${t(ev.labelKey)} — ${ch}`}
							class="toggle"
							class:on={isOn(ev.key, ch)}
							onclick={() => toggleNotify(ev.key, ch)}
							disabled={notifyBusy || (ch === 'push' && !push.subscribed)}
						>
							<span class="knob"></span>
						</button>
					{/each}
				</li>
			{/each}
		</ul>
		{#if push.supported && !push.subscribed}
			<p class="muted hint">{t('settings.enablePushHint')}</p>
		{/if}
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
	.notify-list {
		list-style: none;
		margin: 0.5rem 0 0;
		padding: 0;
	}
	.push-device {
		margin: 0 0 0.5rem;
	}
	.push-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
	}
	.push-actions {
		display: flex;
		gap: 0.5rem;
	}
	.btn.tiny {
		padding: 0.3rem 0.7rem;
		font-size: 0.8rem;
	}
	.notify-row {
		display: grid;
		grid-template-columns: 1fr 44px 44px;
		align-items: center;
		gap: 1rem;
		padding: 0.85rem 0;
		border-top: 1px solid var(--border);
	}
	.notify-head {
		padding: 0.2rem 0 0.4rem;
		border-top: none;
	}
	.col-label {
		text-align: center;
		font-size: 0.7rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--muted);
	}
	.notify-text {
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
	}
	.notify-label {
		font-size: 0.95rem;
		font-weight: 600;
	}
	.notify-hint {
		font-size: 0.8rem;
		line-height: 1.4;
	}
	.toggle {
		flex: none;
		width: 44px;
		height: 26px;
		border-radius: var(--radius-pill);
		border: 1px solid var(--border);
		background: var(--surface-2);
		padding: 2px;
		cursor: pointer;
		transition:
			background 0.15s ease,
			border-color 0.15s ease;
	}
	.toggle:disabled {
		opacity: 0.6;
		cursor: default;
	}
	.toggle.on {
		background: var(--accent);
		border-color: var(--accent);
	}
	.knob {
		display: block;
		width: 20px;
		height: 20px;
		border-radius: 50%;
		background: var(--text);
		transition: transform 0.15s ease;
	}
	.toggle.on .knob {
		transform: translateX(18px);
		background: var(--accent-fg);
	}
</style>
