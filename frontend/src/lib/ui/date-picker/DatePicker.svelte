<script lang="ts">
	import { DatePicker } from 'bits-ui';
	import { CalendarDate, parseDate, today, getLocalTimeZone, type DateValue } from '@internationalized/date';
	import { ChevronLeft, ChevronRight, CalendarDays } from '@lucide/svelte';
	import * as m from '$lib/paraglide/messages';

	let {
		value = $bindable(null),
		placeholder = $bindable(null),
		disabled = false,
		placeholderText = m.select_date(),
		onValueChange,
	}: {
		value?: Date | null;
		placeholder?: Date | null;
		disabled?: boolean;
		placeholderText?: string;
		onValueChange?: (date: Date | null) => void;
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

	let internalValue = $state<DateValue | undefined>(toDateValue(value));
	let internalPlaceholder = $state<DateValue>(toDateValue(placeholder) ?? today(tz));

	$effect(() => {
		internalValue = toDateValue(value);
	});

	$effect(() => {
		if (placeholder) internalPlaceholder = toDateValue(placeholder)!;
	});

	function handleValueChange(v: DateValue | undefined) {
		internalValue = v;
		const d = toDate(v) ?? null;
		value = d;
		onValueChange?.(d);
	}
</script>

<DatePicker.Root
	value={internalValue}
	onValueChange={handleValueChange}
	placeholder={internalPlaceholder}
	onPlaceholderChange={(p) => { if (p) internalPlaceholder = p; }}
	{disabled}
	closeOnDateSelect={true}
>
	<div class="relative">
		<DatePicker.Input
			class="flex h-8 items-center gap-1.5 rounded-lg border border-gray-200 bg-gray-50 px-2.5 text-sm text-gray-700 outline-none transition-colors hover:border-gray-300 focus-within:border-blue-400 focus-within:bg-white"
		>
			{#snippet children({ segments })}
				{#each segments as { part, value }}
					{#if part === 'literal'}
						<span class="text-gray-400">{value}</span>
					{:else}
						<DatePicker.Segment
							{part}
							class="rounded px-0.5 tabular-nums outline-none focus:bg-blue-100 focus:text-blue-700"
						>
							{value}
						</DatePicker.Segment>
					{/if}
				{/each}
				<CalendarDays size={14} class="ml-auto shrink-0 text-gray-400" />
			{/snippet}
		</DatePicker.Input>

		<DatePicker.Portal>
			<DatePicker.Content
				class="z-50 rounded-xl border border-gray-100 bg-white p-3 shadow-lg
					data-[state=open]:animate-in data-[state=closed]:animate-out
					data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0
					data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95
					data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2"
				sideOffset={4}
			>
				<DatePicker.Calendar>
					{#snippet children({ months, weekdays })}
						<DatePicker.Header class="flex items-center justify-between pb-2">
							<DatePicker.PrevButton
								class="inline-flex h-7 w-7 items-center justify-center rounded-lg transition-colors hover:bg-gray-100"
							>
								<ChevronLeft size={14} />
							</DatePicker.PrevButton>
							<DatePicker.Heading class="text-sm font-medium text-gray-900" />
							<DatePicker.NextButton
								class="inline-flex h-7 w-7 items-center justify-center rounded-lg transition-colors hover:bg-gray-100"
							>
								<ChevronRight size={14} />
							</DatePicker.NextButton>
						</DatePicker.Header>
						{#each months as month}
							<DatePicker.Grid class="border-collapse">
								<DatePicker.GridHead>
									<DatePicker.GridRow class="flex">
										{#each weekdays as day}
											<DatePicker.HeadCell
												class="w-8 pb-1 text-center text-xs font-medium text-gray-400"
											>
												{day}
											</DatePicker.HeadCell>
										{/each}
									</DatePicker.GridRow>
								</DatePicker.GridHead>
								<DatePicker.GridBody>
									{#each month.weeks as weekDates}
										<DatePicker.GridRow class="flex">
											{#each weekDates as date}
														<DatePicker.Cell {date} month={month.value} class="p-0.5">
													<DatePicker.Day
														class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-sm transition-colors
															hover:bg-gray-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500
															data-[selected]:bg-blue-600 data-[selected]:text-white data-[selected]:hover:bg-blue-700
															data-[outside-month]:text-gray-300 data-[today]:font-medium data-[today]:text-blue-600
															data-[disabled]:cursor-not-allowed data-[disabled]:opacity-40"
													>
														{date.day}
													</DatePicker.Day>
												</DatePicker.Cell>
											{/each}
										</DatePicker.GridRow>
									{/each}
								</DatePicker.GridBody>
							</DatePicker.Grid>
						{/each}
					{/snippet}
				</DatePicker.Calendar>
			</DatePicker.Content>
		</DatePicker.Portal>
	</div>
</DatePicker.Root>
