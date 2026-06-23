<script lang="ts">
	import { AlertDialog } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils/cn';

	let {
		children,
		class: className = '',
		disabled = false,
		...restProps
	}: {
		children: Snippet;
		class?: string;
		disabled?: boolean;
	} = $props();
</script>

<AlertDialog.Cancel
	class={cn(
		'inline-flex h-9 items-center justify-center rounded-lg border border-line bg-white px-4 text-sm font-medium text-ink-2',
		'hover:bg-surface-muted active:bg-surface-sunken',
		'disabled:pointer-events-none disabled:opacity-50',
		className,
	)}
	{disabled}
	{...restProps}
>
	{@render children()}
</AlertDialog.Cancel>
