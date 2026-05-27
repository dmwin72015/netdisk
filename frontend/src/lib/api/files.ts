import { api } from './client';

export type FileItem = {
	slug: string;
	file_name: string;
	is_dir: boolean;
	file_size: number;
	mime_type: string | null;
	is_starred: boolean;
	created_at: string;
	updated_at: string;
};

export type ConflictResponse = {
	status: 'OK' | 'NAME_CONFLICT' | 'SAME_FILE_CONFLICT';
	message?: string;
	existing?: FileItem;
};

export type DuplicateResponse = {
	status: 'OK' | 'DUPLICATE_FILE';
	message?: string;
	existing?: FileItem;
};

export type ImportResponse = {
	file_slug: string;
	file_name: string;
};

export async function listFiles(parentSlug?: string, page = 1, pageSize = 50) {
	const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
	if (parentSlug) params.set('parent_slug', parentSlug);
	return api<{ files: FileItem[]; total: number }>(`/api/v1/files?${params}`);
}

export async function mkdir(dirName: string, parentSlug?: string) {
	return api<FileItem>('/api/v1/files/mkdir', {
		method: 'POST',
		body: JSON.stringify({ dir_name: dirName, parent_slug: parentSlug || '' })
	});
}

export async function checkConflict(fileName: string, fileSize: number, preHash: string, parentSlug?: string) {
	return api<ConflictResponse>('/api/v1/files/check-conflict', {
		method: 'POST',
		body: JSON.stringify({ file_name: fileName, file_size: fileSize, pre_hash: preHash, parent_slug: parentSlug || '' })
	});
}

export async function checkDuplicate(fileHash: string, parentSlug?: string) {
	return api<DuplicateResponse>('/api/v1/files/check-duplicate', {
		method: 'POST',
		body: JSON.stringify({ file_hash: fileHash, parent_slug: parentSlug || '' })
	});
}

export async function importFile(physicalFileSlug: string, fileName: string, parentSlug?: string) {
	return api<ImportResponse>('/api/v1/files/import', {
		method: 'POST',
		body: JSON.stringify({ physical_file_slug: physicalFileSlug, file_name: fileName, parent_slug: parentSlug || '' })
	});
}

export async function trashFile(slug: string) {
	return api(`/api/v1/files/${slug}`, { method: 'DELETE' });
}

export async function restoreFile(slug: string) {
	return api(`/api/v1/files/${slug}/restore`, { method: 'POST' });
}

export async function permanentDelete(slug: string) {
	return api(`/api/v1/files/${slug}/permanent`, { method: 'DELETE' });
}

export async function renameFile(slug: string, newName: string) {
	return api(`/api/v1/files/${slug}/rename`, {
		method: 'POST',
		body: JSON.stringify({ new_name: newName })
	});
}

export async function moveFile(slug: string, targetParentSlug: string) {
	return api(`/api/v1/files/${slug}/move`, {
		method: 'POST',
		body: JSON.stringify({ target_parent_slug: targetParentSlug })
	});
}

export async function setStarred(slug: string, starred: boolean) {
	return api(`/api/v1/files/${slug}/star`, {
		method: 'POST',
		body: JSON.stringify({ starred })
	});
}

export async function listTrashed(page = 1, pageSize = 50) {
	return api<{ files: FileItem[]; total: number }>(`/api/v1/files/trash?page=${page}&page_size=${pageSize}`);
}

export async function listStarred(page = 1, pageSize = 50) {
	return api<{ files: FileItem[]; total: number }>(`/api/v1/files/starred?page=${page}&page_size=${pageSize}`);
}

export function downloadUrl(slug: string) {
	return `/api/v1/files/${slug}/download`;
}
