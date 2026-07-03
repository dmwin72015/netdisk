import { useQuery } from '@tanstack/react-query';
import { request } from '@/utils/request';
import type { ActivityLogParams, AdminActionLabel, AdminActivityLog, AdminActivityLogList } from './types';

export type { AdminActivityLog, AdminActivityLogList, AdminActionLabel, ActivityLogParams };

// --- Raw API ---

export async function fetchActivityLogs(params?: ActivityLogParams): Promise<AdminActivityLogList> {
  return request.get<AdminActivityLogList>('/api/v1/admin/activity-logs', { params });
}

export async function fetchActivityLogActions(lang?: string): Promise<AdminActionLabel[]> {
  return request.get<AdminActionLabel[]>('/api/v1/admin/activity-logs/actions', {
    params: { lang },
  });
}

// --- Hooks ---

export function useActivityLogActions(lang?: string) {
  return useQuery({
    queryKey: ['admin', 'activity-logs', 'actions', lang],
    queryFn: () => fetchActivityLogActions(lang),
  });
}