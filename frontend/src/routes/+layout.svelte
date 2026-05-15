<script lang="ts">
	import '../app.css';
	import { auth } from '$lib/auth.svelte';
	import Nav from '$lib/components/Nav.svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';

	let { children } = $props();

	const publicRoutes = ['/login', '/register'];
	let path = $derived($page.url.pathname);
	let isPublic = $derived(publicRoutes.includes(path));

	// SPA auth guard: bounce unauthenticated users to /login, and authed
	// users away from the auth pages.
	$effect(() => {
		if (!auth.isAuthed && !isPublic) goto('/login', { replaceState: true });
		if (auth.isAuthed && isPublic) goto('/', { replaceState: true });
	});
</script>

{#if auth.isAuthed && !isPublic}
	<Nav />
{/if}

<div class="app-shell">
	{@render children()}
</div>
