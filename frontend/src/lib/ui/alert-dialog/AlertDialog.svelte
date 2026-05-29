<script lang="ts">
	import * as AlertDialogBase from './base';

	let {
		open = $bindable(false),
		title,
		description,
		confirmText = 'Confirm',
		cancelText = 'Cancel',
		variant = 'default',
		onConfirm,
		onCancel,
		contentClass = '',
	}: {
		open?: boolean;
		title: string;
		description?: string;
		confirmText?: string;
		cancelText?: string;
		variant?: 'default' | 'destructive';
		onConfirm?: () => void;
		onCancel?: () => void;
		contentClass?: string;
	} = $props();

	let confirmed = $state(false);

	const actionClass = $derived(
		variant === 'destructive'
			? 'inline-flex h-9 items-center justify-center rounded-lg bg-red-600 px-4 text-sm font-medium text-white hover:bg-red-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-red-500 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50'
			: 'inline-flex h-9 items-center justify-center rounded-lg bg-blue-600 px-4 text-sm font-medium text-white hover:bg-blue-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50'
	);

	function handleConfirm() {
		confirmed = true;
		open = false;
	}

	function handleOpenChange(v: boolean) {
		if (!v) {
			if (confirmed) {
				onConfirm?.();
			} else {
				onCancel?.();
			}
			confirmed = false;
		}
	}
</script>

<AlertDialogBase.Root bind:open onOpenChange={handleOpenChange}>
	<AlertDialogBase.Content class={contentClass}>
		<AlertDialogBase.Title>{title}</AlertDialogBase.Title>
		{#if description}
			<AlertDialogBase.Description>{description}</AlertDialogBase.Description>
		{/if}

		<div class="mt-6 flex justify-end gap-3">
			<AlertDialogBase.Cancel>{cancelText}</AlertDialogBase.Cancel>
			<button class={actionClass} onclick={handleConfirm}>
				{confirmText}
			</button>
		</div>
	</AlertDialogBase.Content>
</AlertDialogBase.Root>
