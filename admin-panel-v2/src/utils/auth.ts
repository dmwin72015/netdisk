import { getToken, removeToken, setToken } from './token';
import { get, post } from './request';

interface User {
  id: string;
  email: string;
  name: string;
  role: string;
}

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  login: (email: string, password: string, autoLogin?: boolean) => Promise<{ access_token: string; user: User }>;
  logout: () => Promise<void>;
  fetchUser: () => Promise<User | null>;
}

let userState: User | null = null;

export function useAuthStore(): AuthState {
  return {
    get user() {
      return userState;
    },
    get isAuthenticated() {
      return !!getToken();
    },

    async login(email: string, password: string, autoLogin = false) {
      const data = (await post<{ access_token: string; user: User }>('/auth/login', {
        email,
        password,
      })) as { access_token: string; user: User };
      setToken(data.access_token, autoLogin);
      userState = data.user;
      return data;
    },

    async logout() {
      try {
        await post('/auth/logout');
      } finally {
        removeToken();
        userState = null;
      }
    },

    async fetchUser() {
      const data = (await get<User>('/user/me')) as User;
      userState = data;
      return data;
    },
  };
}
