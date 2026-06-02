const CODE_EXTENSIONS = new Set([
	'bash', 'bat', 'c', 'cc', 'cmd', 'conf', 'config', 'cpp', 'cs', 'css', 'cxx',
	'dart', 'dockerfile', 'env', 'go', 'graphql', 'h', 'hpp', 'htm', 'html', 'ini', 'java',
	'js', 'json', 'jsx', 'kt', 'kts', 'less', 'lua', 'm', 'md', 'mjs', 'mm', 'php',
	'pl', 'properties', 'ps1', 'py', 'r', 'rb', 'rs', 'sass', 'scala', 'scss', 'sh', 'sql',
	'svelte', 'swift', 'toml', 'ts', 'tsx', 'vue', 'xml', 'yaml', 'yml', 'zsh'
]);

const CODE_FILENAMES = new Set([
	'.bashrc', '.env', '.gitignore', '.npmrc', 'dockerfile', 'makefile', 'readme'
]);

export function getFileExtension(fileName: string): string {
	const normalized = fileName.trim().toLowerCase();
	const index = normalized.lastIndexOf('.');
	if (index <= 0 || index === normalized.length - 1) return '';
	return normalized.slice(index + 1);
}

export function isCodeLikeMimeType(mimeType: string | null | undefined): boolean {
	const value = mimeType ?? '';
	return (
		value === 'application/json' ||
		value === 'application/javascript' ||
		value === 'application/typescript' ||
		value === 'application/xml' ||
		value === 'application/x-sh' ||
		value === 'application/x-httpd-php' ||
		value === 'image/svg+xml' ||
		value === 'text/css' ||
		value === 'text/html' ||
		value === 'text/javascript' ||
		value === 'text/markdown' ||
		value === 'text/xml' ||
		value.startsWith('text/x-')
	);
}

export function isCodeLikeFile(fileName: string, mimeType?: string | null): boolean {
	const normalized = fileName.trim().toLowerCase();
	return isCodeLikeMimeType(mimeType) || CODE_FILENAMES.has(normalized) || CODE_EXTENSIONS.has(getFileExtension(normalized));
}

export function isJsonFile(fileName: string, mimeType?: string | null): boolean {
	return mimeType === 'application/json' || getFileExtension(fileName) === 'json';
}

export function isTextPreviewFile(fileName: string, mimeType?: string | null): boolean {
	return isCodeLikeFile(fileName, mimeType) || (mimeType ?? '').startsWith('text/');
}
