import { request } from '../request';
import type { SystemConfigItem, UpdateSystemConfigInput } from './types';

export async function fetchSystemConfig(): Promise<SystemConfigItem[]> {
  return request.get<SystemConfigItem[]>('/api/v1/admin/system/config');
}

export async function updateSystemConfig(updates: UpdateSystemConfigInput): Promise<SystemConfigItem[]> {
  return request.put<SystemConfigItem[]>('/api/v1/admin/system/config', updates);
}

export async function resetSystemConfig(key?: string): Promise<SystemConfigItem[]> {
  return request.post<SystemConfigItem[]>('/api/v1/admin/system/config/reset', key ? { key } : {});
}
