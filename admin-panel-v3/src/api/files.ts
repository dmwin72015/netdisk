import { message } from 'antd';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { request } from '@/utils/request';
import type { AdminFile, AdminFileList, DeleteActionResult, FileListParams } from './types';

export type { AdminFile, AdminFileList, DeleteActionResult, FileListParams };

// --- Raw API ---

export async function fetchFiles(params?: FileListParams): Promise<AdminFileList> {
  return request.get<AdminFileList>('/api/v1/admin/files', { params });
}

export async function deleteFile(id: string): Promise<DeleteActionResult> {
  return request.delete<DeleteActionResult>(`/api/v1/admin/files/${id}`);
}

export async function restoreFile(id: string): Promise<DeleteActionResult> {
  return request.patch<DeleteActionResult>(`/api/v1/admin/files/${id}/restore`);
}

// --- Hooks ---

export function useDeleteFile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteFile(id),
    onSuccess: () => {
      message.success('文件已删除');
      qc.invalidateQueries({ queryKey: ['admin', 'files'] });
    },
  });
}

export function useRestoreFile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => restoreFile(id),
    onSuccess: () => {
      message.success('文件已恢复');
      qc.invalidateQueries({ queryKey: ['admin', 'files'] });
    },
  });
}