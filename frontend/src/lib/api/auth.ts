import { api, setSession, type Tokens, type UserInfo } from './client';

export type LoginResponse = {
	user: UserInfo;
	tokens: Tokens;
};

export async function login(email: string, password: string, deviceId?: string): Promise<LoginResponse> {
	const body: Record<string, string> = { email, password };
	if (deviceId) body.deviceId = deviceId;
	const data = await api<LoginResponse>('/api/v1/auth/login', {
		method: 'POST',
		body: JSON.stringify(body),
		auth: false
	});
	setSession(data.user, data.tokens);
	return data;
}

export async function register(username: string, email: string, password: string, deviceId?: string): Promise<UserInfo> {
	const body: Record<string, string> = { username, email, password };
	if (deviceId) body.deviceId = deviceId;
	return api<UserInfo>('/api/v1/auth/register', {
		method: 'POST',
		body: JSON.stringify(body),
		auth: false
	});
}

export async function logout(): Promise<void> {
	try {
		const refresh = localStorage.getItem('nd.refresh');
		if (refresh) {
			await api('/api/v1/auth/logout', {
				method: 'POST',
				body: JSON.stringify({ refreshToken: refresh })
			});
		}
	} catch {
		// best-effort
	} finally {
		setSession(null, null);
	}
}
