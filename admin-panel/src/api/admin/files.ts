import { request } from '../request';
import type { AdminFileList, FileListParams } from './types';

export async function fetchFiles(params?: FileListParams): Promise<AdminFileList> {
  return request.get<AdminFileList>('/api/v1/admin/files', { params });
}

export async function deleteFile(id: string): Promise<void> {
  await request.delete<void>(`/api/v1/admin/files/${id}`);
}

export async function restoreFile(id: string): Promise<void> {
  await request.patch<void>(`/api/v1/admin/files/${id}/restore`);
}
