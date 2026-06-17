import { browser } from '$app/environment';

/**
 * 记住「文件」tab 下最后一次访问的 URL（含子目录），
 * 用户从照片 / 媒体库切回文件时，可直接恢复到上次的位置。
 * 仅缓存 /files 路径，photos / media 永远跳到各自根路径。
 */

const LS_KEY = 'nd.lastFilesUrl';
const DEFAULT_FILES_URL = '/files/all';

function isFilesPath(pathname: string): boolean {
	return pathname === '/files' || pathname.startsWith('/files/');
}

export function rememberFilesUrl(pathname: string, search = ''): void {
	if (!browser) return;
	if (!isFilesPath(pathname)) return;
	try {
		localStorage.setItem(LS_KEY, pathname + search);
	} catch {
		// ignore quota errors
	}
}

export function getFilesUrl(): string {
	if (!browser) return DEFAULT_FILES_URL;
	try {
		const saved = localStorage.getItem(LS_KEY);
		if (saved && saved.startsWith('/files')) return saved;
	} catch {
		// ignore
	}
	return DEFAULT_FILES_URL;
}
