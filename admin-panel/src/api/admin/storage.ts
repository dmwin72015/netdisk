import { request } from '../request';
import type { CategoryStat } from './types';

export async function fetchStorageStats(): Promise<CategoryStat[]> {
  return request.get<CategoryStat[]>('/api/v1/admin/storage/stats');
}
