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
	chunkSize?: number,
	opts?: { skipPreHash?: boolean }
): Promise<{ preHash: string; hash: string; totalChunks: number }> {
	if (!file || file.size === undefined) {
		throw new Error('invalid file');
	}
	const totalChunks = Math.ceil(file.size / (chunkSize ?? 4 * 1024 * 1024));
	const fileInfo = `"${file.name}" (${file.size} bytes, ${totalChunks} upload chunks)`;
	console.debug(`[hash] computeSHA256Chunked: starting for ${fileInfo}`);

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
				console.debug(`[hash] pre_hash received: ${preHash.slice(0, 16)}...`);
				callbacks.onPreHash?.(preHash);
			} else if (e.data.type === 'progress') {
				callbacks.onProgress?.(e.data.percent);
			} else if (e.data.type === 'complete') {
				console.debug(`[hash] complete: hash=${e.data.hash.slice(0, 16)}... preHash=${preHash.slice(0, 16)}... totalChunks=${e.data.totalChunks}`);
				settled = true;
				resolve({ preHash, hash: e.data.hash, totalChunks });
				worker.terminate();
			}
		};
		worker.onerror = (ev) => {
			if (settled) return;
			settled = true;
			console.error(`[hash] worker error:`, ev);
			reject(new Error('SHA-256 computation failed'));
			worker.terminate();
		};

		(async () => {
			try {
				if (opts?.skipPreHash) {
					worker.postMessage({ type: 'init', skipPreHash: true });
				}
				const hashChunks = Math.ceil(file.size / HASH_CHUNK_SIZE);
				console.debug(`[hash] sending ${hashChunks} hash chunks to worker`);
				for (let i = 0; i < hashChunks; i++) {
					const start = i * HASH_CHUNK_SIZE;
					const end = Math.min(start + HASH_CHUNK_SIZE, file.size);
					const readStart = performance.now();
					const buf = await file.slice(start, end).arrayBuffer();
					const readTime = performance.now() - readStart;
					if (settled) return;
					worker.postMessage({ type: 'chunk', index: i, data: buf }, [buf]);
					if (i % 10 === 0) {
						console.debug(`[hash] sent chunk ${i + 1}/${hashChunks} (${buf.byteLength} bytes, read in ${readTime.toFixed(0)}ms)`);
					}
				}
				if (!settled) {
					console.debug(`[hash] all ${hashChunks} chunks sent, signaling done`);
					worker.postMessage({ type: 'done', totalChunks: hashChunks });
				}
			} catch (e) {
				if (settled) return;
				settled = true;
				worker.terminate();
				console.error(`[hash] error reading file for hashing:`, e);
				reject(e instanceof Error ? e : new Error('failed to read file for hashing'));
			}
		})();
	});
}
