<script lang="ts">
	// Standalone, chrome-less view of the marketing landing page, reachable from
	// the in-app help (?) link. Anonymous visitors also get the landing at `/`;
	// this route lets signed-in users revisit it (with a "Back to the app" CTA).
	import Landing from '$lib/components/Landing.svelte';
</script>

<Landing />
