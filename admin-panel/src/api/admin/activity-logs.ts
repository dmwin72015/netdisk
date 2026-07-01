import { request } from '../request';
import type { AdminActivityLogList, ActivityLogParams, AdminActionLabel } from './types';

export async function fetchActivityLogs(params?: ActivityLogParams): Promise<AdminActivityLogList> {
  return request.get<AdminActivityLogList>('/api/v1/admin/activity-logs', { params });
}

export async function fetchActivityLogActions(lang?: string): Promise<AdminActionLabel[]> {
  return request.get<AdminActionLabel[]>('/api/v1/admin/activity-logs/actions', {
    params: { lang },
  });
}
