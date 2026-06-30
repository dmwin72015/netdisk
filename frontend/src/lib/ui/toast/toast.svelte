<script lang="ts">
  import { Toaster } from 'svelte-sonner';

  let {
    closeButton = true,
    theme = 'light',
    ...rest
  }: { closeButton?: boolean; theme?: 'light' | 'dark' | 'system'; [key: string]: any } = $props();
</script>

<Toaster
  position="top-right"
  richColors
  {closeButton}
  {theme}
  toastOptions={{ duration: 3000 }}
  {...rest}
/>
