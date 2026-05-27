import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import { getStoredUser, type UserInfo } from '$lib/api/client';

const initial: UserInfo | null = browser ? getStoredUser() : null;

export const user = writable<UserInfo | null>(initial);

/** False during SSR; set true in layout onMount after reading localStorage. */
export const authReady = writable(false);

export function setUser(u: UserInfo | null) {
	user.set(u);
	if (browser) {
		localStorage.setItem('nd.user', JSON.stringify(u));
		authReady.set(true);
	}
}
