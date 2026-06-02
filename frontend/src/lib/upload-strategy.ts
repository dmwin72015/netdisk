export const SMALL_FILE_MAX_CHUNKS = 1;

export function getUploadChunkCount(fileSize: number, chunkSize: number): number {
	if (chunkSize <= 0) return 0;
	return Math.ceil(fileSize / chunkSize);
}

export function shouldUseSmallFileFastPath(fileSize: number, chunkSize: number): boolean {
	return fileSize > 0 && chunkSize > 0 && getUploadChunkCount(fileSize, chunkSize) <= SMALL_FILE_MAX_CHUNKS;
}
