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

<AlertDialog.Action
	class={cn(
		'inline-flex h-9 items-center justify-center rounded-lg bg-danger px-4 text-sm font-medium text-primary-on',
		'hover:bg-danger-hover active:bg-danger-active',
		'disabled:pointer-events-none disabled:opacity-50',
		className,
	)}
	{disabled}
	{...restProps}
>
	{@render children()}
</AlertDialog.Action>
