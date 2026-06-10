import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));
vi.mock('$lib/paraglide/messages', () => ({}));

import {
	listFiles,
	listRecentFiles,
	mkdir,
	checkConflict,
	checkDuplicate,
	importFile,
	trashFile,
	restoreFile,
	permanentDelete,
	emptyTrash,
	restoreAll,
	renameFile,
	moveFile,
	setStarred,
	listTrashed,
	listStarred,
	downloadUrl,
	getBreadcrumb,
} from './files';

// ── helpers ────────────────────────────────────────────────────────

let store: Record<string, string>;

function jsonResponse(data: unknown, status = 200): Response {
	return new Response(JSON.stringify({ data }), {
		status,
		headers: { 'Content-Type': 'application/json' },
	});
}

beforeEach(() => {
	store = {};
	vi.stubGlobal('localStorage', {
		getItem: (k: string) => store[k] ?? null,
		setItem: (k: string, v: string) => { store[k] = v; },
		removeItem: (k: string) => { delete store[k]; },
		clear: () => { store = {}; },
	});
});

afterEach(() => {
	vi.restoreAllMocks();
});

// ── listFiles ─────────────────────────────────────────────────────

describe('listFiles', () => {
	it('builds query with default params', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ files: [], total: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await listFiles();

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('/api/v1/files?');
		expect(url).toContain('page=1');
		expect(url).toContain('pageSize=50');
	});

	it('includes parentSlug and filters when provided', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ files: [], total: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await listFiles('parent-1', 2, 10, 'image/jpeg', 'image', 'name', 'asc', true);

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('parentSlug=parent-1');
		expect(url).toContain('page=2');
		expect(url).toContain('pageSize=10');
		expect(url).toContain('mimeType=image%2Fjpeg');
		expect(url).toContain('fileCategory=image');
		expect(url).toContain('sortBy=name');
		expect(url).toContain('sortDir=asc');
		expect(url).toContain('onlyDirs=true');
	});

	it('can hide system directories', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ files: [], total: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await listFiles(undefined, 1, 50, undefined, undefined, undefined, undefined, false, false);

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('includeSystem=false');
	});

	it('includes search query when provided', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ files: [], total: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await listFiles(undefined, 1, 20, undefined, null, 'created_at', 'DESC', false, false, 'hello world');

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('searchQuery=hello+world');
		expect(url).toContain('sortBy=created_at');
		expect(url).toContain('sortDir=DESC');
		expect(url).toContain('includeSystem=false');
	});
});

// ── listRecentFiles ───────────────────────────────────────────────

describe('listRecentFiles', () => {
	it('uses default limit', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ files: [], total: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await listRecentFiles();

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('/api/v1/files/recent?limit=10');
	});

	it('uses custom limit', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ files: [], total: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await listRecentFiles(25);

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('limit=25');
	});
});

// ── mkdir ─────────────────────────────────────────────────────────

describe('mkdir', () => {
	it('posts dirName and parentSlug', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ slug: 'new-dir', fileName: 'New Dir', isDir: true }));
		vi.stubGlobal('fetch', fetchSpy);

		await mkdir('New Dir', 'parent-1');

		const [, init] = fetchSpy.mock.calls[0];
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({ dirName: 'New Dir', parentSlug: 'parent-1' });
	});

	it('defaults parentSlug to empty string', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ slug: 'new-dir' }));
		vi.stubGlobal('fetch', fetchSpy);

		await mkdir('Test');

		expect(JSON.parse(fetchSpy.mock.calls[0][1].body).parentSlug).toBe('');
	});
});

// ── checkConflict ─────────────────────────────────────────────────

describe('checkConflict', () => {
	it('posts conflict check params', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ status: 'OK' }));
		vi.stubGlobal('fetch', fetchSpy);

		await checkConflict('file.txt', 1024, 'hash123', 'parent-1');

		const [, init] = fetchSpy.mock.calls[0];
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({
			fileName: 'file.txt',
			fileSize: 1024,
			preHash: 'hash123',
			parentSlug: 'parent-1',
		});
	});
});

// ── checkDuplicate ────────────────────────────────────────────────

describe('checkDuplicate', () => {
	it('posts fileHash and parentSlug', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ status: 'OK' }));
		vi.stubGlobal('fetch', fetchSpy);

		await checkDuplicate('sha256hash', 'parent-1');

		const body = JSON.parse(fetchSpy.mock.calls[0][1].body);
		expect(body).toEqual({ fileHash: 'sha256hash', parentSlug: 'parent-1' });
	});
});

// ── importFile ────────────────────────────────────────────────────

describe('importFile', () => {
	it('posts import params', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ fileSlug: 'new-slug', fileName: 'imported.txt' }));
		vi.stubGlobal('fetch', fetchSpy);

		await importFile('phys-slug', 'imported.txt', 'parent-1');

		const body = JSON.parse(fetchSpy.mock.calls[0][1].body);
		expect(body).toEqual({
			physicalFileSlug: 'phys-slug',
			fileName: 'imported.txt',
			parentSlug: 'parent-1',
		});
	});
});

// ── trashFile ─────────────────────────────────────────────────────

describe('trashFile', () => {
	it('sends DELETE to file slug', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		await trashFile('file-abc');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/files/file-abc');
		expect(init.method).toBe('DELETE');
	});
});

// ── restoreFile ───────────────────────────────────────────────────

describe('restoreFile', () => {
	it('sends POST to restore endpoint', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		await restoreFile('file-abc');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/files/file-abc/restore');
		expect(init.method).toBe('POST');
	});
});

// ── permanentDelete ───────────────────────────────────────────────

describe('permanentDelete', () => {
	it('sends DELETE to permanent endpoint', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		await permanentDelete('file-abc');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/files/file-abc/permanent');
		expect(init.method).toBe('DELETE');
	});
});

// ── emptyTrash ────────────────────────────────────────────────────

describe('emptyTrash', () => {
	it('sends POST to trash/empty', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ deleted: 5 }));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await emptyTrash();

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/files/trash/empty');
		expect(init.method).toBe('POST');
		expect(result).toEqual({ deleted: 5 });
	});
});

// ── restoreAll ────────────────────────────────────────────────────

describe('restoreAll', () => {
	it('sends POST to trash/restore-all', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ restored: 3 }));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await restoreAll();

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/files/trash/restore-all');
		expect(init.method).toBe('POST');
		expect(result).toEqual({ restored: 3 });
	});
});

// ── renameFile ────────────────────────────────────────────────────

describe('renameFile', () => {
	it('posts new name', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		await renameFile('file-abc', 'new-name.txt');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/files/file-abc/rename');
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({ newName: 'new-name.txt' });
	});
});

// ── moveFile ──────────────────────────────────────────────────────

describe('moveFile', () => {
	it('posts target parent slug', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		await moveFile('file-abc', 'target-dir');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/files/file-abc/move');
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({ targetParentSlug: 'target-dir' });
	});
});

// ── setStarred ────────────────────────────────────────────────────

describe('setStarred', () => {
	it('posts starred state', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		await setStarred('file-abc', true);

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/files/file-abc/star');
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({ starred: true });
	});

	it('can unstar', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		await setStarred('file-abc', false);

		expect(JSON.parse(fetchSpy.mock.calls[0][1].body)).toEqual({ starred: false });
	});
});

// ── listTrashed ───────────────────────────────────────────────────

describe('listTrashed', () => {
	it('builds query with default params', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ files: [], total: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await listTrashed();

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('/api/v1/files/trash?');
		expect(url).toContain('page=1');
		expect(url).toContain('pageSize=50');
	});

	it('uses custom pagination', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ files: [], total: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await listTrashed(3, 20);

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('page=3');
		expect(url).toContain('pageSize=20');
	});
});

// ── listStarred ───────────────────────────────────────────────────

describe('listStarred', () => {
	it('builds query with default params', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ files: [], total: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await listStarred();

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('/api/v1/files/starred?');
		expect(url).toContain('page=1');
		expect(url).toContain('pageSize=50');
	});
});

// ── downloadUrl ───────────────────────────────────────────────────

describe('downloadUrl', () => {
	it('returns correct path', () => {
		expect(downloadUrl('file-abc')).toBe('/api/v1/files/file-abc/download');
	});
});

// ── getBreadcrumb ─────────────────────────────────────────────────

describe('getBreadcrumb', () => {
	it('fetches breadcrumb for slug', async () => {
		const breadcrumb = [
			{ slug: 'root', fileName: 'Root' },
			{ slug: 'folder-1', fileName: 'Folder 1' },
		];
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(breadcrumb));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await getBreadcrumb('folder-1');

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toBe('/api/v1/files/folder-1/breadcrumb');
		expect(result).toEqual(breadcrumb);
	});
});
