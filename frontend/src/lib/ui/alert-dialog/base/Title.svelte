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
	class="text-lg font-semibold text-gray-900 {className}"
	{...restProps}
>
	{@render children()}
</AlertDialog.Title>
