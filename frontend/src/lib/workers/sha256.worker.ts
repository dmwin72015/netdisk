const CHUNK_SIZE = 8 * 1024 * 1024; // 8MB

self.onmessage = async (e: MessageEvent<File>) => {
	const file = e.data;
	const totalChunks = Math.ceil(file.size / CHUNK_SIZE);

	// Use Web Crypto API for streaming hash
	const hashBuffer = await crypto.subtle.digest('SHA-256', await file.arrayBuffer());
	const hashArray = Array.from(new Uint8Array(hashBuffer));
	const hash = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');

	self.postMessage({ type: 'complete', hash, totalChunks });
};
