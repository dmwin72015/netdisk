import { api } from './client';

export type FileItem = {
	slug: string;
	fileName: string;
	isDir: boolean;
	fileSize: number;
	mimeType: string | null;
	fileCategory: string;
	isStarred: boolean;
	parentSlug?: string;
	parentName?: string;
	createdAt: string;
	updatedAt: string;
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
	fileSlug: string;
	fileName: string;
};

export type BreadcrumbItem = {
	slug: string;
	fileName: string;
};

export async function listFiles(
	parentSlug?: string,
	page = 1,
	pageSize = 50,
	mimeType?: string,
	fileCategory?: string | null,
	sortBy?: string,
	sortDir?: string
) {
	const params = new URLSearchParams({ page: String(page), pageSize: String(pageSize) });
	if (parentSlug) params.set('parentSlug', parentSlug);
	if (mimeType) params.set('mimeType', mimeType);
	if (fileCategory) params.set('fileCategory', fileCategory);
	if (sortBy) params.set('sortBy', sortBy);
	if (sortDir) params.set('sortDir', sortDir);
	return api<{ files: FileItem[]; total: number }>(`/api/v1/files?${params}`);
}

export async function listRecentFiles(limit = 10) {
	return api<{ files: FileItem[]; total: number }>(`/api/v1/files/recent?limit=${limit}`);
}

export async function mkdir(dirName: string, parentSlug?: string) {
	return api<FileItem>('/api/v1/files/mkdir', {
		method: 'POST',
		body: JSON.stringify({ dirName, parentSlug: parentSlug || '' })
	});
}

export async function checkConflict(fileName: string, fileSize: number, preHash: string, parentSlug?: string) {
	return api<ConflictResponse>('/api/v1/files/check-conflict', {
		method: 'POST',
		body: JSON.stringify({ fileName, fileSize, preHash, parentSlug: parentSlug || '' })
	});
}

export async function checkDuplicate(fileHash: string, parentSlug?: string) {
	return api<DuplicateResponse>('/api/v1/files/check-duplicate', {
		method: 'POST',
		body: JSON.stringify({ fileHash, parentSlug: parentSlug || '' })
	});
}

export async function importFile(physicalFileSlug: string, fileName: string, parentSlug?: string) {
	return api<ImportResponse>('/api/v1/files/import', {
		method: 'POST',
		body: JSON.stringify({ physicalFileSlug, fileName, parentSlug: parentSlug || '' })
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

export async function emptyTrash() {
	return api<{ deleted: number }>('/api/v1/files/trash/empty', { method: 'POST' });
}

export async function restoreAll() {
	return api<{ restored: number }>('/api/v1/files/trash/restore-all', { method: 'POST' });
}

export async function renameFile(slug: string, newName: string) {
	return api(`/api/v1/files/${slug}/rename`, {
		method: 'POST',
		body: JSON.stringify({ newName })
	});
}

export async function moveFile(slug: string, targetParentSlug: string) {
	return api(`/api/v1/files/${slug}/move`, {
		method: 'POST',
		body: JSON.stringify({ targetParentSlug })
	});
}

export async function setStarred(slug: string, starred: boolean) {
	return api(`/api/v1/files/${slug}/star`, {
		method: 'POST',
		body: JSON.stringify({ starred })
	});
}

export async function listTrashed(page = 1, pageSize = 50) {
	return api<{ files: FileItem[]; total: number }>(`/api/v1/files/trash?page=${page}&pageSize=${pageSize}`);
}

export async function listStarred(page = 1, pageSize = 50) {
	return api<{ files: FileItem[]; total: number }>(`/api/v1/files/starred?page=${page}&pageSize=${pageSize}`);
}

export function downloadUrl(slug: string) {
	return `/api/v1/files/${slug}/download`;
}

export async function getBreadcrumb(slug: string) {
	return api<BreadcrumbItem[]>(`/api/v1/files/${slug}/breadcrumb`);
}
