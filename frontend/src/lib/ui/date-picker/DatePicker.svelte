<script lang="ts">
	import { DatePicker } from 'bits-ui';
	import { CalendarDate, parseDate, today, getLocalTimeZone, type DateValue } from '@internationalized/date';
	import { ChevronLeft, ChevronRight, CalendarDays } from '@lucide/svelte';
	import { cn } from '$lib/utils/cn';
	import * as m from '$lib/paraglide/messages';

	let {
		value = $bindable(null),
		placeholder = $bindable(null),
		disabled = false,
		placeholderText = m.select_date(),
		onValueChange,
		class: className = '',
	}: {
		value?: Date | null;
		placeholder?: Date | null;
		disabled?: boolean;
		placeholderText?: string;
		onValueChange?: (date: Date | null) => void;
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

	let internalValue = $state<DateValue | undefined>(toDateValue(value));
	let internalPlaceholder = $state<DateValue>(toDateValue(placeholder) ?? today(tz));
	let open = $state(false);

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

	function openCalendar() {
		if (!disabled) open = true;
	}
</script>

<DatePicker.Root
	value={internalValue}
	onValueChange={handleValueChange}
	placeholder={internalPlaceholder}
	onPlaceholderChange={(p) => { if (p) internalPlaceholder = p; }}
	bind:open
	{disabled}
	closeOnDateSelect={true}
>
	<div class={cn('relative', className)}>
		<DatePicker.Input
			class="flex h-8 cursor-pointer items-center gap-1.5 rounded-lg border border-line bg-surface-sunken px-2.5 text-sm text-ink-2 outline-none transition-colors hover:border-line focus-within:border-primary"
			onclick={openCalendar}
		>
			{#snippet children({ segments })}
				{#each segments as { part, value }}
					{#if part === 'literal'}
						<span class="text-ink-4">{value}</span>
					{:else}
						<DatePicker.Segment
							{part}
							class="rounded px-0.5 tabular-nums outline-none focus:bg-primary-softer focus:text-primary"
						>
							{value}
						</DatePicker.Segment>
					{/if}
				{/each}
				<DatePicker.Trigger
					class="ml-auto inline-flex h-5 w-5 shrink-0 items-center justify-center rounded text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-3 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
					aria-label={placeholderText}
				>
					<CalendarDays size={14} />
				</DatePicker.Trigger>
			{/snippet}
		</DatePicker.Input>

		<DatePicker.Portal>
			<DatePicker.Content
				class="z-50 rounded-xl border border-line-soft bg-surface p-3 shadow-pop
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
								class="inline-flex h-7 w-7 items-center justify-center rounded-lg transition-colors hover:bg-surface-sunken"
							>
								<ChevronLeft size={14} />
							</DatePicker.PrevButton>
							<DatePicker.Heading class="text-sm font-medium text-ink" />
							<DatePicker.NextButton
								class="inline-flex h-7 w-7 items-center justify-center rounded-lg transition-colors hover:bg-surface-sunken"
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
												class="w-8 pb-1 text-center text-xs font-medium text-ink-4"
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
															hover:bg-surface-sunken focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary
															data-[selected]:bg-primary data-[selected]:text-primary-on data-[selected]:hover:bg-primary-hover
															data-[outside-month]:text-ink-4 data-[today]:font-medium data-[today]:text-primary
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
