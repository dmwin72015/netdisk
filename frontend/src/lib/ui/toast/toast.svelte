<script lang="ts">
	import { Toaster } from 'svelte-sonner';

	let { closeButton = true, ...rest }: { closeButton?: boolean; [key: string]: any } = $props();
</script>

<Toaster
	position="top-right"
	richColors
	{closeButton}
	toastOptions={{ duration: 3000 }}
	{...rest}
/>
