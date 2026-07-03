import { useQuery } from '@tanstack/react-query';
import { request } from '@/utils/request';
import type { CategoryStat } from './types';

export type { CategoryStat };

export async function fetchStorageStats(): Promise<CategoryStat[]> {
  return request.get<CategoryStat[]>('/api/v1/admin/storage/stats');
}

export function useStorageStats() {
  return useQuery({
    queryKey: ['admin', 'storage', 'stats'],
    queryFn: fetchStorageStats,
  });
}