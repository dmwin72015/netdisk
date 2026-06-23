import { describe, expect, it, vi } from 'vitest';

import { createAlertDialogController } from '$lib/ui/alert-dialog/AlertDialog.svelte';

describe('AlertDialog', () => {
	it('runs the confirm callback when the confirm button is clicked', () => {
		const onConfirm = vi.fn();
		const onCancel = vi.fn();
		const controller = createAlertDialogController(onConfirm, onCancel);

		controller.handleConfirm();

		expect(onConfirm).toHaveBeenCalledTimes(1);
		expect(onCancel).not.toHaveBeenCalled();
	});
});
