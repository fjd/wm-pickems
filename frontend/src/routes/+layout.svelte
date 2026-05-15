<script lang="ts">
	import '../app.css';
	import { auth } from '$lib/auth.svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import Logo from '$lib/components/Logo.svelte';
	import UserMenu from '$lib/components/UserMenu.svelte';
	import NavLinks from '$lib/components/NavLinks.svelte';
	import { serverClock } from '$lib/serverclock.svelte';

	let { children } = $props();

	// Pull the (possibly simulated) server clock once so lock checks and the
	// dev-tools link are correct app-wide.
	$effect(() => {
		if (auth.isAuthed && !serverClock.loaded) serverClock.refresh();
	});

	const publicRoutes = ['/login', '/register'];
	let path = $derived($page.url.pathname);
	let isPublic = $derived(publicRoutes.includes(path));
	let chrome = $derived(auth.isAuthed && !isPublic);

	// SPA auth guard.
	$effect(() => {
		if (!auth.isAuthed && !isPublic) goto('/login', { replaceState: true });
		if (auth.isAuthed && isPublic) goto('/', { replaceState: true });
	});
</script>

{#if chrome}
	<!-- Mobile: top header (logo / user menu) -->
	<header class="topbar">
		<Logo />
		<div class="spacer"></div>
		<UserMenu align="right" />
	</header>

	<!-- Desktop: left rail (logo top, links, user menu bottom) -->
	<aside class="siderail">
		<div class="rail-logo"><Logo /></div>
		<NavLinks variant="rail" />
		<div class="spacer"></div>
		<div class="rail-user"><UserMenu align="left" up showName /></div>
	</aside>

	<!-- Mobile: bottom tab bar -->
	<nav class="tabbar"><NavLinks variant="tab" /></nav>
{/if}

<div class="app-shell" class:with-chrome={chrome}>
	{@render children()}
</div>
