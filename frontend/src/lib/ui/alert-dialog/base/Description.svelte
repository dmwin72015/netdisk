<script lang="ts">
	import { AlertDialog } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let {
		children,
		class: className = '',
		...restProps
	}: {
		children: Snippet;
		class?: string;
	} = $props();
</script>

<AlertDialog.Description
	class="mt-1 text-xs text-gray-500 {className}"
	{...restProps}
>
	{@render children()}
</AlertDialog.Description>
