import { useQuery } from '@tanstack/react-query';
import { request } from '@/utils/request';
import type { AdminDashboardStats } from './types';

export type { AdminDashboardStats };

export async function fetchDashboardStats(): Promise<AdminDashboardStats> {
  return request.get<AdminDashboardStats>('/api/v1/admin/dashboard/stats');
}

export function useDashboardStats() {
  return useQuery({
    queryKey: ['admin', 'dashboard', 'stats'],
    queryFn: fetchDashboardStats,
  });
}