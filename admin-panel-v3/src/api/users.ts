import { message } from 'antd';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { request } from '@/utils/request';
import type {
  AdminUser,
  AdminUserList,
  CreateUserInput,
  DeleteActionResult,
  UpdateRoleInput,
  UpdateStorageBaseInput,
  UserListParams,
} from './types';
import type { ListQueryResult } from '@/types/query';

export type {
  AdminUser,
  AdminUserList,
  CreateUserInput,
  DeleteActionResult,
  UserListParams,
};

// --- Raw API ---

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

export async function deleteUser(id: string): Promise<DeleteActionResult> {
  return request.delete<DeleteActionResult>(`/api/v1/admin/users/${id}`);
}

export async function searchUsers(query: string): Promise<AdminUser[]> {
  return request.get<AdminUser[]>('/api/v1/admin/users/search', {
    params: { q: query },
  });
}

// --- Hooks ---

export function useUsers(params: UserListParams): ListQueryResult<AdminUser> {
  const query = useQuery({
    queryKey: ['admin', 'users', params],
    queryFn: () => fetchUsers(params),
  });
  return {
    data: query.data?.items ?? [],
    total: query.data?.total ?? 0,
    isLoading: query.isLoading,
    isFetching: query.isFetching,
    isError: query.isError,
    error: query.error as Error | null,
    refetch: () => query.refetch(),
  };
}

export function useUser(id: string) {
  return useQuery({
    queryKey: ['admin', 'users', id],
    queryFn: () => fetchUser(id),
    enabled: !!id,
  });
}

export function useCreateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateUserInput) => createUser(data),
    onSuccess: () => {
      message.success('创建成功');
      qc.invalidateQueries({ queryKey: ['admin', 'users'] });
    },
  });
}

export function useUpdateUserRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, role }: { id: string; role: string }) => updateUserRole(id, role),
    onSuccess: () => {
      message.success('角色已更新');
      qc.invalidateQueries({ queryKey: ['admin', 'users'] });
    },
  });
}

export function useUpdateStorageBase() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, baseBytes }: { id: string; baseBytes: number }) => updateStorageBase(id, baseBytes),
    onSuccess: () => {
      message.success('存储配额已更新');
      qc.invalidateQueries({ queryKey: ['admin', 'users'] });
    },
  });
}

export function useDeleteUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteUser(id),
    onSuccess: () => {
      message.success('用户已删除');
      qc.invalidateQueries({ queryKey: ['admin', 'users'] });
    },
  });
}

export function useSearchUsers(query: string) {
  return useQuery({
    queryKey: ['admin', 'users', 'search', query],
    queryFn: () => searchUsers(query),
    enabled: query.length > 0,
  });
}