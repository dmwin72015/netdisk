import type { NormalizedFile } from "$lib/types/file";
import { isTextPreviewFile } from "$lib/utils/code-files";
import { thumbnailUrl } from "$lib/api/photos";
import { getAccessToken } from "$lib/api/client";

export function isImageFile(file: NormalizedFile): boolean {
  if (file.fileCategory === "image") return true;
  if (file.mimeType?.startsWith("image/")) return true;
  return /\.(avif|bmp|gif|heic|heif|jpe?g|png|svg|webp)$/i.test(file.name);
}

export function isVideoFile(file: NormalizedFile): boolean {
  return file.mimeType?.startsWith("video/") ?? false;
}

export function canPreview(file: NormalizedFile): boolean {
  const mimeType = file.mimeType ?? "";
  return (
    mimeType.startsWith("image/") ||
    mimeType.startsWith("video/") ||
    mimeType.startsWith("audio/") ||
    mimeType === "application/pdf" ||
    isTextPreviewFile(file.name, mimeType)
  );
}

export function authedFileUrl(file: NormalizedFile, downloadUrlFn: (id: string) => string): string {
  const token = getAccessToken();
  const url = downloadUrlFn(file.id);
  if (!token) return url;
  const sep = url.includes("?") ? "&" : "?";
  return `${url}${sep}access_token=${encodeURIComponent(token)}`;
}

export function authedThumbnailUrl(file: NormalizedFile): string {
  const token = getAccessToken();
  const url = thumbnailUrl(file.id);
  if (!token) return url;
  const sep = url.includes("?") ? "&" : "?";
  return `${url}${sep}access_token=${encodeURIComponent(token)}`;
}
