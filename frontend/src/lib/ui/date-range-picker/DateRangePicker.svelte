<script lang="ts">
  import { DateRangePicker } from "bits-ui";
  import {
    CalendarDate,
    today,
    getLocalTimeZone,
    type DateValue,
  } from "@internationalized/date";
  import { ChevronLeft, ChevronRight, CalendarDays } from "@lucide/svelte";
  import { cn } from "$lib/utils/cn";
  import * as m from "$lib/paraglide/messages";

  let {
    value = $bindable({ start: null, end: null }),
    placeholder = $bindable(null),
    disabled = false,
    placeholderText = m.select_date_range(),
    onValueChange,
    numberOfMonths = 2,
    class: className = "",
  }: {
    value?: { start: Date | null; end: Date | null };
    placeholder?: Date | null;
    disabled?: boolean;
    placeholderText?: string;
    onValueChange?: (range: { start: Date | null; end: Date | null }) => void;
    numberOfMonths?: number;
    class?: string;
  } = $props();

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
    if (!value.start && !value.end) return placeholderText;
    if (value.start && !value.end) return `${formatDateRange(value.start)} ~`;
    if (value.start && value.end) return `${formatDateRange(value.start)} ~ ${formatDateRange(value.end)}`;
    return placeholderText;
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
  weekdayFormat="short"
>
  <div class={cn("relative", className)}>
    <!-- Trigger: AntD-style placeholder display -->
    <DateRangePicker.Trigger
      class="flex h-8 w-full items-center rounded-lg border border-line bg-surface px-3 text-sm text-ink-3 outline-none transition-colors hover:border-line focus:border-primary"
    >
      <span class={cn(!displayValue ? "text-ink-4" : "text-ink-3")}>
        {displayValue || placeholderText}
      </span>
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
      <!-- Single header that applies to all months -->
      <DateRangePicker.Calendar>
        {#snippet children({ months, weekdays })}
          <!-- Custom month headers -->
          <div class="flex gap-4 pb-2">
            {#each months as month}
              <div class="flex-1 text-center">
                <span class="text-sm font-medium text-ink">
                  {month.value.month} / {month.value.year}
                </span>
              </div>
            {/each}
          </div>

          <!-- Shared navigation controls -->
          <div class="flex items-center justify-between pb-2">
            <DateRangePicker.PrevButton
              class="inline-flex h-7 w-7 items-center justify-center rounded-lg transition-colors hover:bg-surface-sunken"
            >
              <ChevronLeft size={14} />
            </DateRangePicker.PrevButton>
            <DateRangePicker.NextButton
              class="inline-flex h-7 w-7 items-center justify-center rounded-lg transition-colors hover:bg-surface-sunken"
            >
              <ChevronRight size={14} />
            </DateRangePicker.NextButton>
          </div>

          <!-- Months displayed horizontally -->
          <div class="flex gap-4">
            {#each months as month}
              <div class="flex-1">
                <DateRangePicker.Grid class="border-collapse">
                  <DateRangePicker.GridHead>
                    <DateRangePicker.GridRow class="flex">
                      {#each weekdays as day}
                        <DateRangePicker.HeadCell
                          class="w-8 pb-1 text-center text-xs font-medium text-ink-4"
                        >
                          {day.slice(0, 2)}
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
                              class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-sm transition-colors
                                hover:bg-surface-sunken focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary
                                data-[selected]:bg-primary-softer data-[selected]:text-primary
                                data-[selection-start]:bg-primary-softer data-[selection-start]:text-primary
                                data-[selection-end]:bg-primary-softer data-[selection-end]:text-primary
                                data-[highlighted]:bg-primary-softer
                                data-[outside-month]:invisible
                                data-[today]:font-medium data-[today]:text-primary
                                data-[disabled]:cursor-not-allowed data-[disabled]:opacity-40"
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
