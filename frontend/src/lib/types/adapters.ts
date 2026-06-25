import type { FileItem } from '$lib/api/files';
import type { DriveFile } from '$lib/api/drive';
import type { NormalizedFile } from './file';

export function normalizeFileItem(f: FileItem): NormalizedFile {
	return {
		id: f.slug,
		slug: f.slug,
		name: f.fileName,
		isDir: f.isDir,
		hasPassword: f.hasPassword,
		size: f.fileSize,
		mimeType: f.mimeType,
		fileCategory: f.fileCategory,
		isStarred: f.isStarred,
		isSystem: f.isSystem,
		systemKind: f.systemKind,
		createdAt: f.createdAt,
		updatedAt: f.updatedAt,
		hashAlgo: f.hashAlgo,
		fileHash: f.fileHash,
	};
}

export function normalizeDriveFile(f: DriveFile): NormalizedFile {
	return {
		id: f.id,
		slug: f.id,
		name: f.name,
		isDir: f.isDir,
		hasPassword: false,
		size: f.size,
		mimeType: f.mimeType || null,
		fileCategory: f.fileCategory,
		isStarred: false,
		isSystem: false,
		createdAt: new Date(f.createdAt * 1000).toISOString(),
		updatedAt: new Date(f.createdAt * 1000).toISOString(),
	};
}
