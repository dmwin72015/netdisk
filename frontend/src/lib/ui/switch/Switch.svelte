<script lang="ts">
  import { Switch as SwitchPrimitive } from "bits-ui";
  import { cn } from "$lib/utils/cn";

  let {
    checked = $bindable(false),
    onCheckedChange,
    disabled = false,
    class: className = "",
  }: {
    checked?: boolean;
    onCheckedChange?: (checked: boolean) => void;
    disabled?: boolean;
    class?: string;
  } = $props();
</script>

<SwitchPrimitive.Root
  bind:checked
  {onCheckedChange}
  {disabled}
  class={cn(
    "peer inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors",
    "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/30 focus-visible:ring-offset-2 focus-visible:ring-offset-surface",
    "disabled:cursor-not-allowed disabled:opacity-50",
    "data-[state=checked]:bg-primary data-[state=unchecked]:bg-line",
    "hover:data-[state=checked]:bg-primary-hover hover:data-[state=unchecked]:bg-line-hover",
    className,
  )}
>
  <SwitchPrimitive.Thumb
    class={cn(
      "pointer-events-none block h-4 w-4 rounded-full bg-white shadow-sm ring-0 transition-transform",
      "data-[state=checked]:translate-x-4 data-[state=unchecked]:translate-x-0",
    )}
  />
</SwitchPrimitive.Root>