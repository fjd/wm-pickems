<script lang="ts">
	// An animated GIF in chat that pauses on click. Pausing snapshots the current
	// frame onto a <canvas> (drawImage works on a cross-origin image; we only
	// display the canvas, never read it back, so no CORS is required) and hides
	// the live <img>. Clicking again plays it.
	import { Play } from '@lucide/svelte';

	let { src, alt = 'GIF' }: { src: string; alt?: string } = $props();

	let img = $state<HTMLImageElement | null>(null);
	let canvas = $state<HTMLCanvasElement | null>(null);
	let paused = $state(false);

	function toggle() {
		if (paused) {
			paused = false;
			return;
		}
		const c = canvas;
		const i = img;
		if (c && i && i.naturalWidth) {
			c.width = i.naturalWidth;
			c.height = i.naturalHeight;
			c.getContext('2d')?.drawImage(i, 0, 0);
			paused = true;
		}
	}
</script>

<button
	class="cg"
	onclick={toggle}
	aria-label={paused ? 'Play GIF' : 'Pause GIF'}
	title={paused ? 'Play' : 'Pause'}
>
	<img bind:this={img} {src} {alt} class:hidden={paused} loading="lazy" />
	<canvas bind:this={canvas} class:hidden={!paused}></canvas>
	{#if paused}
		<span class="play" aria-hidden="true"><Play size={20} /></span>
	{/if}
</button>

<style>
	.cg {
		position: relative;
		display: block;
		padding: 0;
		border: none;
		background: none;
		line-height: 0;
		cursor: pointer;
	}
	.cg img,
	.cg canvas {
		display: block;
		max-width: min(220px, 60vw);
		height: auto;
		border-radius: 11px;
	}
	/* Scoped above `.cg img/canvas` so it actually wins the cascade. */
	.cg .hidden {
		display: none;
	}
	.play {
		position: absolute;
		inset: 0;
		margin: auto;
		width: 40px;
		height: 40px;
		display: grid;
		place-items: center;
		border-radius: 999px;
		background: rgba(0, 0, 0, 0.5);
		color: #fff;
		pointer-events: none;
	}
</style>
