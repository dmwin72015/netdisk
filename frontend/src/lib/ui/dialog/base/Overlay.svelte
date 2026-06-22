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
	class="bg-overlay fixed inset-0 z-50 backdrop-blur-[2px]
		data-[state=open]:animate-in data-[state=closed]:animate-out
		data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0
		duration-200
		{className}"
	{...restProps}
/>
