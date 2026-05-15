import { pb } from './pb';

// The app's clock comes from the server (/api/now) so client-side lock checks
// honor the dev virtual clock. now() = Date.now() + measured offset.
class ServerClock {
	offset = $state(0);
	dev = $state(false);
	simulated = $state(false);
	simTime = $state<string | null>(null);
	loaded = $state(false);

	async refresh() {
		try {
			const r = await pb.send('/api/now', { method: 'GET' });
			this.offset = r.now - Date.now();
			this.dev = !!r.dev;
			this.simulated = !!r.simulated;
			this.simTime = r.simTime ?? null;
		} catch {
			/* fall back to local time */
		} finally {
			this.loaded = true;
		}
	}

	now(): number {
		return Date.now() + this.offset;
	}
}

export const serverClock = new ServerClock();
