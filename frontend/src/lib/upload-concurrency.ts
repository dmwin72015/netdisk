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

	constructor(private readonly limit: number) {}

	schedule<T>(run: () => Promise<T>, signal?: AbortSignal): Promise<T> {
		if (this.limit <= 0) return run();
		if (signal?.aborted) return Promise.reject(signal.reason ?? new DOMException('Aborted', 'AbortError'));

		return new Promise<T>((resolve, reject) => {
			let item: QueueItem<T>;
			const onAbort = () => {
				this.queue = this.queue.filter((queued) => queued !== item);
				reject(signal?.reason ?? new DOMException('Aborted', 'AbortError'));
			};
			item = { run, resolve, reject, signal, onAbort };
			signal?.addEventListener('abort', onAbort, { once: true });
			this.queue.push(item as QueueItem<unknown>);
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
			item.signal?.removeEventListener('abort', item.onAbort);
			item.run()
				.then(item.resolve, item.reject)
				.finally(() => {
					this.active--;
					this.pump();
				});
		}
	}
}

export const uploadRequestPool = new UploadRequestPool(UPLOAD_REQUEST_POOL_SIZE);
