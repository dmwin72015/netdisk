import { browser } from "$app/environment";

type PinOptions = {
	title: string;
	message?: string;
	confirmText?: string;
	cancelText?: string;
	placeholder?: string;
};

type PendingPin = {
	opts: PinOptions;
	resolve: (value: string | null) => void;
};

let pending = $state<PendingPin | null>(null);
let pinValue = $state("");

export function getPending() {
	return pending;
}

export function getPinValue() {
	return pinValue;
}

export function setPinValue(v: string) {
	pinValue = v;
}

export function closePinDialog() {
	pending = null;
	pinValue = "";
}

export function openPin(opts: PinOptions): Promise<string | null> {
	pinValue = "";
	return new Promise<string | null>((resolve) => {
		pending = { opts, resolve: (v) => resolve(v) };
	});
}
