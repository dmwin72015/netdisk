import { request } from '../request';
import type {
  AdminUser,
  AdminUserList,
  CreateUserInput,
  UserListParams,
  UpdateRoleInput,
  UpdateStorageBaseInput,
} from './types';

export async function fetchUsers(params?: UserListParams): Promise<AdminUserList> {
  return request.get<AdminUserList>('/api/v1/admin/users', { params });
}

export async function fetchUser(id: string): Promise<AdminUser> {
  return request.get<AdminUser>(`/api/v1/admin/users/${id}`);
}

export async function createUser(data: CreateUserInput): Promise<AdminUser> {
  return request.post<AdminUser>('/api/v1/admin/users', data);
}

export async function updateUserRole(id: string, role: string): Promise<AdminUser> {
  return request.patch<AdminUser>(`/api/v1/admin/users/${id}`, { role } satisfies UpdateRoleInput);
}

export async function updateStorageBase(id: string, baseBytes: number): Promise<AdminUser> {
  return request.patch<AdminUser>(
    `/api/v1/admin/users/${id}/storage-base`,
    { baseBytes } satisfies UpdateStorageBaseInput,
  );
}

export async function deleteUser(id: string): Promise<void> {
  await request.delete<void>(`/api/v1/admin/users/${id}`);
}

export async function searchUsers(query: string): Promise<{ id: string; username: string }[]> {
  return request.get<{ id: string; username: string }[]>('/api/v1/admin/users/search', {
    params: { q: query },
  });
}
