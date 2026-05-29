const CHUNK_SIZE = 4 * 1024 * 1024; // 4MB

export async function computeSHA256Chunked(
	file: File,
	callbacks: { onPreHash?: (hash: string) => void; onProgress?: (percent: number) => void }
): Promise<{ preHash: string; hash: string; totalChunks: number }> {
	const totalChunks = Math.ceil(file.size / CHUNK_SIZE);

	return new Promise((resolve, reject) => {
		let preHash = '';
		const worker = new Worker(
			new URL('$lib/workers/sha256.worker.ts', import.meta.url),
			{ type: 'module' }
		);
		worker.onmessage = (e: MessageEvent) => {
			if (e.data.type === 'pre_hash') {
				preHash = e.data.hash;
				callbacks.onPreHash?.(preHash);
			} else if (e.data.type === 'progress') {
				callbacks.onProgress?.(e.data.percent);
			} else if (e.data.type === 'complete') {
				resolve({ preHash, hash: e.data.hash, totalChunks });
				worker.terminate();
			}
		};
		worker.onerror = () => {
			reject(new Error('SHA-256 computation failed'));
			worker.terminate();
		};

		(async () => {
			for (let i = 0; i < totalChunks; i++) {
				const start = i * CHUNK_SIZE;
				const end = Math.min(start + CHUNK_SIZE, file.size);
				const buf = await file.slice(start, end).arrayBuffer();
				worker.postMessage({ type: 'chunk', index: i, data: buf }, [buf]);
			}
			worker.postMessage({ type: 'done', totalChunks });
		})();
	});
}
