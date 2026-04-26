import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export interface User {
	id: string;
	username: string;
	balance: number;
	borrowed: number;
	is_admin: boolean;
}

interface AuthState {
	user: User | null;
	token: string | null;
}

function createAuthStore() {
	// Initialize from localStorage if in browser
	const initialState: AuthState = browser
		? {
				token: localStorage.getItem('token'),
				user: localStorage.getItem('user') ? JSON.parse(localStorage.getItem('user')!) : null
			}
		: { token: null, user: null };

	const { subscribe, set, update } = writable<AuthState>(initialState);

	return {
		subscribe,
		login: (token: string, user: User) => {
			if (browser) {
				localStorage.setItem('token', token);
				localStorage.setItem('user', JSON.stringify(user));
			}
			set({ token, user });
		},
		logout: () => {
			if (browser) {
				localStorage.removeItem('token');
				localStorage.removeItem('user');
			}
			set({ token: null, user: null });
		},
		updateBalance: (newBalance: number) => {
			update((state) => {
				if (state.user) {
					const updatedUser = { ...state.user, balance: newBalance };
					if (browser) {
						localStorage.setItem('user', JSON.stringify(updatedUser));
					}
					return { ...state, user: updatedUser };
				}
				return state;
			});
		}
	};
}

export const auth = createAuthStore();
