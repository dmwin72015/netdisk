// Dual SHA-256 worker: computes pre_hash (first 512KB) and full file hash
// Messages: { type: 'chunk', index: number, data: ArrayBuffer } (transferred)
//           { type: 'done', totalChunks: number, preHashChunks: number }
// Responses: { type: 'pre_hash', hash: string }
//            { type: 'progress', percent: number }
//            { type: 'complete', hash: string, totalChunks: number }

const K = new Uint32Array([
	0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
	0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
	0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
	0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
	0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
	0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
	0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
	0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2
]);

function rotr(x: number, n: number): number {
	return (x >>> n) | (x << (32 - n));
}

const BLOCK_SIZE = 64;
const w = new Uint32Array(64);

function sha256Block(state: Uint32Array, block: DataView, offset: number): void {
	for (let i = 0; i < 16; i++) w[i] = block.getUint32(offset + i * 4, false);
	for (let i = 16; i < 64; i++) {
		w[i] = (rotr(w[i - 2], 17) ^ rotr(w[i - 2], 19) ^ (w[i - 2] >>> 10)) + w[i - 7] +
			(rotr(w[i - 15], 7) ^ rotr(w[i - 15], 18) ^ (w[i - 15] >>> 3)) + w[i - 16];
	}

	let a = state[0], b = state[1], c = state[2], d = state[3];
	let e = state[4], f = state[5], g = state[6], h = state[7];

	for (let i = 0; i < 64; i++) {
		const S1 = rotr(e, 6) ^ rotr(e, 11) ^ rotr(e, 25);
		const ch = (e & f) ^ (~e & g);
		const temp1 = (h + S1 + ch + K[i] + w[i]) | 0;
		const S0 = rotr(a, 2) ^ rotr(a, 13) ^ rotr(a, 22);
		const maj = (a & b) ^ (a & c) ^ (b & c);
		const temp2 = (S0 + maj) | 0;

		h = g; g = f; f = e; e = (d + temp1) | 0;
		d = c; c = b; b = a; a = (temp1 + temp2) | 0;
	}

	state[0] = (state[0] + a) | 0;
	state[1] = (state[1] + b) | 0;
	state[2] = (state[2] + c) | 0;
	state[3] = (state[3] + d) | 0;
	state[4] = (state[4] + e) | 0;
	state[5] = (state[5] + f) | 0;
	state[6] = (state[6] + g) | 0;
	state[7] = (state[7] + h) | 0;
}

function createState(): Uint32Array {
	return new Uint32Array([0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a, 0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19]);
}

function processBlocks(state: Uint32Array, data: Uint8Array): void {
	const view = new DataView(data.buffer, data.byteOffset, data.byteLength);
	for (let i = 0; i + BLOCK_SIZE <= data.length; i += BLOCK_SIZE) {
		sha256Block(state, view, i);
	}
}

interface HashContext {
	state: Uint32Array;
	buffer: Uint8Array;
	totalBytes: number;
}

function createHashContext(): HashContext {
	return { state: createState(), buffer: new Uint8Array(0), totalBytes: 0 };
}

function updateHash(ctx: HashContext, data: Uint8Array): void {
	ctx.totalBytes += data.length;

	// Prepend leftover buffer, then process complete 64-byte blocks in-place
	const prevLen = ctx.buffer.length;
	const totalLen = prevLen + data.length;

	// Process complete blocks: first from leftover + beginning of new data
	if (totalLen >= BLOCK_SIZE) {
		// Build a view over leftover + new data without copying where possible
		const firstBlockEnd = BLOCK_SIZE - prevLen;
		if (prevLen > 0) {
			// Need to combine leftover with start of new data for first block(s)
			const combined = new Uint8Array(totalLen);
			combined.set(ctx.buffer);
			combined.set(data, prevLen);
			const completeLen = Math.floor(totalLen / BLOCK_SIZE) * BLOCK_SIZE;
			processBlocks(ctx.state, combined.subarray(0, completeLen));
			ctx.buffer = combined.subarray(completeLen);
		} else {
			// No leftover, process new data directly
			const completeLen = Math.floor(data.length / BLOCK_SIZE) * BLOCK_SIZE;
			if (completeLen > 0) {
				processBlocks(ctx.state, data.subarray(0, completeLen));
			}
			ctx.buffer = data.slice(completeLen);
		}
	} else {
		// Not enough data for a block, just buffer it
		const combined = new Uint8Array(totalLen);
		combined.set(ctx.buffer);
		combined.set(data, prevLen);
		ctx.buffer = combined;
	}
}

function finalizeHash(ctx: HashContext): string {
	const bitLen = ctx.totalBytes * 8;
	const padLen = (BLOCK_SIZE - ((ctx.buffer.length + 1 + 8) % BLOCK_SIZE)) % BLOCK_SIZE;
	const padded = new Uint8Array(ctx.buffer.length + 1 + 8 + padLen);
	padded.set(ctx.buffer);
	padded[ctx.buffer.length] = 0x80;

	const view = new DataView(padded.buffer);
	view.setUint32(padded.length - 8, Math.floor(bitLen / 0x100000000), false);
	view.setUint32(padded.length - 4, bitLen >>> 0, false);

	processBlocks(ctx.state, padded);

	return Array.from(ctx.state).map(v => v.toString(16).padStart(8, '0')).join('');
}

// --- Main ---

const PRE_HASH_SIZE = 512 * 1024; // 512KB

let fullCtx = createHashContext();
let preCtx = createHashContext();
let preHashSent = false;
let preHashTargetBytes = PRE_HASH_SIZE;
let preHashAccumBytes = 0;
let totalChunks = 0;
let chunksReceived = 0;

self.onmessage = (e: MessageEvent) => {
	if (e.data.type === 'chunk') {
		const data = new Uint8Array(e.data.data);
		chunksReceived++;

		// Always feed full hash
		updateHash(fullCtx, data);

		// Feed pre hash until we have enough bytes
		if (!preHashSent) {
			const remaining = preHashTargetBytes - preHashAccumBytes;
			if (remaining > 0) {
				const slice = data.subarray(0, Math.min(data.length, remaining));
				updateHash(preCtx, slice);
				preHashAccumBytes += slice.length;
			}
			if (preHashAccumBytes >= preHashTargetBytes) {
				const preHash = finalizeHash(preCtx);
				preHashSent = true;
				self.postMessage({ type: 'pre_hash', hash: preHash });
			}
		}

		// Report progress
		if (totalChunks > 0) {
			self.postMessage({ type: 'progress', percent: Math.round((chunksReceived / totalChunks) * 100) });
		}

	} else if (e.data.type === 'done') {
		totalChunks = e.data.totalChunks;

		// If pre_hash wasn't sent yet (file < 512KB), send it now
		if (!preHashSent) {
			const preHash = finalizeHash(preCtx);
			preHashSent = true;
			self.postMessage({ type: 'pre_hash', hash: preHash });
		}

		const hash = finalizeHash(fullCtx);
		self.postMessage({ type: 'complete', hash, totalChunks });
	}
};
