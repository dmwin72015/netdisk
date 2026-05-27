<script lang="ts">
	import { AlertDialog, Dialog } from 'bits-ui';
	import { getPending, getInputValue, setInputValue, closeDialog } from '$lib/dialog-state.svelte';
	import * as m from '$lib/paraglide/messages';

	let pending = $derived(getPending());
	let inputVal = $derived(getInputValue());
	let inputEl: HTMLInputElement | undefined = $state();
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
		<Dialog.Root {open} onOpenChange={onOpenChange}>
			<Dialog.Portal>
				<Dialog.Overlay forceMount>
					{#snippet child({ props })}
						<div
							{...props}
							class="fixed inset-0 z-50 bg-black/40 backdrop-blur-sm data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0"
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
							class="fixed left-1/2 top-1/2 z-50 w-full max-w-md -translate-x-1/2 -translate-y-1/2 rounded-xl border border-gray-100 bg-white p-6 shadow-xl data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0 data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95"
						>
							<Dialog.Title class="text-base font-semibold text-gray-900">
								{pending?.opts.title ?? ''}
							</Dialog.Title>
							{#if pending?.opts.message}
								<Dialog.Description class="mt-1.5 text-sm text-gray-500">
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
								class="mt-4 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm outline-none transition-colors focus:border-blue-400 focus:ring-2 focus:ring-blue-50"
							/>
							<div class="mt-5 flex justify-end gap-2">
								<Dialog.Close
									class="rounded-lg px-4 py-2 text-sm text-gray-600 transition-colors hover:bg-gray-100"
								>
									{pending?.opts.cancelText ?? m.cancel()}
								</Dialog.Close>
								<button
									type="button"
									onclick={onConfirm}
									class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-blue-700"
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
		<AlertDialog.Root {open} onOpenChange={onOpenChange}>
			<AlertDialog.Portal>
				<AlertDialog.Overlay forceMount>
					{#snippet child({ props })}
						<div
							{...props}
							class="fixed inset-0 z-50 bg-black/40 backdrop-blur-sm data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0"
						></div>
					{/snippet}
				</AlertDialog.Overlay>
				<AlertDialog.Content forceMount>
					{#snippet child({ props })}
						<div
							{...props}
							bind:this={contentEl}
							class="fixed left-1/2 top-1/2 z-50 w-full max-w-md -translate-x-1/2 -translate-y-1/2 rounded-xl border border-gray-100 bg-white p-6 shadow-xl data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0 data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95"
						>
							<AlertDialog.Title class="text-base font-semibold text-gray-900">
								{pending?.opts.title ?? ''}
							</AlertDialog.Title>
							<AlertDialog.Description class="mt-1.5 text-sm text-gray-500">
								{pending?.opts.message ?? ''}
							</AlertDialog.Description>
							<div class="mt-5 flex justify-end gap-2">
								<AlertDialog.Cancel
									onclick={onCancel}
									class="rounded-lg px-4 py-2 text-sm text-gray-600 transition-colors hover:bg-gray-100"
								>
									{pending?.opts.cancelText ?? m.cancel()}
								</AlertDialog.Cancel>
								<AlertDialog.Action
									onclick={onConfirm}
									class="rounded-lg bg-red-500 px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-red-600"
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
