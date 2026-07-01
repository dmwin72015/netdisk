export function formatBytes(b: number): string {
  if (b === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(b) / Math.log(k));
  return parseFloat((b / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export function formatDate(epoch: number): string {
  return new Date(epoch * 1000).toLocaleString();
}

export function formatDateShort(epoch: number): string {
  return new Date(epoch * 1000).toLocaleDateString();
}

export function formatISODate(iso: string): string {
  return new Date(iso).toLocaleString();
}