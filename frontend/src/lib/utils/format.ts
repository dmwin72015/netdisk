export function fmtSize(bytes: number): string {
	if (bytes === 0) return '0 B';
	const k = 1024;
	const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
	const i = Math.floor(Math.log(bytes) / Math.log(k));
	const decimals = i >= 3 ? 2 : i > 0 ? 1 : 0;
	return (bytes / Math.pow(k, i)).toFixed(decimals) + ' ' + sizes[i];
}

export function fmtTime(iso: string): string {
	const d = new Date(iso);
	const pad = (n: number) => String(n).padStart(2, '0');
	return `${d.getFullYear()}/${pad(d.getMonth() + 1)}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

export function fmtSpeed(bytesPerSec: number): string {
	if (bytesPerSec <= 0) return '';
	if (bytesPerSec >= 1024 * 1024) return (bytesPerSec / (1024 * 1024)).toFixed(1) + ' MB/s';
	if (bytesPerSec >= 1024) return (bytesPerSec / 1024).toFixed(1) + ' KB/s';
	return bytesPerSec + ' B/s';
}
