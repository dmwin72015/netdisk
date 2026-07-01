import { request } from '../request';
import type {
  AdminPhysicalFileList,
  PhysicalFileListParams,
  PhysicalFileDetailResult,
} from './types';

export async function fetchPhysicalFiles(params?: PhysicalFileListParams): Promise<AdminPhysicalFileList> {
  return request.get<AdminPhysicalFileList>('/api/v1/admin/files/physical', { params });
}

export async function fetchPhysicalFileDetail(id: number): Promise<PhysicalFileDetailResult> {
  return request.get<PhysicalFileDetailResult>(`/api/v1/admin/files/physical/${id}`);
}
