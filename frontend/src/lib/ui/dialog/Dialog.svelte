<script lang="ts">
  import type { Snippet } from "svelte";
  import { X } from "@lucide/svelte";
  import * as DialogBase from "./base";
  import { cn } from "$lib/utils/cn";

  let {
    open = $bindable(false),
    title,
    description,
    okText,
    cancelText,
    confirmText,
    onOpenChange,
    onOpenChangeComplete,
    afterOpenChange,
    onOk,
    onCancel,
    onConfirm,
    children,
    headerExtra,
    footer = true,
    showFooter,
    showCancel,
    showConfirm,
    headerClass = "",
    contentStyle = "",
    titleClass = "",
    descriptionClass = "",
    bodyClass = "",
    closeButtonClass = "",
    closeIconSize = 18,
    closable = true,
    width,
    class: className = "",
  }: {
    open?: boolean;
    title?: string;
    description?: string;
    okText?: string;
    cancelText?: string;
    confirmText?: string;
    onOpenChange?: (open: boolean) => void;
    onOpenChangeComplete?: (open: boolean) => void;
    afterOpenChange?: (open: boolean) => void;
    onOk?: () => void;
    onCancel?: () => void;
    onConfirm?: () => void;
    children?: Snippet;
    headerExtra?: Snippet;
    footer?: boolean;
    showFooter?: boolean;
    showCancel?: boolean;
    showConfirm?: boolean;
    headerClass?: string;
    contentStyle?: string;
    titleClass?: string;
    descriptionClass?: string;
    bodyClass?: string;
    closeButtonClass?: string;
    closeIconSize?: number;
    closable?: boolean;
    width?: string;
    class?: string;
  } = $props();

  const shouldShowFooter = $derived(showFooter ?? footer);
  const shouldShowCancel = $derived(showCancel ?? true);
  const shouldShowOk = $derived(showConfirm ?? true);
  const resolvedOkText = $derived(okText ?? confirmText ?? "Confirm");
  const resolvedCancelText = $derived(cancelText ?? "Cancel");

  function handleOpenChangeComplete(isOpen: boolean) {
    onOpenChangeComplete?.(isOpen);
    afterOpenChange?.(isOpen);
  }

  function handleConfirm() {
    onOk?.();
    onConfirm?.();
    open = false;
  }

  function handleCancel() {
    onCancel?.();
    open = false;
  }
</script>

<DialogBase.Root bind:open {onOpenChange} onOpenChangeComplete={handleOpenChangeComplete}>
  <DialogBase.Content
    class={cn('max-h-[90vh] flex flex-col overflow-hidden', className)}
    style={width ? `max-width: ${width}; ${contentStyle}` : contentStyle}
  >
    <div
      class="border-line-soft flex {description
        ? 'items-start'
        : 'items-center'} gap-3 border-b px-5 py-3 {headerClass}"
    >
      <DialogBase.Header class="min-w-0 flex-1 space-y-1">
        {#if title}
          <DialogBase.Title
            class="text-ink truncate text-[15px] font-semibold leading-5 tracking-tight {titleClass}"
            >{title}</DialogBase.Title
          >
        {/if}
        {#if description}
          <DialogBase.Description class="text-ink-3 text-xs {descriptionClass}"
            >{description}</DialogBase.Description
          >
        {/if}
      </DialogBase.Header>
      {#if headerExtra}
        <div class="flex shrink-0 items-center gap-1">
          {@render headerExtra()}
        </div>
      {/if}
      {#if closable}
        <DialogBase.Close
          class="text-ink-4 hover:bg-surface-sunken hover:text-ink rounded-md p-1.5 transition-colors duration-150 {closeButtonClass}"
          aria-label="Close"
          onclick={handleCancel}
        >
          <X size={closeIconSize} strokeWidth={1.75} />
        </DialogBase.Close>
      {/if}
    </div>

    <div class="flex-1 overflow-auto {cn('px-5 py-4', bodyClass)}">
      {#if children}
        {@render children()}
      {/if}
    </div>

    {#if shouldShowFooter && (shouldShowCancel || shouldShowOk)}
      <DialogBase.Footer class="border-line-soft border-t px-5 py-3">
        {#if shouldShowCancel}
          <DialogBase.Close
            class="text-ink-2 hover:bg-surface-sunken hover:text-ink inline-flex h-8 items-center justify-center rounded-md px-3 text-sm transition-colors duration-150"
            onclick={handleCancel}
          >
            {resolvedCancelText}
          </DialogBase.Close>
        {/if}
        {#if shouldShowOk}
          <button
            type="button"
            class="bg-primary hover:bg-primary-hover active:bg-primary-active text-primary-on inline-flex h-8 items-center justify-center rounded-md px-3.5 text-sm font-medium transition-colors duration-150"
            onclick={handleConfirm}
          >
            {resolvedOkText}
          </button>
        {/if}
      </DialogBase.Footer>
    {/if}
  </DialogBase.Content>
</DialogBase.Root>
