import { request } from '../request';
import type { CleanupQueryInput, CleanupQueryResult, DeleteActionResult } from './types';

export async function queryCleanup(data: CleanupQueryInput): Promise<CleanupQueryResult> {
  return request.post<CleanupQueryResult>('/api/v1/admin/cleanup/query', data);
}

export async function deleteUserFile(userFileId: number): Promise<DeleteActionResult> {
  return request.post<DeleteActionResult>('/api/v1/admin/cleanup/delete-user-file', { userFileId });
}

export async function deletePhysicalFile(physicalFileId: number): Promise<DeleteActionResult> {
  return request.post<DeleteActionResult>('/api/v1/admin/cleanup/delete-physical-file', {
    physicalFileId,
  });
}
