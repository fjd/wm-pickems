<script lang="ts">
	import { page } from '$app/stores';
	import { navItems, isActive } from '$lib/nav';

	let { variant = 'tab' as 'tab' | 'rail' } = $props();
	let path = $derived($page.url.pathname);
</script>

<div class="links {variant}">
	{#each navItems as it (it.href)}
		{@const Icon = it.icon}
		<a href={it.href} class:active={isActive(it.href, path)}>
			<Icon size={variant === 'rail' ? 20 : 22} />
			<span>{it.label}</span>
		</a>
	{/each}
</div>

<style>
	.links {
		display: flex;
	}
	.links a {
		display: flex;
		align-items: center;
		color: var(--muted);
	}
	.links a.active {
		color: var(--accent);
	}

	/* Mobile bottom tab bar */
	.tab {
		flex: 1;
	}
	.tab a {
		flex: 1;
		flex-direction: column;
		justify-content: center;
		gap: 3px;
		font-size: 0.68rem;
		padding: 0.4rem 0;
	}

	/* Desktop side rail */
	.rail {
		flex-direction: column;
		gap: 0.2rem;
		width: 100%;
	}
	.rail a {
		gap: 0.8rem;
		padding: 0.7rem 1.4rem;
		font-size: 0.95rem;
		font-weight: 600;
		border-radius: 10px;
		margin: 0 0.6rem;
	}
	.rail a.active {
		background: var(--surface-2);
	}
</style>
