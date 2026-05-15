import { pb } from './pb';

// Reactive auth state backed by PocketBase's authStore. Svelte 5 runes class;
// a single shared instance is exported below.
class Auth {
	user = $state<{ id: string; name: string; email: string } | null>(null);

	constructor() {
		this.sync();
		pb.authStore.onChange(() => this.sync());
	}

	private sync() {
		const r = pb.authStore.record;
		this.user =
			pb.authStore.isValid && r
				? { id: r.id, name: r.name ?? r.email, email: r.email }
				: null;
	}

	get isAuthed() {
		return this.user !== null;
	}

	async login(identity: string, password: string) {
		await pb.collection('users').authWithPassword(identity, password);
	}

	async register(name: string, email: string, password: string) {
		await pb.collection('users').create({
			name,
			email,
			password,
			passwordConfirm: password
		});
		await this.login(email, password);
	}

	logout() {
		pb.authStore.clear();
	}
}

export const auth = new Auth();
