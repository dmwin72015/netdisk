type DialogType = 'confirm' | 'prompt';

type DialogOptions = {
	title: string;
	message: string;
	confirmText: string;
	cancelText: string;
	inputPlaceholder?: string;
	defaultValue?: string;
	maxLength?: number;
};

type PendingDialog = {
	type: DialogType;
	opts: DialogOptions;
	resolve: (value: boolean | string | null) => void;
};

let pending = $state<PendingDialog | null>(null);
let inputValue = $state('');

export function getPending() {
	return pending;
}

export function getInputValue() {
	return inputValue;
}

export function setInputValue(v: string) {
	inputValue = v;
}

export function closeDialog() {
	pending = null;
	inputValue = '';
}

export function openConfirm(opts: DialogOptions): Promise<boolean> {
	return new Promise<boolean>((resolve) => {
		pending = { type: 'confirm', opts, resolve: (v) => resolve(!!v) };
	});
}

export function openPrompt(opts: DialogOptions): Promise<string | null> {
	inputValue = opts.defaultValue ?? '';
	return new Promise<string | null>((resolve) => {
		pending = { type: 'prompt', opts, resolve: (v) => resolve(typeof v === 'string' ? v : null) };
	});
}
