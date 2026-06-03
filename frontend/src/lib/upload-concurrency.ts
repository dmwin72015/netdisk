export const UPLOAD_FILE_CONCURRENCY = 3;
export const UPLOAD_CHUNK_CONCURRENCY_PER_FILE = 4;
export const UPLOAD_REQUEST_POOL_SIZE = UPLOAD_FILE_CONCURRENCY * UPLOAD_CHUNK_CONCURRENCY_PER_FILE;

type QueueItem<T> = {
	run: () => Promise<T>;
	resolve: (value: T) => void;
	reject: (reason?: unknown) => void;
	signal?: AbortSignal;
	onAbort: () => void;
};

export class UploadRequestPool {
	private active = 0;
	private queue: QueueItem<unknown>[] = [];
	private logInterval: ReturnType<typeof setInterval> | null = null;

	constructor(private readonly limit: number) {
		if (typeof setInterval !== 'undefined') {
			this.logInterval = setInterval(() => this.logStats(), 10_000);
		}
	}

	private logStats() {
		if (this.active > 0 || this.queue.length > 0) {
			console.debug(`[request-pool] active=${this.active} queued=${this.queue.length} limit=${this.limit}`);
		}
	}

	schedule<T>(run: () => Promise<T>, signal?: AbortSignal): Promise<T> {
		if (this.limit <= 0) return run();
		if (signal?.aborted) return Promise.reject(signal.reason ?? new DOMException('Aborted', 'AbortError'));

		return new Promise<T>((resolve, reject) => {
			let item: QueueItem<T>;
			const onAbort = () => {
				console.debug(`[request-pool] abort: removing queued item (active=${this.active} queued=${this.queue.length})`);
				this.queue = this.queue.filter((queued) => queued !== item);
				reject(signal?.reason ?? new DOMException('Aborted', 'AbortError'));
			};
			item = { run, resolve, reject, signal, onAbort };
			signal?.addEventListener('abort', onAbort, { once: true });
			this.queue.push(item as QueueItem<unknown>);
			console.debug(`[request-pool] schedule: queued (active=${this.active} queued=${this.queue.length} limit=${this.limit})`);
			this.pump();
		});
	}

	private pump() {
		while (this.active < this.limit) {
			const item = this.queue.shift();
			if (!item) return;
			if (item.signal?.aborted) {
				item.signal.removeEventListener('abort', item.onAbort);
				item.reject(item.signal.reason ?? new DOMException('Aborted', 'AbortError'));
				continue;
			}

			this.active++;
			console.debug(`[request-pool] pump: starting (active=${this.active} queued=${this.queue.length})`);
			item.signal?.removeEventListener('abort', item.onAbort);
			item.run()
				.then(item.resolve, item.reject)
				.finally(() => {
					this.active--;
					console.debug(`[request-pool] done: completed (active=${this.active} queued=${this.queue.length})`);
					this.pump();
				});
		}
	}
}

export const uploadRequestPool = new UploadRequestPool(UPLOAD_REQUEST_POOL_SIZE);
