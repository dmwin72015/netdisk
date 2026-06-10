export type PasteFileFilterResult = {
	accepted: File[];
	rejected: File[];
};

export function extractClipboardFiles(clipboardData: DataTransfer | null | undefined): File[] {
	if (!clipboardData) return [];

	const fromFiles = Array.from(clipboardData.files ?? []);
	if (fromFiles.length > 0) return fromFiles;

	return Array.from(clipboardData.items ?? [])
		.filter((item) => item.kind === 'file')
		.map((item) => item.getAsFile())
		.filter((candidate): candidate is File => candidate instanceof File);
}

export function filterPasteFiles(
	files: File[],
	acceptFile?: (file: File) => boolean,
): PasteFileFilterResult {
	if (!acceptFile) return { accepted: files, rejected: [] };

	const accepted: File[] = [];
	const rejected: File[] = [];
	for (const file of files) {
		if (acceptFile(file)) accepted.push(file);
		else rejected.push(file);
	}
	return { accepted, rejected };
}

export function isEditablePasteTarget(target: EventTarget | null): boolean {
	if (!target || typeof target !== 'object') return false;

	const element = target as {
		isContentEditable?: boolean;
		tagName?: string;
	};
	if (element.isContentEditable) return true;

	const tagName = element.tagName?.toLowerCase();
	return tagName === 'input' || tagName === 'textarea' || tagName === 'select';
}
