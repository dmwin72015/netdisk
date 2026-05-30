import { getAccessToken } from '$lib/api/client';
import * as m from '$lib/paraglide/messages';

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

/** Format duration as h:mm:ss or m:ss. Returns null if invalid. */
export function fmtDurationHMS(sec?: number): string | null {
	if (!sec || sec <= 0) return null;
	const h = Math.floor(sec / 3600);
	const mi = Math.floor((sec % 3600) / 60);
	const s = sec % 60;
	if (h > 0) return `${h}:${String(mi).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
	return `${mi}:${String(s).padStart(2, '0')}`;
}

/** Format duration as localized text: "Xh Ymin", "Xmin Ys", "Xs". */
export function fmtDurationText(sec?: number): string {
	if (!sec || sec <= 0) return m.duration_seconds();
	const s = Math.round(sec);
	if (s >= 3600) {
		const h = Math.floor(s / 3600);
		const mi = Math.floor((s % 3600) / 60);
		return mi > 0 ? m.duration_h_m({ h, m: mi }) : m.duration_h({ h });
	}
	if (s >= 60) {
		const mi = Math.floor(s / 60);
		const r = s % 60;
		return r > 0 ? m.duration_m_s({ m: mi, s: r }) : m.duration_m({ m: mi });
	}
	return m.duration_s({ s });
}

/** Relative time string from a Unix timestamp (seconds). */
export function timeAgo(unix: number): string {
	const diff = Math.max(0, Date.now() / 1000 - unix);
	if (diff < 60) return m.just_now();
	if (diff < 3600) return m.minutes_ago({ n: Math.floor(diff / 60) });
	if (diff < 86400) return m.hours_ago({ n: Math.floor(diff / 3600) });
	if (diff < 86400 * 30) return m.days_ago({ n: Math.floor(diff / 86400) });
	if (diff < 86400 * 365) return m.months_ago({ n: Math.floor(diff / 86400 / 30) });
	return m.years_ago({ n: Math.floor(diff / 86400 / 365) });
}

/** Append access_token query param to a URL for authenticated resource loading. */
export function authedUrl(url: string): string {
	const token = getAccessToken();
	if (!token) return url;
	const u = new URL(url, window.location.origin);
	u.searchParams.set('access_token', token);
	return u.pathname + '?' + u.searchParams.toString();
}
