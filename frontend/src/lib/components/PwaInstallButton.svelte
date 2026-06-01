<script lang="ts">
	import { pwa } from '$lib/pwa.svelte';
	import { t } from '$lib/i18n.svelte';
</script>

{#if pwa.available}
	<button
		type="button"
		class="btn install"
		class:outlined={!pwa.dismissed}
		aria-label={t('pwa.installApp')}
		onclick={() => pwa.install()}
	>
		{t('common.install')}
	</button>
{/if}

<style>
	.install {
		width: auto;
		padding: 0.45rem 0.85rem;
		font-size: 0.78rem;
		flex: none;
	}
	.install.outlined {
		background-color: transparent;
		border-color: var(--accent);
		color: var(--accent);
	}
	@media (min-width: 900px) {
		.install {
			display: none;
		}
	}
</style>
