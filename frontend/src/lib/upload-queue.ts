export function createUploadQueue<T extends { uid: string; phase: string }>(
	getItems: () => T[],
	setItems: (items: T[]) => void,
	isProcessable: (item: T) => boolean,
	processItem: (item: T) => Promise<void>,
	onAllDone?: () => void,
) {
	let running = false;

	async function start() {
		if (running) return;
		running = true;

		try {
			while (true) {
				const items = getItems();
				const idx = items.findIndex(isProcessable);
				if (idx === -1) break;

				const item = items[idx];
				try {
					await processItem(item);
				} catch {
					// processItem is responsible for setting error state
				}
			}
		} finally {
			running = false;
		}

		onAllDone?.();
	}

	return { start };
}
