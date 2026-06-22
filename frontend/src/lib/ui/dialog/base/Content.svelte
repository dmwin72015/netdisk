<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import Overlay from './Overlay.svelte';

	let {
		children,
		class: className = '',
		style = '',
		...restProps
	}: {
		children: Snippet;
		class?: string;
		style?: string;
	} = $props();
	</script>

	<Dialog.Portal>
		<Overlay />
		<Dialog.Content
			class="bg-surface text-ink border-line shadow-dialog fixed left-1/2 top-1/2 z-50 w-full -translate-x-1/2 -translate-y-1/2
				rounded-xl border
				data-[state=open]:animate-in data-[state=closed]:animate-out
				data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0
				data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95
				duration-200
				{className}"
			{style}
			{...restProps}
		>
			{@render children()}
		</Dialog.Content>
	</Dialog.Portal>
