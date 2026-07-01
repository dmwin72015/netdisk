import { request } from '../request';
import type { AdminDashboardStats } from './types';

export async function fetchDashboardStats(): Promise<AdminDashboardStats> {
  return request.get<AdminDashboardStats>('/api/v1/admin/dashboard/stats');
}
