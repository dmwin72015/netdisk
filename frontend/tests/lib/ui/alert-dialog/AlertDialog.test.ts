import { describe, expect, it, vi } from 'vitest';

import { createAlertDialogController } from '$lib/ui/alert-dialog/AlertDialog.svelte';

// ── Controller unit tests ──────────────────────────────────────────
// Note: Full component rendering tests require a client-side Svelte 5
// runtime. The controller logic below provides behavioral coverage for
// the module-level createAlertDialogController function.

describe('createAlertDialogController', () => {
	it('calls onConfirm and close on handleConfirm', () => {
		const onConfirm = vi.fn();
		const onCancel = vi.fn();
		const close = vi.fn();

		const ctrl = createAlertDialogController(onConfirm, onCancel, close);
		ctrl.handleConfirm();

		expect(onConfirm).toHaveBeenCalledTimes(1);
		expect(close).toHaveBeenCalledTimes(1);
		expect(onCancel).not.toHaveBeenCalled();
	});

	it('calls onCancel when handleOpenChange(false) before confirm', () => {
		const onConfirm = vi.fn();
		const onCancel = vi.fn();
		const close = vi.fn();

		const ctrl = createAlertDialogController(onConfirm, onCancel, close);
		ctrl.handleOpenChange(false);

		expect(onCancel).toHaveBeenCalledTimes(1);
		expect(onConfirm).not.toHaveBeenCalled();
		expect(close).not.toHaveBeenCalled();
	});

	it('does NOT call onCancel when handleOpenChange(false) after confirm', () => {
		const onConfirm = vi.fn();
		const onCancel = vi.fn();
		const close = vi.fn();

		const ctrl = createAlertDialogController(onConfirm, onCancel, close);
		ctrl.handleConfirm();
		ctrl.handleOpenChange(false);

		expect(onConfirm).toHaveBeenCalledTimes(1);
		expect(onCancel).not.toHaveBeenCalled();
		expect(close).toHaveBeenCalledTimes(1);
	});

	it('resets confirmed flag after handleOpenChange(false)', () => {
		const onConfirm = vi.fn();
		const onCancel = vi.fn();
		const close = vi.fn();

		const ctrl = createAlertDialogController(onConfirm, onCancel, close);
		ctrl.handleConfirm();
		ctrl.handleOpenChange(false);
		// After reset, confirmed=false, so closing again should call onCancel
		ctrl.handleOpenChange(false);

		expect(onConfirm).toHaveBeenCalledTimes(1);
		expect(onCancel).toHaveBeenCalledTimes(1);
		expect(close).toHaveBeenCalledTimes(1);
	});

	it('does nothing on handleOpenChange(true)', () => {
		const onConfirm = vi.fn();
		const onCancel = vi.fn();
		const close = vi.fn();

		const ctrl = createAlertDialogController(onConfirm, onCancel, close);
		ctrl.handleOpenChange(true);

		expect(onConfirm).not.toHaveBeenCalled();
		expect(onCancel).not.toHaveBeenCalled();
		expect(close).not.toHaveBeenCalled();
	});

	it('handles undefined callbacks gracefully', () => {
		const ctrl = createAlertDialogController();
		// Should not throw when all callbacks are undefined
		ctrl.handleConfirm();
		ctrl.handleOpenChange(false);
		ctrl.handleOpenChange(true);
	});
});
