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

	it('calls onCancel when dialog closes without confirming', () => {
		const onConfirm = vi.fn();
		const onCancel = vi.fn();
		const controller = createAlertDialogController(onConfirm, onCancel);

		controller.handleOpenChange(false);

		expect(onCancel).toHaveBeenCalledTimes(1);
		expect(onConfirm).not.toHaveBeenCalled();
	});

	it('does not call onCancel when dialog opens', () => {
		const onConfirm = vi.fn();
		const onCancel = vi.fn();
		const controller = createAlertDialogController(onConfirm, onCancel);

		controller.handleOpenChange(true);

		expect(onCancel).not.toHaveBeenCalled();
		expect(onConfirm).not.toHaveBeenCalled();
	});

	it('does not call onCancel when closing after confirming', () => {
		const onConfirm = vi.fn();
		const onCancel = vi.fn();
		const controller = createAlertDialogController(onConfirm, onCancel);

		controller.handleConfirm();
		controller.handleOpenChange(false);

		expect(onConfirm).toHaveBeenCalledTimes(1);
		expect(onCancel).not.toHaveBeenCalled();
	});

	it('resets confirmed flag after cancel so second close also triggers onCancel', () => {
		const onConfirm = vi.fn();
		const onCancel = vi.fn();
		const controller = createAlertDialogController(onConfirm, onCancel);

		controller.handleOpenChange(false); // cancel
		expect(onCancel).toHaveBeenCalledTimes(1);

		controller.handleOpenChange(false); // cancel again
		expect(onCancel).toHaveBeenCalledTimes(2);
	});
});
