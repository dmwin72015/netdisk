import { api } from './client';

export type ShareFileItem = {
	slug: string;
	fileSlug: string;
	fileName: string;
	fileSize: number;
	mimeType: string | null;
	fileCategory: string;
	parentSlug: string;
};

export type ShareItem = {
	slug: string;
	files: ShareFileItem[];
	hasPassword: boolean;
	passwordCode?: string;
	expiresAt: string | null;
	disabledAt: string | null;
	createdAt: string;
	updatedAt: string;
	isExpired: boolean;
};

export type PublicShareInfo = {
	slug: string;
	files: ShareFileItem[];
	hasPassword: boolean;
	expiresAt: string | null;
	createdAt: string;
};

export async function createShare(input: {
	fileSlugs: string[];
	expiresAt?: string | null;
	passwordCode?: string | null;
}) {
	return api<ShareItem>('/api/v1/shares', {
		method: 'POST',
		body: JSON.stringify(input)
	});
}

export async function listShares(page = 1, pageSize = 50) {
	return api<{ shares: ShareItem[]; total: number }>(`/api/v1/shares?page=${page}&pageSize=${pageSize}`);
}

export async function updateShare(
	slug: string,
	input: { expiresAt?: string | null; passwordCode?: string | null }
) {
	return api<ShareItem>(`/api/v1/shares/${slug}`, {
		method: 'PATCH',
		body: JSON.stringify(input)
	});
}

export async function cancelShare(slug: string) {
	return api(`/api/v1/shares/${slug}`, { method: 'DELETE' });
}

export async function deleteShare(slug: string) {
	return api(`/api/v1/shares/${slug}?permanent=1`, { method: 'DELETE' });
}

export async function getPublicShare(slug: string) {
	return api<PublicShareInfo>(`/api/v1/public/shares/${slug}`, { auth: false });
}

export async function verifyPublicShare(slug: string, passwordCode: string) {
	return api<PublicShareInfo>(`/api/v1/public/shares/${slug}/verify`, {
		auth: false,
		method: 'POST',
		body: JSON.stringify({ passwordCode })
	});
}

export function publicShareFileUrl(slug: string, fileSlug?: string, passwordCode?: string, download = false) {
	const params = new URLSearchParams();
	if (fileSlug) params.set('fileSlug', fileSlug);
	if (passwordCode) params.set('code', passwordCode);
	if (download) params.set('download', '1');
	const query = params.toString();
	return `/api/v1/public/shares/${slug}/file${query ? `?${query}` : ''}`;
}
