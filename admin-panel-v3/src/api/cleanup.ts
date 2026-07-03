import { useMutation } from '@tanstack/react-query';
import { request } from '@/utils/request';
import type { CleanupQueryInput, CleanupQueryResult, CleanupQueryUserFile, DeleteActionResult } from './types';

export type { CleanupQueryInput, CleanupQueryResult, CleanupQueryUserFile, DeleteActionResult };

// --- Raw API ---

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

// --- Hooks ---

export function useCleanupQuery() {
  return useMutation({
    mutationFn: (data: CleanupQueryInput) => queryCleanup(data),
  });
}

export function useDeleteUserFile() {
  return useMutation({
    mutationFn: (userFileId: number) => deleteUserFile(userFileId),
  });
}

export function useDeletePhysicalFile() {
  return useMutation({
    mutationFn: (physicalFileId: number) => deletePhysicalFile(physicalFileId),
  });
}