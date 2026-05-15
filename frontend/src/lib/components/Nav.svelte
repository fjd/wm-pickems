<script lang="ts">
	import { page } from '$app/stores';

	const items = [
		{ href: '/', label: 'Home', ico: '⚽' },
		{ href: '/tips', label: 'Tips', ico: '🎯' },
		{ href: '/forecast', label: 'Forecast', ico: '🔮' },
		{ href: '/leagues', label: 'Leagues', ico: '🏆' }
	];

	function isActive(href: string, path: string) {
		return href === '/' ? path === '/' : path.startsWith(href);
	}
</script>

<nav class="nav">
	{#each items as it (it.href)}
		<a href={it.href} class:active={isActive(it.href, $page.url.pathname)}>
			<span class="ico">{it.ico}</span>
			<span>{it.label}</span>
		</a>
	{/each}
</nav>
