<script lang="ts">
  import { DateRangePicker } from "bits-ui";
  import {
    CalendarDate,
    today,
    getLocalTimeZone,
    type DateValue,
  } from "@internationalized/date";
  import { ChevronLeft, ChevronRight, CalendarDays, X } from "@lucide/svelte";
  import { cn } from "$lib/utils/cn";

  /** Self-contained i18n data — no external dependencies. */
  const MONTH_NAMES: Record<string, string[]> = {
    zh: ["1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"],
    en: ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"],
  };

  const WEEKDAY_NAMES: Record<string, string[]> = {
    zh: ["日", "一", "二", "三", "四", "五", "六"],
    en: ["Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"],
  };

  const PLACEHOLDER_DEFAULT: Record<string, string> = {
    zh: "选择日期范围",
    en: "Select date range",
  };

  let {
    value = $bindable({ start: null, end: null }),
    placeholder = $bindable(null),
    disabled = false,
    locale = "zh",
    placeholderText,
    onValueChange,
    onClear,
    numberOfMonths = 2,
    class: className = "",
  }: {
    value?: { start: Date | null; end: Date | null };
    placeholder?: Date | null;
    disabled?: boolean;
    locale?: string;
    placeholderText?: string;
    onValueChange?: (range: { start: Date | null; end: Date | null }) => void;
    onClear?: () => void;
    numberOfMonths?: number;
    class?: string;
  } = $props();

  const effectiveLocale = MONTH_NAMES[locale] ? locale : "zh";
  const effectivePlaceholder = placeholderText ?? PLACEHOLDER_DEFAULT[effectiveLocale] ?? PLACEHOLDER_DEFAULT.zh;

  const tz = getLocalTimeZone();

  function toDate(d: DateValue | undefined): Date | null {
    if (!d) return null;
    return d.toDate(tz);
  }

  function toDateValue(d: Date | null): DateValue | undefined {
    if (!d) return undefined;
    return new CalendarDate(d.getFullYear(), d.getMonth() + 1, d.getDate());
  }

  function formatDateRange(d: Date | null): string {
    if (!d) return "";
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");
    return `${y}-${m}-${day}`;
  }

  function formatDisplayValue(): string {
    if (!value.start && !value.end) return effectivePlaceholder;
    if (value.start && !value.end) return `${formatDateRange(value.start)} ~`;
    if (value.start && value.end) return `${formatDateRange(value.start)} ~ ${formatDateRange(value.end)}`;
    return effectivePlaceholder;
  }

  const monthNames = $derived(MONTH_NAMES[effectiveLocale]);
  const weekdayNames = $derived(WEEKDAY_NAMES[effectiveLocale]);

  function monthLabel(monthNum: number, year: number): string {
    return `${monthNames[monthNum - 1]} ${year}`;
  }

  let internalValue = $state<{
    start: DateValue | undefined;
    end: DateValue | undefined;
  }>({
    start: value?.start ? toDateValue(value.start) : undefined,
    end: value?.end ? toDateValue(value.end) : undefined,
  });
  let internalPlaceholder = $state<DateValue>(
    toDateValue(placeholder) ?? today(tz),
  );
  let open = $state(false);

  $effect(() => {
    internalValue = {
      start: value?.start ? toDateValue(value.start) : undefined,
      end: value?.end ? toDateValue(value.end) : undefined,
    };
  });

  $effect(() => {
    if (placeholder) internalPlaceholder = toDateValue(placeholder)!;
  });

  function handleValueChange(v: {
    start: DateValue | undefined;
    end: DateValue | undefined;
  }) {
    internalValue = v;
    const newValue = {
      start: toDate(v.start),
      end: toDate(v.end),
    };
    value = newValue;
    onValueChange?.(newValue);
  }

  let displayValue = $derived(formatDisplayValue());

  function isInRange(date: DateValue): boolean {
    if (!internalValue.start || !internalValue.end) return false;
    return date.compare(internalValue.start) > 0 && date.compare(internalValue.end) < 0;
  }

  function isStart(date: DateValue): boolean {
    return !!internalValue.start && date.compare(internalValue.start) === 0;
  }

  function isEnd(date: DateValue): boolean {
    return !!internalValue.end && date.compare(internalValue.end) === 0;
  }
</script>

<DateRangePicker.Root
  value={internalValue}
  onValueChange={handleValueChange}
  placeholder={internalPlaceholder}
  onPlaceholderChange={(p) => {
    if (p) internalPlaceholder = p;
  }}
  bind:open
  {disabled}
  {numberOfMonths}
  closeOnRangeSelect={true}
>
  <div class={cn("relative", className)}>
    <!-- Trigger: AntD-style placeholder display -->
    <DateRangePicker.Trigger
      class="flex h-8 w-full items-center rounded-lg border border-line bg-surface px-3 text-sm text-ink-3 outline-none transition-colors hover:border-line focus:border-primary"
    >
      <span class={cn(!displayValue ? "text-ink-4" : "text-ink-3")}>
        {displayValue || effectivePlaceholder}
      </span>
      {#if onClear && (value.start || value.end)}
        <button
          type="button"
          onclick={(e) => { e.stopPropagation(); onClear?.(); }}
          class="ml-1 flex h-5 w-5 shrink-0 items-center justify-center rounded-full text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-2"
        >
          <X size={12} />
        </button>
      {/if}
      <CalendarDays size={14} class="ml-auto shrink-0 text-ink-4" />
    </DateRangePicker.Trigger>

    <!-- Calendar popover: horizontal months -->
    <DateRangePicker.Content
      class="z-50 rounded-xl border border-line-soft bg-white p-3 shadow-pop
        data-[state=open]:animate-in data-[state=closed]:animate-out
        data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0
        data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95
        data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2"
      sideOffset={4}
    >
      <DateRangePicker.Calendar>
        {#snippet children({ months })}
          <!-- Months displayed horizontally with individual headers and nav -->
          <div class="flex gap-4">
            {#each months as month}
              <div class="flex-1">
                <!-- Per-panel month title: "Jun 2026" or "6月 2026" -->
                <div class="mb-2 text-center text-sm font-medium text-ink">
                  {monthLabel(month.value.month, month.value.year)}
                </div>

                <!-- Per-panel navigation: < > -->
                <div class="mb-2 flex items-center justify-between">
                  <DateRangePicker.PrevButton
                    class="inline-flex h-6 w-6 items-center justify-center rounded transition-colors hover:bg-surface-sunken text-ink-4"
                  >
                    <ChevronLeft size={12} />
                  </DateRangePicker.PrevButton>
                  <DateRangePicker.NextButton
                    class="inline-flex h-6 w-6 items-center justify-center rounded transition-colors hover:bg-surface-sunken text-ink-4"
                  >
                    <ChevronRight size={12} />
                  </DateRangePicker.NextButton>
                </div>

                <DateRangePicker.Grid class="border-collapse">
                  <DateRangePicker.GridHead>
                    <DateRangePicker.GridRow class="flex">
                      {#each weekdayNames as day}
                        <DateRangePicker.HeadCell
                          class="w-8 pb-1 text-center text-xs font-medium text-ink-4"
                        >
                          {day}
                        </DateRangePicker.HeadCell>
                      {/each}
                    </DateRangePicker.GridRow>
                  </DateRangePicker.GridHead>
                  <DateRangePicker.GridBody>
                    {#each month.weeks as weekDates}
                      <DateRangePicker.GridRow class="flex">
                        {#each weekDates as date}
                          <DateRangePicker.Cell
                            {date}
                            month={month.value}
                            class="p-0.5"
                          >
                            <DateRangePicker.Day
                              class={cn(
                                "inline-flex h-8 w-8 items-center justify-center text-sm transition-colors",
                                (isStart(date) || isEnd(date)) && "bg-[#165DFF] text-white rounded-full",
                                !isStart(date) && !isEnd(date) && "rounded-lg",
                                isInRange(date) && !isStart(date) && !isEnd(date) && "bg-[#E6F4FF]",
                                !isStart(date) && !isEnd(date) && !isInRange(date) && "hover:bg-surface-sunken",
                                "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary",
                                "data-[outside-month]:text-ink-4",
                                "data-[today]:font-medium data-[today]:text-primary",
                                "data-[disabled]:cursor-not-allowed data-[disabled]:opacity-40"
                              )}
                            >
                              {date.day}
                            </DateRangePicker.Day>
                          </DateRangePicker.Cell>
                        {/each}
                      </DateRangePicker.GridRow>
                    {/each}
                  </DateRangePicker.GridBody>
                </DateRangePicker.Grid>
              </div>
            {/each}
          </div>
        {/snippet}
      </DateRangePicker.Calendar>
    </DateRangePicker.Content>
  </div>
</DateRangePicker.Root>
