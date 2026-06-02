export function normalizeCompletedChunks(completedChunks: number[] | undefined, totalChunks: number): Set<number> {
	const chunks = new Set<number>();
	for (const chunk of completedChunks ?? []) {
		if (!Number.isInteger(chunk) || chunk < 0 || chunk >= totalChunks) continue;
		chunks.add(chunk);
	}
	return chunks;
}

export function getUploadedBytesFromCompletedChunks(completedChunks: Set<number>, totalChunks: number, chunkSize: number, fileSize: number): number {
	let uploaded = 0;
	for (const chunk of completedChunks) {
		if (chunk < 0 || chunk >= totalChunks) continue;
		const start = chunk * chunkSize;
		const end = Math.min(start + chunkSize, fileSize);
		if (end > start) uploaded += end - start;
	}
	return uploaded;
}

export function getUploadProgress(uploadedBytes: number, fileSize: number): number {
	if (fileSize <= 0) return 0;
	return Math.min(100, Math.round((uploadedBytes / fileSize) * 100));
}
