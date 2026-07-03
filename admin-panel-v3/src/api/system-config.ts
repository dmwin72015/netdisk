import { message } from 'antd';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { request } from '@/utils/request';
import type { SystemConfigItem, UpdateSystemConfigInput } from './types';

export type { SystemConfigItem, UpdateSystemConfigInput };

// --- Raw API ---

export async function fetchSystemConfig(): Promise<SystemConfigItem[]> {
  return request.get<SystemConfigItem[]>('/api/v1/admin/system/config');
}

export async function updateSystemConfig(updates: UpdateSystemConfigInput): Promise<SystemConfigItem[]> {
  return request.put<SystemConfigItem[]>('/api/v1/admin/system/config', updates);
}

export async function resetSystemConfig(key?: string): Promise<SystemConfigItem[]> {
  return request.post<SystemConfigItem[]>('/api/v1/admin/system/config/reset', key ? { key } : {});
}

// --- Hooks ---

export function useSystemConfig() {
  return useQuery({
    queryKey: ['admin', 'system', 'config'],
    queryFn: fetchSystemConfig,
  });
}

export function useUpdateSystemConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (updates: UpdateSystemConfigInput) => updateSystemConfig(updates),
    onSuccess: () => {
      message.success('配置已更新');
      qc.invalidateQueries({ queryKey: ['admin', 'system', 'config'] });
    },
  });
}

export function useResetSystemConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (key?: string) => resetSystemConfig(key),
    onSuccess: () => {
      message.success('配置已重置');
      qc.invalidateQueries({ queryKey: ['admin', 'system', 'config'] });
    },
  });
}