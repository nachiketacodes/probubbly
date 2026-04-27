const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export async function apiRequest(endpoint: string, options: RequestInit = {}): Promise<Response> {
	const token = localStorage.getItem('token');

	const headers: HeadersInit = {
		'Content-Type': 'application/json',
		...options.headers
	};

	if (token) {
		headers['Authorization'] = `Bearer ${token}`;
	}

	const response = await fetch(`${API_URL}${endpoint}`, {
		...options,
		headers,
		credentials: 'include'
	});

	return response;
}

export const api = {
	// Auth
	signup: (loginId: string, pin: string, username: string) =>
		apiRequest('/api/auth/signup', {
			method: 'POST',
			body: JSON.stringify({ login_id: loginId, pin, username })
		}),

	login: (loginId: string, pin: string) =>
		apiRequest('/api/auth/login', {
			method: 'POST',
			body: JSON.stringify({ login_id: loginId, pin })
		}),

	// Events
	getEvents: (status?: string) => {
		const query = status ? `?status=${status}` : '';
		return apiRequest(`/api/events${query}`);
	},

	getEvent: (id: string) => apiRequest(`/api/events/${id}`),

	createEvent: (data: {
		title: string;
		description: string;
		event_date: string;
		event_time?: string;
	}) =>
		apiRequest('/api/events', {
			method: 'POST',
			body: JSON.stringify(data)
		}),

	predict: (eventId: string, side: string, amount: number) =>
		apiRequest(`/api/events/${eventId}/predict`, {
			method: 'POST',
			body: JSON.stringify({ side, amount })
		}),

	resolveEvent: (eventId: string, outcome: string) =>
		apiRequest(`/api/events/${eventId}/resolve`, {
			method: 'POST',
			body: JSON.stringify({ outcome })
		}),

	// Wallet
	getWallet: () => apiRequest('/api/wallet'),

	borrowCoins: () =>
		apiRequest('/api/wallet/borrow', {
			method: 'POST'
		}),

	// Admin
	getAdminStats: () => apiRequest('/api/admin/stats'),
	getAllUsers: () => apiRequest('/api/admin/users'),
	getUserDetail: (id: string) => apiRequest(`/api/admin/user?id=${id}`),
	getHouseLedger: () => apiRequest('/api/admin/house-ledger')
};
