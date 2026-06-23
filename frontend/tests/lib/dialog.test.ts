import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));
vi.mock('$lib/paraglide/messages', () => ({
	confirm_delete: () => 'Confirm Delete',
	delete_btn: () => 'Delete',
	cancel: () => 'Cancel',
	confirm: () => 'Confirm',
}));

const openConfirmMock = vi.fn();
const openPromptMock = vi.fn();

vi.mock('$lib/dialog-state.svelte', () => ({
	openConfirm: (...args: unknown[]) => openConfirmMock(...args),
	openPrompt: (...args: unknown[]) => openPromptMock(...args),
}));

import { confirmDelete, promptInput, confirmAction } from '$lib/dialog';

beforeEach(() => {
	openConfirmMock.mockReset();
	openPromptMock.mockReset();
});

// ── confirmDelete ──────────────────────────────────────────────────

describe('confirmDelete', () => {
	it('calls openConfirm with correct options', async () => {
		openConfirmMock.mockResolvedValue(true);

		const result = await confirmDelete('Are you sure?');

		expect(openConfirmMock).toHaveBeenCalledWith({
			title: 'Confirm Delete',
			message: 'Are you sure?',
			confirmText: 'Delete',
			cancelText: 'Cancel',
		});
		expect(result).toBe(true);
	});

	it('returns false when user cancels', async () => {
		openConfirmMock.mockResolvedValue(false);

		const result = await confirmDelete('Delete this?');
		expect(result).toBe(false);
	});
});

// ── promptInput ────────────────────────────────────────────────────

describe('promptInput', () => {
	it('calls openPrompt with correct options', async () => {
		openPromptMock.mockResolvedValue('user input');

		const result = await promptInput('Rename', 'Enter name');

		expect(openPromptMock).toHaveBeenCalledWith({
			title: 'Rename',
			message: '',
			confirmText: 'Confirm',
			cancelText: 'Cancel',
			inputPlaceholder: 'Enter name',
			defaultValue: undefined,
			maxLength: undefined,
		});
		expect(result).toBe('user input');
	});

	it('passes defaultValue and maxLength', async () => {
		openPromptMock.mockResolvedValue('old');

		await promptInput('Rename', 'Enter name', 'old', 100);

		expect(openPromptMock).toHaveBeenCalledWith({
			title: 'Rename',
			message: '',
			confirmText: 'Confirm',
			cancelText: 'Cancel',
			inputPlaceholder: 'Enter name',
			defaultValue: 'old',
			maxLength: 100,
		});
	});

	it('returns null when user cancels', async () => {
		openPromptMock.mockResolvedValue(null);

		const result = await promptInput('Rename', 'Enter name');
		expect(result).toBeNull();
	});
});

// ── confirmAction ──────────────────────────────────────────────────

describe('confirmAction', () => {
	it('calls openConfirm with correct options', async () => {
		openConfirmMock.mockResolvedValue(true);

		const result = await confirmAction('Overwrite', 'File exists', 'Overwrite');

		expect(openConfirmMock).toHaveBeenCalledWith({
			title: 'Overwrite',
			message: 'File exists',
			confirmText: 'Overwrite',
			cancelText: 'Cancel',
		});
		expect(result).toBe(true);
	});

	it('returns false when user cancels', async () => {
		openConfirmMock.mockResolvedValue(false);

		const result = await confirmAction('Overwrite', 'File exists', 'Overwrite');
		expect(result).toBe(false);
	});
});
