<script lang="ts">
	import type { Snippet } from 'svelte';
	import { X } from '@lucide/svelte';
	import * as DialogBase from './base';

	let {
		open = $bindable(false),
		title,
		description,
		okText,
		cancelText,
		confirmText,
		onOpenChange,
		onOpenChangeComplete,
		onOk,
		onCancel,
		onConfirm,
		children,
		headerExtra,
		footer = true,
		showFooter,
		showCancel,
		showConfirm,
		headerClass = '',
		contentStyle = '',
		titleClass = '',
		descriptionClass = '',
		bodyClass = '',
		closeButtonClass = '',
		closeIconSize = 18,
		closable = true,
		class: className = '',
	}: {
		open?: boolean;
		title?: string;
		description?: string;
		okText?: string;
		cancelText?: string;
		confirmText?: string;
		onOpenChange?: (open: boolean) => void;
		onOpenChangeComplete?: (open: boolean) => void;
		onOk?: () => void;
		onCancel?: () => void;
		onConfirm?: () => void;
		children?: Snippet;
		headerExtra?: Snippet;
		footer?: boolean;
		showFooter?: boolean;
		showCancel?: boolean;
		showConfirm?: boolean;
		headerClass?: string;
		contentStyle?: string;
		titleClass?: string;
		descriptionClass?: string;
		bodyClass?: string;
		closeButtonClass?: string;
		closeIconSize?: number;
		closable?: boolean;
		class?: string;
	} = $props();

	const shouldShowFooter = $derived(showFooter ?? footer);
	const shouldShowCancel = $derived(showCancel ?? true);
	const shouldShowOk = $derived(showConfirm ?? true);
	const resolvedOkText = $derived(okText ?? confirmText ?? 'Confirm');
	const resolvedCancelText = $derived(cancelText ?? 'Cancel');

	function handleConfirm() {
		onOk?.();
		onConfirm?.();
		open = false;
	}

	function handleCancel() {
		onCancel?.();
		open = false;
	}
</script>

<DialogBase.Root bind:open {onOpenChange} {onOpenChangeComplete}>
	<DialogBase.Content class="max-h-[90vh] p-0! flex flex-col overflow-hidden {className}" style={contentStyle}>
		<div class="flex {description ? 'items-start' : 'items-center'} gap-3 border-b border-gray-100 px-5 py-3 {headerClass}">
			<DialogBase.Header class="min-w-0 flex-1 space-y-1">
				{#if title}
					<DialogBase.Title class="truncate text-sm font-medium leading-5 text-gray-800 {titleClass}">{title}</DialogBase.Title>
				{/if}
				{#if description}
					<DialogBase.Description class="text-xs text-gray-400 {descriptionClass}">{description}</DialogBase.Description>
				{/if}
			</DialogBase.Header>
			{#if headerExtra}
				<div class="flex shrink-0 items-center gap-1">
					{@render headerExtra()}
				</div>
			{/if}
			{#if closable}
				<DialogBase.Close>
					<button
						type="button"
						onclick={handleCancel}
						class="rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 {closeButtonClass}"
						aria-label="Close"
					>
						<X size={closeIconSize} />
					</button>
				</DialogBase.Close>
			{/if}
		</div>

		<div class="flex-1 overflow-auto px-5 py-4 {bodyClass}">
			{#if children}
				{@render children()}
			{/if}
		</div>

		{#if shouldShowFooter && (shouldShowCancel || shouldShowOk)}
			<DialogBase.Footer class="border-t border-gray-100 px-5 py-3">
				{#if shouldShowCancel}
					<DialogBase.Close>
						<button
							type="button"
							class="inline-flex h-8 items-center justify-center rounded-lg px-3.5 text-sm text-gray-600 transition-colors hover:bg-gray-100"
							onclick={handleCancel}
						>
							{resolvedCancelText}
						</button>
					</DialogBase.Close>
				{/if}
				{#if shouldShowOk}
					<button
						type="button"
						class="inline-flex h-8 items-center justify-center rounded-lg bg-blue-600 px-3.5 text-sm font-medium text-white transition-colors hover:bg-blue-700"
						onclick={handleConfirm}
					>
						{resolvedOkText}
					</button>
				{/if}
			</DialogBase.Footer>
		{/if}
	</DialogBase.Content>
</DialogBase.Root>
