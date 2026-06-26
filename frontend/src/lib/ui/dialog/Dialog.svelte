<script lang="ts">
  import type { Snippet } from "svelte";
  import { X } from "@lucide/svelte";
  import * as DialogBase from "./base";
  import Spinner from "$lib/ui/spinner/Spinner.svelte";
  import { cn } from "$lib/utils/cn";
  import { isDefined } from "$lib/utils/valid";

  type DialogSize = "sm" | "md" | "lg" | "xl";

  const sizeWidths: Record<DialogSize, string> = {
    sm: "w-[380px] xl:w-[400px] 2xl:w-[420px] min-[2200px]:w-[440px]",
    md: "w-[600px] xl:w-[640px] 2xl:w-[680px] min-[2200px]:w-[720px]",
    lg: "w-[800px] xl:w-[860px] 2xl:w-[920px] min-[2200px]:w-[1000px]",
    xl: "w-[1000px] xl:w-[1100px] 2xl:w-[1200px] min-[2200px]:w-[1300px]",
  };
  const DEFAULT_SIZE: DialogSize = "sm";

  let {
    open = $bindable(false),
    title,
    description,
    cancelText = "Cancel",
    confirmText = "Confirm",
    onOpenChange,
    onOpenChangeComplete,
    onOpenAutoFocus,
    onCloseAutoFocus,
    onCancel,
    onConfirm,
    children,
    headerExtra,
    footer = true,
    showCancel = true,
    showConfirm = true,
    bodyClass = "",
    closable = true,
    size,
    width,
    contentStyle = "",
    class: className = "",
    ...restProps
  }: {
    open?: boolean;
    title?: string;
    description?: string;
    cancelText?: string;
    confirmText?: string;
    onOpenChange?: (open: boolean) => void;
    onOpenChangeComplete?: (open: boolean) => void;
    onOpenAutoFocus?: (e: Event) => void;
    onCloseAutoFocus?: (e: Event) => void;
    onCancel?: () => void;
    onConfirm?: () => void | Promise<void>;
    children?: Snippet;
    headerExtra?: Snippet;
    footer?: boolean;
    showCancel?: boolean;
    showConfirm?: boolean;
    bodyClass?: string;
    closable?: boolean;
    size?: DialogSize;
    width?: string | number;
    contentStyle?: string;
    class?: string;
  } = $props();

  const sizeClass = width ? "" : sizeWidths[size ?? DEFAULT_SIZE];

  const contentClass = $derived(
    cn("max-h-[90vh] flex flex-col overflow-hidden", sizeClass, className),
  );

  const widthStyle = isDefined(width)
    ? `width: ${typeof width === "number" ? `${width}px` : width};`
    : "";

  const contentInlineStyle = [contentStyle, widthStyle].filter(Boolean).join(" ").trim() || undefined;

  let confirming = false;

  async function handleConfirm() {
    if (!onConfirm) {
      open = false;
      return;
    }
    const result = onConfirm();
    if (result instanceof Promise) {
      confirming = true;
      try {
        await result;
        open = false;
      } catch {
        // keep dialog open on failure
      } finally {
        confirming = false;
      }
    } else {
      open = false;
    }
  }
</script>

<DialogBase.Root
  bind:open
  {onOpenChange}
  {onOpenChangeComplete}
>
  <DialogBase.Content class={contentClass} style={contentInlineStyle} {onOpenAutoFocus} {onCloseAutoFocus} {...restProps}>
    <div
      class="border-line-soft flex items-center gap-3 border-b px-5 py-3"
    >
      <DialogBase.Header class="min-w-0 flex-1 space-y-1">
        {#if title}
          <DialogBase.Title class="text-ink truncate text-[15px] font-semibold leading-5 tracking-tight">{title}</DialogBase.Title>
        {/if}
        {#if description}
          <DialogBase.Description class="text-ink-3 text-xs">{description}</DialogBase.Description>
        {/if}
      </DialogBase.Header>
      {#if headerExtra}
        <div class="flex shrink-0 items-center gap-1">
          {@render headerExtra()}
        </div>
      {/if}
      {#if closable}
        <DialogBase.Close
          class="text-ink-4 hover:bg-surface-sunken hover:text-ink rounded-md p-1.5 transition-colors duration-150 disabled:opacity-50"
          aria-label="Close"
          disabled={confirming}
          onclick={onCancel}
        >
          <X size={18} strokeWidth={1.75} />
        </DialogBase.Close>
      {/if}
    </div>

    <div class="flex-1 overflow-auto {cn('px-5 py-4', bodyClass)}">
      {#if children}
        {@render children()}
      {/if}
    </div>

    {#if footer && (showCancel || showConfirm)}
      <DialogBase.Footer class="border-line-soft border-t px-5 py-3">
        {#if showCancel}
          <DialogBase.Close
            class="text-ink-2 hover:bg-surface-sunken hover:text-ink inline-flex h-8 items-center justify-center rounded-md px-3 text-sm transition-colors duration-150 disabled:opacity-50"
            disabled={confirming}
            onclick={onCancel}
          >
            {cancelText}
          </DialogBase.Close>
        {/if}
        {#if showConfirm}
          <button
            type="button"
            class="bg-primary hover:bg-primary-hover active:bg-primary-active text-primary-on inline-flex h-8 items-center justify-center gap-1.5 rounded-md px-3.5 text-sm font-medium transition-colors duration-150 disabled:opacity-50"
            disabled={confirming}
            onclick={handleConfirm}
          >
            {#if confirming}
              <Spinner size={14} strokeWidth={2.5} />
            {/if}
            {confirmText}
          </button>
        {/if}
      </DialogBase.Footer>
    {/if}
  </DialogBase.Content>
</DialogBase.Root>
