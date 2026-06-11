// Reactive Web Push state for the current device. Mirrors pwa.svelte.ts: a
// single singleton (`push`) exposes whether push is supported, the current
// permission, whether this device is subscribed, and enable()/disable()
// actions that manage the PushManager subscription and sync it to the backend.
import { pb } from './pb';

// localStorage key: the account id this device's push subscription is currently
// registered to on the server. Lets syncAccount() skip the common no-change case
// and only re-bind when the signed-in account actually differs.
const BOUND_KEY = 'push-bound-user';

// VAPID public keys are base64url; PushManager wants a Uint8Array.
function urlBase64ToUint8Array(base64: string): Uint8Array<ArrayBuffer> {
	const padding = '='.repeat((4 - (base64.length % 4)) % 4);
	const b64 = (base64 + padding).replace(/-/g, '+').replace(/_/g, '/');
	const raw = atob(b64);
	const buf = new ArrayBuffer(raw.length);
	const out = new Uint8Array(buf);
	for (let i = 0; i < raw.length; i++) out[i] = raw.charCodeAt(i);
	return out;
}

class Push {
	supported = $state(false);
	permission = $state<NotificationPermission>('default');
	subscribed = $state(false);
	ready = $state(false); // initial subscription check has completed
	busy = $state(false);
	error = $state('');

	constructor() {
		if (typeof window === 'undefined') return;
		this.supported =
			'serviceWorker' in navigator &&
			'PushManager' in window &&
			'Notification' in window;
		if (!this.supported) {
			this.ready = true;
			return;
		}
		this.permission = Notification.permission;
		void this.refresh();
		// A browser push subscription is per-device, not per-account. On a shared
		// device, signing in as a different account would otherwise leave the
		// server-side subscription pointing at the previous account (so pushes go
		// to the wrong person and the new account silently gets none). Re-bind the
		// device's subscription to whoever is signed in — on initial load and on
		// every login/logout. syncAccount() is a no-op when nothing changed.
		pb.authStore.onChange(() => void this.syncAccount(), true);
	}

	// Current signed-in user id, or null when signed out.
	private get uid(): string | null {
		return pb.authStore.isValid ? (pb.authStore.record?.id ?? null) : null;
	}

	// Re-point this device's existing push subscription at the current account
	// when it differs from the one it's bound to. The server's /subscribe upserts
	// by endpoint and reassigns the owner, so this moves the row to the new user.
	private async syncAccount() {
		if (!this.supported) return;
		const uid = this.uid;
		if (!uid) return; // signed out — leave the sub until someone signs in
		let bound: string | null = null;
		try {
			bound = localStorage.getItem(BOUND_KEY);
		} catch {
			/* private mode */
		}
		if (bound === uid) return; // already bound to this account
		try {
			const reg = await navigator.serviceWorker.ready;
			const sub = await reg.pushManager.getSubscription();
			if (!sub) return; // device isn't subscribed — nothing to re-bind
			await pb.send('/api/push/subscribe', { method: 'POST', body: sub.toJSON() });
			this.subscribed = true;
			this.setBound(uid);
		} catch {
			/* best effort — a transient failure self-heals on the next auth change */
		}
	}

	private setBound(uid: string) {
		try {
			localStorage.setItem(BOUND_KEY, uid);
		} catch {
			/* ignore */
		}
	}

	private clearBound() {
		try {
			localStorage.removeItem(BOUND_KEY);
		} catch {
			/* ignore */
		}
	}

	// Sync `subscribed` with the actual PushManager state.
	private async refresh() {
		try {
			const reg = await navigator.serviceWorker.ready;
			const sub = await reg.pushManager.getSubscription();
			this.subscribed = !!sub;
		} catch {
			this.subscribed = false;
		} finally {
			this.ready = true;
		}
	}

	// True when the browser blocked notifications (user must re-allow in
	// browser settings — we can't re-prompt).
	get blocked() {
		return this.permission === 'denied';
	}

	async enable() {
		if (!this.supported || this.busy) return;
		this.error = '';
		this.busy = true;
		try {
			this.permission = await Notification.requestPermission();
			if (this.permission !== 'granted') {
				this.error =
					this.permission === 'denied'
						? 'Notifications are blocked in your browser settings.'
						: 'Permission was not granted.';
				return;
			}
			const { publicKey } = await pb.send<{ publicKey: string }>(
				'/api/push/key',
				{ method: 'GET' }
			);
			if (!publicKey) {
				this.error = 'Push is not configured on the server.';
				return;
			}
			const reg = await navigator.serviceWorker.ready;
			const sub = await reg.pushManager.subscribe({
				userVisibleOnly: true,
				applicationServerKey: urlBase64ToUint8Array(publicKey)
			});
			const json = sub.toJSON() as {
				endpoint: string;
				keys: { p256dh: string; auth: string };
			};
			await pb.send('/api/push/subscribe', { method: 'POST', body: json });
			this.subscribed = true;
			if (this.uid) this.setBound(this.uid);
			// Fire an instant confirmation so the user sees push actually works.
			this.test().catch(() => {});
		} catch (err: unknown) {
			this.error =
				(err as { message?: string })?.message ??
				'Could not enable push notifications.';
		} finally {
			this.busy = false;
		}
	}

	// Ask the server to push a test notification to this account's devices.
	// Returns how many endpoints accepted it (for surfacing in the UI).
	async test(): Promise<{ sent: number; total: number }> {
		this.error = '';
		const res = await pb.send<{ sent: number; total: number }>(
			'/api/push/test',
			{ method: 'POST' }
		);
		return res;
	}

	async disable() {
		if (!this.supported || this.busy) return;
		this.error = '';
		this.busy = true;
		try {
			const reg = await navigator.serviceWorker.ready;
			const sub = await reg.pushManager.getSubscription();
			if (sub) {
				const endpoint = sub.endpoint;
				await sub.unsubscribe();
				await pb
					.send('/api/push/unsubscribe', { method: 'POST', body: { endpoint } })
					.catch(() => {});
			}
			this.subscribed = false;
			this.clearBound();
		} catch (err: unknown) {
			this.error =
				(err as { message?: string })?.message ??
				'Could not disable push notifications.';
		} finally {
			this.busy = false;
		}
	}
}

export const push = new Push();
