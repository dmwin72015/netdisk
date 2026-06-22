/**
 * Paste text upload utility functions.
 *
 * Provides helpers for extracting text from clipboard data,
 * validating text size, creating File objects, and generating
 * default filenames for pasted text content.
 */

/** Maximum allowed pasted text size: 2MB */
export const MAX_PASTE_TEXT_SIZE = 2 * 1024 * 1024;

/**
 * Extract plain text from clipboard DataTransfer.
 * Returns null if clipboardData is null/undefined or the text is empty.
 */
export function extractClipboardText(
	clipboardData: DataTransfer | null | undefined,
): string | null {
	if (!clipboardData) return null;

	const text = clipboardData.getData('text/plain');
	return text ? text : null;
}

/**
 * Validate that text does not exceed the maximum allowed size.
 * Returns validation result with size in bytes and optional error message.
 */
export function validateTextSize(text: string): {
	valid: boolean;
	size: number;
	error?: string;
} {
	const size = new Blob([text]).size;

	if (size > MAX_PASTE_TEXT_SIZE) {
		return {
			valid: false,
			size,
			error: `Text size (${formatSize(size)}) exceeds the maximum allowed size (2MB)`,
		};
	}

	return { valid: true, size };
}

/**
 * Create a File object from text content with the given filename.
 */
export function createTextFile(text: string, fileName: string): File {
	return new File([text], fileName, { type: 'text/plain;charset=utf-8' });
}

/**
 * Generate a default filename from pasted text.
 * Uses the first line (truncated to 50 chars) or a timestamp if empty.
 * Removes illegal filename characters and ensures .txt extension.
 */
export function getDefaultFileName(text: string): string {
	// Extract first line
	const firstLine = text.split('\n')[0]?.trim() ?? '';

	// Use timestamp if no content
	if (!firstLine) {
		const d = new Date();
		const pad = (n: number) => String(n).padStart(2, '0');
		return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}_${pad(d.getHours())}-${pad(d.getMinutes())}-${pad(d.getSeconds())}.txt`;
	}

	// Truncate to 50 chars
	const truncated = firstLine.length > 50 ? firstLine.slice(0, 50) : firstLine;

	// Remove illegal filename characters: \ / : * ? " < > |
	const cleaned = truncated.replace(/[\\/:*?"<>|]/g, '');

	// Ensure .txt extension
	return cleaned.endsWith('.txt') ? cleaned : `${cleaned}.txt`;
}

/** Format bytes to human-readable size for error messages */
function formatSize(bytes: number): string {
	if (bytes < 1024) return `${bytes} B`;
	if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
	return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
