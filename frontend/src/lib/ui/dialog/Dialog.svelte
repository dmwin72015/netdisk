<script lang="ts">
	import type { Snippet } from 'svelte';
	import * as DialogBase from './base';

	let {
		open = $bindable(false),
		title,
		description,
		confirmText = 'Confirm',
		cancelText = 'Cancel',
		onConfirm,
		onCancel,
		children,
		class: className = '',
	}: {
		open?: boolean;
		title?: string;
		description?: string;
		confirmText?: string;
		cancelText?: string;
		onConfirm?: () => void;
		onCancel?: () => void;
		children?: Snippet;
		class?: string;
	} = $props();

	function handleConfirm() {
		onConfirm?.();
		open = false;
	}

	function handleCancel() {
		onCancel?.();
		open = false;
	}
</script>

<DialogBase.Root bind:open>
	<DialogBase.Content class={className}>
		<DialogBase.Header>
			{#if title}
				<DialogBase.Title>{title}</DialogBase.Title>
			{/if}
			{#if description}
				<DialogBase.Description>{description}</DialogBase.Description>
			{/if}
		</DialogBase.Header>

		{#if children}
			<div class="py-4">
				{@render children()}
			</div>
		{/if}

		<DialogBase.Footer>
			<DialogBase.Close>
				<button
					type="button"
					class="inline-flex h-9 items-center justify-center rounded-lg border border-gray-200 bg-white px-4 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2"
					onclick={handleCancel}
				>
					{cancelText}
				</button>
			</DialogBase.Close>
			<button
				type="button"
				class="inline-flex h-9 items-center justify-center rounded-lg bg-blue-600 px-4 text-sm font-medium text-white transition-colors hover:bg-blue-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2"
				onclick={handleConfirm}
			>
				{confirmText}
			</button>
		</DialogBase.Footer>
	</DialogBase.Content>
</DialogBase.Root>
