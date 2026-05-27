import { api, setSession, type Tokens, type UserInfo } from './client';

export type LoginResponse = {
	user: UserInfo;
	tokens: Tokens;
};

export async function login(email: string, password: string): Promise<LoginResponse> {
	const data = await api<LoginResponse>('/api/v1/auth/login', {
		method: 'POST',
		body: JSON.stringify({ email, password }),
		auth: false
	});
	setSession(data.user, data.tokens);
	return data;
}

export async function register(username: string, email: string, password: string): Promise<UserInfo> {
	return api<UserInfo>('/api/v1/auth/register', {
		method: 'POST',
		body: JSON.stringify({ username, email, password }),
		auth: false
	});
}

export async function logout(): Promise<void> {
	try {
		const refresh = localStorage.getItem('nd.refresh');
		if (refresh) {
			await api('/api/v1/auth/logout', {
				method: 'POST',
				body: JSON.stringify({ refresh_token: refresh })
			});
		}
	} catch {
		// best-effort
	} finally {
		setSession(null, null);
	}
}
