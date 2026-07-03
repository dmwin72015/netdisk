import { useQuery } from '@tanstack/react-query';
import { request } from '@/utils/request';
import type { AdminPhysicalFile, AdminPhysicalFileList, PhysicalFileDetailResult, PhysicalFileListParams } from './types';

export type { AdminPhysicalFile, AdminPhysicalFileList, PhysicalFileDetailResult, PhysicalFileListParams };

// --- Raw API ---

export async function fetchPhysicalFiles(params?: PhysicalFileListParams): Promise<AdminPhysicalFileList> {
  return request.get<AdminPhysicalFileList>('/api/v1/admin/files/physical', { params });
}

export async function fetchPhysicalFileDetail(id: string): Promise<PhysicalFileDetailResult> {
  return request.get<PhysicalFileDetailResult>(`/api/v1/admin/files/physical/${id}`);
}

// --- Hooks ---

export function usePhysicalFileDetail(id: string) {
  return useQuery({
    queryKey: ['admin', 'physical-file-detail', id],
    queryFn: () => fetchPhysicalFileDetail(id),
    enabled: !!id,
  });
}