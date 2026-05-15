<script lang="ts">
	import { Trophy } from '@lucide/svelte';
	let { compact = false }: { compact?: boolean } = $props();
</script>

<a class="logo" href="/" aria-label="WM Pickems home">
	<span class="mark"><Trophy size={20} strokeWidth={2.4} /></span>
	{#if !compact}<span class="word">WM&nbsp;Pickems</span>{/if}
</a>

<style>
	.logo {
		display: inline-flex;
		align-items: center;
		gap: 0.55rem;
		color: var(--text);
		font-weight: 800;
		letter-spacing: -0.02em;
	}
	.mark {
		display: grid;
		place-items: center;
		width: 32px;
		height: 32px;
		border-radius: 9px;
		background: linear-gradient(135deg, var(--accent), var(--accent-2));
		color: var(--accent-ink);
	}
	.word {
		font-size: 1.05rem;
	}
</style>
