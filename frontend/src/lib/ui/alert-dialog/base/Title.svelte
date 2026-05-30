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

<AlertDialog.Title
	class="text-sm font-medium leading-5 text-gray-800 {className}"
	{...restProps}
>
	{@render children()}
</AlertDialog.Title>
