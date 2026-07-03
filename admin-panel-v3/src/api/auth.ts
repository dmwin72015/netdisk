import { request } from '@/utils/request';
import type { LoginResponse } from './types';
export type { LoginResponse };

export async function login(email: string, password: string): Promise<LoginResponse> {
  return request.post<LoginResponse>('/api/v1/auth/login', { email, password });
}
