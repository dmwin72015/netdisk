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

<AlertDialog.Portal>
	<AlertDialog.Overlay
		class="fixed inset-0 z-50 bg-overlay backdrop-blur-sm
			data-[state=open]:animate-in data-[state=closed]:animate-out
			data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0
			duration-200"
	/>
	<AlertDialog.Content
		class="fixed left-1/2 top-1/2 z-50 w-full max-w-lg -translate-x-1/2 -translate-y-1/2
			rounded-xl border border-line bg-white p-6 shadow-dialog
			data-[state=open]:animate-in data-[state=closed]:animate-out
			data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0
			data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95
			duration-200
			{className}"
		{...restProps}
	>
		{@render children()}
	</AlertDialog.Content>
</AlertDialog.Portal>
