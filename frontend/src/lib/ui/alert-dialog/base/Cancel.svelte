<script lang="ts">
	import { AlertDialog } from 'bits-ui';
	import type { Snippet } from 'svelte';

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
	class="inline-flex h-9 items-center justify-center rounded-lg border border-gray-200 bg-white px-4 text-sm font-medium text-gray-700
		hover:bg-gray-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2
		disabled:pointer-events-none disabled:opacity-50
		{className}"
	{disabled}
	{...restProps}
>
	{@render children()}
</AlertDialog.Cancel>
