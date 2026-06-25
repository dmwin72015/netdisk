import type { NormalizedFile } from "$lib/types/file";

export type PreviewInfo = {
  slug: string;
  name: string;
  mimeType: string;
  size: number;
};

class PreviewManager {
  previewFile = $state<PreviewInfo | null>(null);

  get isOpen(): boolean {
    return this.previewFile !== null;
  }

  open(file: NormalizedFile): void {
    this.previewFile = {
      slug: file.id,
      name: file.name,
      mimeType: file.mimeType || "",
      size: file.size,
    };
  }

  openRaw(info: PreviewInfo): void {
    this.previewFile = info;
  }

  close(): void {
    this.previewFile = null;
  }
}

export const previewManager = new PreviewManager();
