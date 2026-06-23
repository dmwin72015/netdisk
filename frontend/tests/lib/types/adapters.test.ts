import { describe, it, expect } from 'vitest';

import { normalizeFileItem, normalizeDriveFile } from '$lib/types/adapters';
import type { FileItem } from '$lib/api/files';
import type { DriveFile } from '$lib/api/drive';
import type { NormalizedFile } from '$lib/types/file';

// ── normalizeFileItem ─────────────────────────────────────────────

describe('normalizeFileItem', () => {
	const fileItem: FileItem = {
		slug: 'file-abc',
		fileName: 'photo.jpg',
		isDir: false,
		isLocked: false,
		fileSize: 1024,
		mimeType: 'image/jpeg',
		fileCategory: 'image',
		isStarred: true,
		isSystem: false,
		createdAt: '2025-01-01T00:00:00Z',
		updatedAt: '2025-06-15T12:00:00Z',
	};

	it('maps FileItem fields to NormalizedFile', () => {
		const result = normalizeFileItem(fileItem);
		const expected: NormalizedFile = {
			id: 'file-abc',
			name: 'photo.jpg',
			isDir: false,
			isLocked: false,
			size: 1024,
			mimeType: 'image/jpeg',
			fileCategory: 'image',
			isStarred: true,
			isSystem: false,
			createdAt: '2025-01-01T00:00:00Z',
			updatedAt: '2025-06-15T12:00:00Z',
		};
		expect(result).toEqual(expected);
	});

	it('handles directory', () => {
		const dir: FileItem = { ...fileItem, slug: 'dir-1', fileName: 'My Folder', isDir: true, mimeType: null };
		const result = normalizeFileItem(dir);
		expect(result.id).toBe('dir-1');
		expect(result.name).toBe('My Folder');
		expect(result.isDir).toBe(true);
		expect(result.mimeType).toBeNull();
	});

	it('handles null mimeType', () => {
		const item: FileItem = { ...fileItem, mimeType: null };
		expect(normalizeFileItem(item).mimeType).toBeNull();
	});
});

// ── normalizeDriveFile ────────────────────────────────────────────

describe('normalizeDriveFile', () => {
	const driveFile: DriveFile = {
		id: 'drive-xyz',
		name: 'video.mp4',
		mimeType: 'video/mp4',
		fileCategory: 'video',
		size: 5000000,
		createdAt: 1700000000, // Unix timestamp in seconds
		isDir: false,
	};

	it('maps DriveFile fields to NormalizedFile', () => {
		const result = normalizeDriveFile(driveFile);
		expect(result.id).toBe('drive-xyz');
		expect(result.name).toBe('video.mp4');
		expect(result.isDir).toBe(false);
		expect(result.size).toBe(5000000);
		expect(result.mimeType).toBe('video/mp4');
		expect(result.fileCategory).toBe('video');
		expect(result.isStarred).toBe(false);
	});

	it('converts createdAt unix timestamp to ISO string', () => {
		const result = normalizeDriveFile(driveFile);
		const expectedDate = new Date(1700000000 * 1000).toISOString();
		expect(result.createdAt).toBe(expectedDate);
		expect(result.updatedAt).toBe(expectedDate);
	});

	it('always sets isStarred to false', () => {
		expect(normalizeDriveFile(driveFile).isStarred).toBe(false);
	});

	it('handles empty mimeType as null', () => {
		const file: DriveFile = { ...driveFile, mimeType: '' };
		const result = normalizeDriveFile(file);
		expect(result.mimeType).toBeNull();
	});

	it('handles directory', () => {
		const dir: DriveFile = { ...driveFile, id: 'dir-2', name: 'Docs', isDir: true, mimeType: '' };
		const result = normalizeDriveFile(dir);
		expect(result.isDir).toBe(true);
		expect(result.name).toBe('Docs');
	});
});
