import { openConfirm, openPrompt } from './dialog-state.svelte';
import * as m from '$lib/paraglide/messages';

export async function confirmDelete(message: string): Promise<boolean> {
	return openConfirm({
		title: m.confirm_delete(),
		message,
		confirmText: m.delete_btn(),
		cancelText: m.cancel()
	});
}

export async function promptInput(title: string, placeholder: string, defaultValue?: string, maxLength?: number): Promise<string | null> {
	return openPrompt({
		title,
		message: '',
		confirmText: m.confirm(),
		cancelText: m.cancel(),
		inputPlaceholder: placeholder,
		defaultValue,
		maxLength
	});
}

export async function confirmAction(title: string, text: string, confirmText: string): Promise<boolean> {
	return openConfirm({
		title,
		message: text,
		confirmText,
		cancelText: m.cancel()
	});
}
