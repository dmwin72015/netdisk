<script lang="ts">
	import { Dialog } from 'bits-ui';

	let {
		class: className = '',
		...restProps
	}: {
		class?: string;
	} = $props();
</script>

<Dialog.Overlay
	class="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm
		data-[state=open]:animate-in data-[state=closed]:animate-out
		data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0
		{className}"
	{...restProps}
/>
