import dayjs from 'dayjs';

export function formatBytes(b: number): string {
  if (b === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(b) / Math.log(k));
  return parseFloat((b / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export function formatDate(epoch: number): string {
  return dayjs(epoch * 1000).format('YYYY-MM-DD HH:mm:ss');
}

export function formatDateShort(epoch: number): string {
  return dayjs(epoch * 1000).format('YYYY-MM-DD');
}

export function formatISODate(iso: string): string {
  return dayjs(iso).format('YYYY-MM-DD HH:mm:ss');
}
