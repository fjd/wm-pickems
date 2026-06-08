<script lang="ts">
	// In-app announcement banners written by an admin in the Admin console.
	// Two kinds:
	//  - dismissible (default): an X removes it for good (per-id in localStorage);
	//    a new announcement (new id) always shows.
	//  - persistent: can't be dismissed — only collapsed to a slim ribbon and
	//    re-expanded. Collapsed state is remembered per-id. It stays until an
	//    admin deactivates it.
	// Multiple active announcements stack, newest first.
	import { onMount } from 'svelte';
	import { api, type Announcement } from '$lib/api';
	import {
		Megaphone,
		Info,
		TriangleAlert,
		X,
		ChevronUp,
		ChevronDown
	} from '@lucide/svelte';

	const DISMISS_KEY = 'announce-dismissed-v1';
	const COLLAPSE_KEY = 'announce-collapsed-v1';

	let items = $state<Announcement[]>([]);
	let dismissed = $state<Set<string>>(new Set());
	let collapsed = $state<Set<string>>(new Set());

	function loadSet(key: string): Set<string> {
		try {
			const raw = localStorage.getItem(key);
			if (raw) return new Set(JSON.parse(raw) as string[]);
		} catch {
			/* private mode / bad json — treat as empty */
		}
		return new Set();
	}

	function persist(key: string, set: Set<string>) {
		try {
			localStorage.setItem(key, JSON.stringify([...set]));
		} catch {
			/* ignore */
		}
	}

	function dismiss(id: string) {
		dismissed = new Set(dismissed).add(id);
		persist(DISMISS_KEY, dismissed);
	}

	function toggleCollapse(id: string) {
		const next = new Set(collapsed);
		if (next.has(id)) next.delete(id);
		else next.add(id);
		collapsed = next;
		persist(COLLAPSE_KEY, collapsed);
	}

	// Drop remembered ids that are no longer live so the sets don't grow forever.
	function prune(key: string, set: Set<string>, live: Set<string>): Set<string> {
		if (![...set].some((id) => !live.has(id))) return set;
		const next = new Set([...set].filter((id) => live.has(id)));
		persist(key, next);
		return next;
	}

	onMount(async () => {
		dismissed = loadSet(DISMISS_KEY);
		collapsed = loadSet(COLLAPSE_KEY);
		try {
			const res = await api.activeAnnouncements();
			items = res.announcements;
			const liveIds = new Set(items.map((a) => a.id));
			dismissed = prune(DISMISS_KEY, dismissed, liveIds);
			collapsed = prune(COLLAPSE_KEY, collapsed, liveIds);
		} catch {
			/* not signed in / offline — show nothing */
		}
	});

	// Persistent items can't be dismissed; dismissible ones drop out once dismissed.
	let visible = $derived(items.filter((a) => a.persistent || !dismissed.has(a.id)));

	const icon = { info: Info, success: Megaphone, warn: TriangleAlert };
</script>

{#if visible.length}
	<div class="stack">
		{#each visible as a (a.id)}
			{@const Icon = icon[a.level] ?? Megaphone}
			{#if a.persistent && collapsed.has(a.id)}
				<!-- collapsed: a slim, tappable ribbon -->
				<button
					class="ribbon {a.level}"
					onclick={() => toggleCollapse(a.id)}
					aria-label="Expand announcement"
				>
					<Icon size={15} />
					<span class="rtitle">{a.title}</span>
					<ChevronDown size={16} class="rchev" />
				</button>
			{:else if a.persistent}
				<!-- persistent + expanded: the whole card collapses on click -->
				<button
					class="banner tap {a.level}"
					title="Collapse"
					onclick={() => toggleCollapse(a.id)}
				>
					<span class="ico"><Icon size={18} /></span>
					<span class="text">
						<strong class="t">{a.title}</strong>
						<span class="b">{a.body}</span>
					</span>
					<span class="x" aria-hidden="true"><ChevronUp size={16} /></span>
				</button>
			{:else}
				<!-- dismissible: static card, only the X removes it -->
				<div class="banner {a.level}" role="status">
					<span class="ico"><Icon size={18} /></span>
					<span class="text">
						<strong class="t">{a.title}</strong>
						<span class="b">{a.body}</span>
					</span>
					<button class="x" aria-label="Dismiss" onclick={() => dismiss(a.id)}>
						<X size={16} />
					</button>
				</div>
			{/if}
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
	/* Persistent expanded card is a button — the whole thing collapses on click. */
	.banner.tap {
		width: 100%;
		font: inherit;
		text-align: left;
		cursor: pointer;
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

	/* ---- collapsed persistent announcement: a slim ribbon (echoes the landing
	   lock-countdown bar for the highlight level) ---- */
	.ribbon {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		width: 100%;
		padding: 0.5rem 0.8rem;
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		background: var(--surface-2);
		color: var(--text);
		font: inherit;
		text-align: left;
		cursor: pointer;
	}
	.ribbon:hover {
		filter: brightness(1.06);
	}
	.rtitle {
		flex: 1;
		min-width: 0;
		font-weight: 700;
		font-size: 0.9rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	:global(.ribbon .rchev) {
		flex-shrink: 0;
		opacity: 0.7;
	}
	.ribbon.info {
		border-left: 3px solid color-mix(in srgb, var(--muted) 45%, var(--border));
	}
	.ribbon.success {
		background: linear-gradient(90deg, var(--accent), var(--accent-2));
		color: var(--accent-ink);
		border-color: transparent;
		font-weight: 700;
	}
	.ribbon.warn {
		background: var(--warning);
		color: #20160a;
		border-color: transparent;
	}

	/* On mobile the collapsed ribbon spans edge-to-edge — break out of the
	   app-shell's 1rem gutter (which is safe: the shell is overflow-x: clip).
	   Inner padding stays 1rem so the text still lines up with page content. */
	@media (max-width: 899px) {
		.ribbon {
			margin-inline: -1rem;
			padding-inline: 1rem;
			border-radius: 0;
			border-left: none;
			border-right: none;
		}
	}
</style>
