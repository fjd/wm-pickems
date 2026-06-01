<script lang="ts">
	import { pwa } from '$lib/pwa.svelte';
	import { t } from '$lib/i18n.svelte';
	import { Download, X, Share } from '@lucide/svelte';
</script>

{#if pwa.bannerOpen}
	<div class="banner" role="region" aria-label={t('pwa.installApp')}>
		<div class="inner">
			<Download size={18} class="ico" />
			<div class="msg">
				<strong>{t('pwa.installApp')}</strong>
				<span class="muted small">{t('pwa.installDesc')}</span>
			</div>
			<button class="btn install" onclick={() => pwa.install()}>{t('common.install')}</button>
			<button
				class="x"
				aria-label={t('common.dismiss')}
				onclick={() => pwa.dismissBanner()}
			>
				<X size={16} />
			</button>
		</div>
	</div>
{/if}

{#if pwa.iosHelpOpen}
	<button
		type="button"
		class="ios-backdrop"
		aria-label={t('common.close')}
		onclick={() => pwa.closeIosHelp()}
	></button>
	<div class="ios-sheet" role="dialog" aria-label={t('pwa.installInstructions')}>
		<h3>{t('pwa.addToHomeScreen')}</h3>
		<ol>
			<li>
				{@html t('pwa.iosStep1', { iconSize: '14' })}
			</li>
			<li>{@html t('pwa.iosStep2')}</li>
			<li>{@html t('pwa.iosStep3')}</li>
		</ol>
		<button class="btn" onclick={() => pwa.closeIosHelp()}>{t('common.gotIt')}</button>
	</div>
{/if}

<style>
	.banner {
		margin: 0 0 1rem;
		padding: 0.7rem 0.85rem;
		background: color-mix(in srgb, var(--accent) 18%, var(--surface));
		border: 1px solid color-mix(in srgb, var(--accent) 35%, var(--border));
		border-radius: var(--radius);
	}
	.inner {
		display: flex;
		align-items: center;
		gap: 0.65rem;
	}
	:global(.banner .ico) {
		color: var(--accent);
		flex: none;
	}
	.msg {
		display: flex;
		flex-direction: column;
		min-width: 0;
		flex: 1;
		line-height: 1.25;
	}
	.msg strong {
		font-size: 0.95rem;
	}
	.msg .small {
		font-size: 0.78rem;
	}
	.btn.install {
		padding: 0.45rem 0.85rem;
		font-size: 0.85rem;
		width: auto;
	}
	.x {
		display: inline-grid;
		place-items: center;
		width: 32px;
		height: 32px;
		border-radius: 999px;
		background: transparent;
		color: var(--muted);
		border: 1px solid transparent;
		cursor: pointer;
	}
	.x:hover {
		color: var(--text);
		background: var(--surface-2);
	}

	.ios-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.45);
		border: none;
		padding: 0;
		z-index: 60;
		cursor: pointer;
	}
	.ios-sheet {
		position: fixed;
		inset: auto 0.75rem calc(var(--nav-h, 0px) + 0.75rem) 0.75rem;
		z-index: 61;
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 1rem 1.1rem 1.1rem;
		box-shadow: var(--shadow-pop);
		max-width: 420px;
		margin: 0 auto;
	}
	.ios-sheet h3 {
		margin: 0 0 0.65rem;
		font-size: 1.05rem;
	}
	.ios-sheet ol {
		margin: 0 0 1rem;
		padding-left: 1.25rem;
		line-height: 1.55;
		font-size: 0.92rem;
	}
	.ios-sheet ol li + li {
		margin-top: 0.4rem;
	}
	.kbd {
		display: inline-flex;
		align-items: center;
		gap: 0.3rem;
		padding: 0.1rem 0.4rem;
		border: 1px solid var(--border);
		border-radius: 4px;
		background: var(--surface-2);
		font-size: 0.85em;
	}

	@media (min-width: 900px) {
		.banner,
		.ios-backdrop,
		.ios-sheet {
			display: none;
		}
	}
</style>
