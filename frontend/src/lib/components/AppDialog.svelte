<script lang="ts">
	import { AlertDialog, Dialog } from 'bits-ui';
	import { getPending, getInputValue, setInputValue, getCheckboxValue, setCheckboxValue, closeDialog } from '$lib/dialog-state.svelte';
	import { fmtSize } from '$lib/utils/format';
	import * as m from '$lib/paraglide/messages';

	let pending = $derived(getPending());
	let inputVal = $derived(getInputValue());
	let inputEl: HTMLInputElement | undefined = $state();
	let sizeHint = $derived.by(() => {
		const n = parseInt(inputVal, 10);
		if (!isNaN(n) && n > 0) return fmtSize(n);
		return '';
	});
	let open = $state(false);
	let closing = $state(false);
	let contentEl: HTMLElement | undefined = $state();
	let resolveValue: boolean | string | null | undefined = undefined;

	$effect(() => {
		if (pending && !closing) {
			requestAnimationFrame(() => {
				open = true;
			});
		}
	});

	function cleanupAfterAnimation() {
		const p = pending;
		open = false;
		closing = false;
		closeDialog();
		if (resolveValue !== undefined) {
			p?.resolve(resolveValue);
			resolveValue = undefined;
		}
	}

	function closeWithAnimation(value: boolean | string | null | undefined) {
		if (closing) return;
		closing = true;
		resolveValue = value;
		open = false;

		const el = contentEl;
		if (el) {
			el.addEventListener('animationend', () => cleanupAfterAnimation(), { once: true });
		}
		setTimeout(() => {
			if (closing) cleanupAfterAnimation();
		}, 300);
	}

	function onOpenChange(isOpen: boolean) {
		if (!isOpen && !closing) {
			closeWithAnimation(undefined);
		}
	}

	function onConfirm() {
		if (pending?.type === 'prompt') {
			const val = inputVal.trim();
			if (!val) return;
			closeWithAnimation(val);
		} else {
			closeWithAnimation(true);
		}
	}

	function onCancel() {
		closeWithAnimation(false);
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && inputEl === document.activeElement) {
			e.preventDefault();
			onConfirm();
		}
	}

	let isPrompt = $derived(pending?.type === 'prompt');
</script>

{#if pending}
	{#if isPrompt}
		<Dialog.Root {open} {onOpenChange}>
			<Dialog.Portal>
				<Dialog.Overlay forceMount>
					{#snippet child({ props })}
						<div
							{...props}
							class="fixed inset-0 z-50 bg-overlay backdrop-blur-sm data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0 duration-200"
						></div>
					{/snippet}
				</Dialog.Overlay>
				<Dialog.Content
					forceMount
					onInteractOutside={(e) => e.preventDefault()}
				>
					{#snippet child({ props })}
						<div
							{...props}
							bind:this={contentEl}
							class="fixed left-1/2 top-1/2 z-50 w-full max-w-md -translate-x-1/2 -translate-y-1/2 rounded-xl border border-line-soft bg-white p-6 shadow-dialog data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0 data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95 duration-200"
						>
							<Dialog.Title class="text-sm font-medium leading-5 text-ink-2">
								{pending?.opts.title ?? ''}
							</Dialog.Title>
							{#if pending?.opts.message}
								<Dialog.Description class="mt-1 text-xs text-ink-3">
									{pending.opts.message}
								</Dialog.Description>
							{/if}
							<input
								bind:this={inputEl}
								type="text"
								value={inputVal}
								oninput={(e) => setInputValue(e.currentTarget.value)}
								onkeydown={handleKeydown}
								placeholder={pending?.opts.inputPlaceholder ?? ''}
								maxlength={pending?.opts.maxLength}
								class="mt-4 w-full rounded-lg border border-line px-3 py-2 text-sm outline-none transition-colors focus:border-primary focus:ring-2 focus:ring-primary/20"
							/>
							<div class="mt-5 flex justify-end gap-2">
								<Dialog.Close
									class="rounded-lg px-4 py-2 text-sm text-ink-3 transition-colors hover:bg-surface-sunken"
								>
									{pending?.opts.cancelText ?? m.cancel()}
								</Dialog.Close>
								<button
									type="button"
									onclick={onConfirm}
									class="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
								>
									{pending?.opts.confirmText ?? m.confirm()}
								</button>
							</div>
						</div>
					{/snippet}
				</Dialog.Content>
			</Dialog.Portal>
		</Dialog.Root>
	{:else}
		<AlertDialog.Root {open} {onOpenChange}>
			<AlertDialog.Portal>
				<AlertDialog.Overlay forceMount>
					{#snippet child({ props })}
						<div
							{...props}
							class="fixed inset-0 z-50 bg-overlay backdrop-blur-sm data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0 duration-200"
						></div>
					{/snippet}
				</AlertDialog.Overlay>
				<AlertDialog.Content forceMount>
					{#snippet child({ props })}
						<div
							{...props}
							bind:this={contentEl}
							class="fixed left-1/2 top-1/2 z-50 w-full max-w-md -translate-x-1/2 -translate-y-1/2 rounded-xl border border-line-soft bg-white p-6 shadow-dialog data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0 data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95 duration-200"
						>
							<AlertDialog.Title class="text-sm font-medium leading-5 text-ink-2">
								{pending?.opts.title ?? ''}
							</AlertDialog.Title>
							<AlertDialog.Description class="mt-1 text-xs text-ink-3">
								{pending?.opts.message ?? ''}
							</AlertDialog.Description>
							{#if pending?.opts.checkboxLabel}
								<label class="mt-4 flex cursor-pointer items-center gap-2">
									<input
										type="checkbox"
										checked={getCheckboxValue()}
										onchange={(e) => setCheckboxValue((e.currentTarget as HTMLInputElement).checked)}
										class="h-4 w-4 rounded border-line text-primary focus:ring-primary"
									/>
									<span class="text-xs text-ink-3">{pending.opts.checkboxLabel}</span>
								</label>
							{/if}
								{#if sizeHint}
									<p class="mt-1.5 text-xs text-ink-4">≈ {sizeHint}</p>
								{/if}
							<div class="mt-5 flex justify-end gap-2">
								<AlertDialog.Cancel
									onclick={onCancel}
									class="rounded-lg px-4 py-2 text-sm text-ink-3 transition-colors hover:bg-surface-sunken"
								>
									{pending?.opts.cancelText ?? m.cancel()}
								</AlertDialog.Cancel>
								<AlertDialog.Action
									onclick={onConfirm}
									class="rounded-lg bg-danger px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-danger"
								>
									{pending?.opts.confirmText ?? m.confirm()}
								</AlertDialog.Action>
							</div>
						</div>
					{/snippet}
				</AlertDialog.Content>
			</AlertDialog.Portal>
		</AlertDialog.Root>
	{/if}
{/if}
