<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils/cn';

	type Side = 'left' | 'right' | 'top' | 'bottom';

	let {
		children,
		class: className = '',
		side = 'right',
		...restProps
	}: {
		children: Snippet;
		class?: string;
		side?: Side;
	} = $props();

	const sideClasses: Record<Side, string> = {
		left:
			'left-0 top-0 h-full w-3/4 max-w-sm border-r ' +
			'data-[state=open]:slide-in-from-left data-[state=closed]:slide-out-to-left',
		right:
			'right-0 top-0 h-full w-3/4 max-w-sm border-l ' +
			'data-[state=open]:slide-in-from-right data-[state=closed]:slide-out-to-right',
		top:
			'top-0 left-0 w-full h-auto max-h-[80vh] border-b ' +
			'data-[state=open]:slide-in-from-top data-[state=closed]:slide-out-to-top',
		bottom:
			'bottom-0 left-0 w-full h-auto max-h-[80vh] border-t ' +
			'data-[state=open]:slide-in-from-bottom data-[state=closed]:slide-out-to-bottom',
	};

	const panelClass = $derived(sideClasses[side]);
</script>

<Dialog.Portal>
	<Dialog.Content
		class={cn(
			'fixed z-50 bg-surface border-line shadow-pop outline-none rounded-lg',
			'data-[state=open]:animate-in data-[state=open]:duration-300',
			'data-[state=closed]:animate-out data-[state=closed]:duration-200',
			panelClass,
			className,
		)}
		{...restProps}
	>
		{@render children()}
	</Dialog.Content>
</Dialog.Portal>
