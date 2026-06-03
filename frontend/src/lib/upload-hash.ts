const HASH_CHUNK_SIZE = 16 * 1024 * 1024; // 16MB for hash I/O

export async function computeSHA256(file: File): Promise<string> {
	const buf = await file.arrayBuffer();
	const hash = await crypto.subtle.digest('SHA-256', buf);
	return Array.from(new Uint8Array(hash))
		.map((b) => b.toString(16).padStart(2, '0'))
		.join('');
}

export async function computeSHA256Chunked(
	file: File,
	callbacks: { onPreHash?: (hash: string) => void; onProgress?: (percent: number) => void },
	chunkSize?: number
): Promise<{ preHash: string; hash: string; totalChunks: number }> {
	if (!file || file.size === undefined) {
		throw new Error('invalid file');
	}
	const totalChunks = Math.ceil(file.size / (chunkSize ?? 4 * 1024 * 1024));

	return new Promise((resolve, reject) => {
		let settled = false;
		let preHash = '';
		const worker = new Worker(
			new URL('$lib/workers/sha256.worker.ts', import.meta.url),
			{ type: 'module' }
		);
		worker.onmessage = (e: MessageEvent) => {
			if (settled) return;
			if (e.data.type === 'pre_hash') {
				preHash = e.data.hash;
				callbacks.onPreHash?.(preHash);
			} else if (e.data.type === 'progress') {
				callbacks.onProgress?.(e.data.percent);
			} else if (e.data.type === 'complete') {
				settled = true;
				resolve({ preHash, hash: e.data.hash, totalChunks });
				worker.terminate();
			}
		};
		worker.onerror = () => {
			if (settled) return;
			settled = true;
			reject(new Error('SHA-256 computation failed'));
			worker.terminate();
		};

		(async () => {
			try {
				const hashChunks = Math.ceil(file.size / HASH_CHUNK_SIZE);
				for (let i = 0; i < hashChunks; i++) {
					const start = i * HASH_CHUNK_SIZE;
					const end = Math.min(start + HASH_CHUNK_SIZE, file.size);
					const buf = await file.slice(start, end).arrayBuffer();
					if (settled) return;
					worker.postMessage({ type: 'chunk', index: i, data: buf }, [buf]);
				}
				if (!settled) {
					worker.postMessage({ type: 'done', totalChunks: hashChunks });
				}
			} catch (e) {
				if (settled) return;
				settled = true;
				worker.terminate();
				reject(e instanceof Error ? e : new Error('failed to read file for hashing'));
			}
		})();
	});
}
