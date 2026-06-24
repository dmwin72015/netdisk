<script lang="ts">
	import { DateRangePicker } from 'bits-ui';
	import {
		CalendarDate,
		today,
		getLocalTimeZone,
		type DateValue,
	} from '@internationalized/date';
	import { ChevronLeft, ChevronRight, CalendarDays } from '@lucide/svelte';
	import { cn } from '$lib/utils/cn';
	import * as m from '$lib/paraglide/messages';

	let {
		value = $bindable({ start: null, end: null }),
		placeholder = $bindable(null),
		disabled = false,
		placeholderText = m.select_date_range(),
		onValueChange,
		numberOfMonths = 2,
		class: className = '',
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

	let internalValue = $state<{ start: DateValue | undefined; end: DateValue | undefined }>({
		start: value?.start ? toDateValue(value.start) : undefined,
		end: value?.end ? toDateValue(value.end) : undefined,
	});
	let internalPlaceholder = $state<DateValue>(toDateValue(placeholder) ?? today(tz));
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

	function handleValueChange(v: { start: DateValue | undefined; end: DateValue | undefined }) {
		internalValue = v;
		const newValue = {
			start: toDate(v.start),
			end: toDate(v.end),
		};
		value = newValue;
		onValueChange?.(newValue);
	}
</script>

<DateRangePicker.Root
	value={internalValue}
	onValueChange={handleValueChange}
	placeholder={internalPlaceholder}
	onPlaceholderChange={(p) => { if (p) internalPlaceholder = p; }}
	bind:open
	{disabled}
	{numberOfMonths}
	closeOnRangeSelect={true}
	weekdayFormat="short"
>
	<div class={cn('relative', className)}>
		<div
			class="flex h-8 items-center gap-0.5 rounded-lg border border-line bg-surface px-3 text-sm text-ink-3 outline-none transition-colors hover:border-line focus-within:border-primary"
		>
			{#each ['start', 'end'] as const as type (type)}
				<DateRangePicker.Input {type}>
					{#snippet children({ segments })}
						{#each segments as { part, value }}
							{#if part === 'literal'}
								<span class="text-ink-4">{value}</span>
							{:else}
								<DateRangePicker.Segment
									{part}
									class="rounded px-0.5 tabular-nums outline-none focus:bg-primary-softer focus:text-primary"
								>
									{value}
								</DateRangePicker.Segment>
							{/if}
						{/each}
					{/snippet}
				</DateRangePicker.Input>
				{#if type === 'start'}
					<span class="text-ink-4 px-0.5" aria-hidden="true">—</span>
				{/if}
			{/each}
			<DateRangePicker.Trigger
				class="ml-auto inline-flex h-5 w-5 shrink-0 items-center justify-center rounded text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-3 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
				aria-label={placeholderText}
			>
				<CalendarDays size={14} />
			</DateRangePicker.Trigger>
		</div>

		<DateRangePicker.Content
			class="z-50 rounded-xl border border-line-soft bg-white p-3 shadow-pop
				data-[state=open]:animate-in data-[state=closed]:animate-out
				data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0
				data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95
				data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2"
			sideOffset={4}
		>
			<DateRangePicker.Calendar>
				{#snippet children({ months, weekdays })}
					<DateRangePicker.Header class="flex items-center justify-between pb-2">
						<DateRangePicker.PrevButton
							class="inline-flex h-7 w-7 items-center justify-center rounded-lg transition-colors hover:bg-surface-sunken"
						>
							<ChevronLeft size={14} />
						</DateRangePicker.PrevButton>
						<DateRangePicker.Heading class="text-sm font-medium text-ink" />
						<DateRangePicker.NextButton
							class="inline-flex h-7 w-7 items-center justify-center rounded-lg transition-colors hover:bg-surface-sunken"
						>
							<ChevronRight size={14} />
						</DateRangePicker.NextButton>
					</DateRangePicker.Header>
					{#each months as month}
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
											<DateRangePicker.Cell {date} month={month.value} class="p-0.5">
												<DateRangePicker.Day
													class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-sm transition-colors
														hover:bg-surface-sunken focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary
														data-[selected]:bg-primary data-[selected]:text-white
														data-[selection-start]:bg-primary data-[selection-start]:text-white data-[selection-start]:rounded-lg
														data-[selection-end]:bg-primary data-[selection-end]:text-white data-[selection-end]:rounded-lg
														data-[highlighted]:bg-primary-softer data-[highlighted]:rounded-none
														data-[outside-month]:text-ink-4 data-[today]:font-medium data-[today]:text-primary
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
					{/each}
				{/snippet}
			</DateRangePicker.Calendar>
		</DateRangePicker.Content>
	</div>
</DateRangePicker.Root>
