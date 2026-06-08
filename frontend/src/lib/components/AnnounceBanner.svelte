<script lang="ts">
	// In-app announcement banner: shows active announcements written by an
	// owner/admin in the Admin console. Each is dismissed per-id in localStorage
	// so it stays gone once read, but a new announcement (new id) always shows.
	// Multiple active announcements stack, newest first.
	import { onMount } from 'svelte';
	import { api, type Announcement } from '$lib/api';
	import { Megaphone, Info, TriangleAlert, X } from '@lucide/svelte';

	const KEY = 'announce-dismissed-v1';

	let items = $state<Announcement[]>([]);
	let dismissed = $state<Set<string>>(new Set());

	function loadDismissed(): Set<string> {
		try {
			const raw = localStorage.getItem(KEY);
			if (raw) return new Set(JSON.parse(raw) as string[]);
		} catch {
			/* private mode / bad json — treat as none dismissed */
		}
		return new Set();
	}

	function persist() {
		try {
			localStorage.setItem(KEY, JSON.stringify([...dismissed]));
		} catch {
			/* ignore */
		}
	}

	function dismiss(id: string) {
		dismissed = new Set(dismissed).add(id);
		persist();
	}

	onMount(async () => {
		dismissed = loadDismissed();
		try {
			const res = await api.activeAnnouncements();
			items = res.announcements;
			// Drop dismissals for announcements that are gone/inactive so the set
			// doesn't grow without bound.
			const live = new Set(items.map((a) => a.id));
			if ([...dismissed].some((id) => !live.has(id))) {
				dismissed = new Set([...dismissed].filter((id) => live.has(id)));
				persist();
			}
		} catch {
			/* not signed in / offline — show nothing */
		}
	});

	let visible = $derived(items.filter((a) => !dismissed.has(a.id)));

	const icon = { info: Info, success: Megaphone, warn: TriangleAlert };
</script>

{#if visible.length}
	<div class="stack">
		{#each visible as a (a.id)}
			{@const Icon = icon[a.level] ?? Megaphone}
			<div class="banner {a.level}" role="status">
				<span class="ico"><Icon size={18} /></span>
				<div class="text">
					<strong class="t">{a.title}</strong>
					<span class="b">{a.body}</span>
				</div>
				<button class="x" aria-label="Dismiss" onclick={() => dismiss(a.id)}>
					<X size={16} />
				</button>
			</div>
		{/each}
	</div>
{/if}

<style>
	.stack {
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
		margin-bottom: 1rem;
	}
	.banner {
		display: grid;
		grid-template-columns: auto 1fr auto;
		align-items: start;
		gap: 0.7rem;
		padding: 0.8rem 0.9rem;
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
	}

	/* Info — calm and neutral: flat surface, thin grey rule, outlined icon. */
	.banner.info {
		border-left: 3px solid color-mix(in srgb, var(--muted) 45%, var(--border));
	}
	.banner.info .ico {
		background: var(--surface-2);
		color: var(--muted);
	}

	/* Highlight — loud and on-brand: lime wash, filled icon chip, soft glow,
	   bright title. The visual opposite of the muted info notice. */
	.banner.success {
		background: color-mix(in srgb, var(--accent) 13%, var(--surface));
		border-color: color-mix(in srgb, var(--accent) 45%, var(--border));
		border-left: 3px solid var(--accent);
		box-shadow: 0 12px 32px -18px color-mix(in srgb, var(--accent) 70%, transparent);
	}
	.banner.success .ico {
		background: var(--accent);
		color: var(--accent-fg);
	}
	.banner.success .t {
		color: var(--accent-2);
	}

	/* Warning — amber wash, filled icon. */
	.banner.warn {
		background: color-mix(in srgb, var(--warning) 12%, var(--surface));
		border-color: color-mix(in srgb, var(--warning) 45%, var(--border));
		border-left: 3px solid var(--warning);
	}
	.banner.warn .ico {
		background: var(--warning);
		color: #20160a;
	}

	.ico {
		display: inline-grid;
		place-items: center;
		width: 28px;
		height: 28px;
		border-radius: var(--radius-pill);
		margin-top: 0.05rem;
	}
	.text {
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
		min-width: 0;
		font-size: 0.9rem;
		line-height: 1.5;
	}
	.t {
		font-weight: 700;
		color: var(--text);
	}
	.b {
		color: var(--muted);
		white-space: pre-line;
	}
	.x {
		display: inline-grid;
		place-items: center;
		width: 30px;
		height: 30px;
		flex-shrink: 0;
		border: none;
		background: transparent;
		color: var(--muted);
		border-radius: var(--radius-pill);
		cursor: pointer;
	}
	.x:hover {
		color: var(--text);
		background: var(--surface-2);
	}
</style>
