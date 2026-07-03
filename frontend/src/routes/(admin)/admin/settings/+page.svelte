<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { LoaderCircle, RotateCcw, Pencil, Settings, RotateCw } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import {
		adminListSystemConfig,
		adminUpdateSystemConfig,
		adminResetSystemConfig,
		type SystemConfigItem
	} from '$lib/api/admin';
	import { fmtSize } from '$lib/utils/format';
	import Dialog from '$lib/ui/dialog/Dialog.svelte';
		import { Dropdown, DropdownBase } from '$lib/ui/dropdown';
		import { ChevronDown, Check } from '@lucide/svelte';
	import * as m from '$lib/paraglide/messages';

	let items = $state<SystemConfigItem[]>([]);
	let loading = $state(true);
	let saving = $state(false);

	let editingItem = $state<SystemConfigItem | null>(null);
	let showDialog = $state(false);
	let editValue = $state<string>('');
	let editUnit = $state<number>(1); // bytes-per-unit multiplier
	let editError = $state('');

	const BYTE_UNITS: { label: string; value: number }[] = [
		{ label: 'B', value: 1 },
		{ label: 'KB', value: 1024 },
		{ label: 'MB', value: 1024 ** 2 },
		{ label: 'GB', value: 1024 ** 3 },
		{ label: 'TB', value: 1024 ** 4 }
	];

	function bestUnit(bytes: number): number {
		if (!Number.isFinite(bytes) || bytes <= 0) return 1;
		for (let i = BYTE_UNITS.length - 1; i >= 0; i--) {
			const u = BYTE_UNITS[i].value;
			if (bytes >= u && bytes % u === 0) return u;
		}
		return 1;
	}

	let bytesPreview = $derived.by(() => {
		if (!editingItem || editingItem.type !== 'bytes') return 0;
		const n = Number(editValue);
		if (!Number.isFinite(n) || n < 0) return 0;
		return Math.round(n * editUnit);
	});

	const typeLabels: Record<string, string> = {
		bytes: m.admin_config_type_bytes(),
		number: m.admin_config_type_number(),
		string: m.admin_config_type_string(),
		bool: m.admin_config_type_bool()
	};

	onMount(() => {
		if (!browser) return;
		load();
	});

	async function load() {
		loading = true;
		try {
			items = await adminListSystemConfig();
		} catch {
			toast.error(m.admin_load_failed());
		} finally {
			loading = false;
		}
	}

	function fmtValue(item: SystemConfigItem): string {
		if (item.type === 'bytes' && typeof item.value === 'number') {
			return fmtSize(item.value);
		}
		return String(item.value);
	}

	function fmtDefault(item: SystemConfigItem): string {
		if (item.type === 'bytes' && typeof item.defaultValue === 'number') {
			return fmtSize(item.defaultValue);
		}
		return String(item.defaultValue);
	}

	function isModified(item: SystemConfigItem): boolean {
		return item.value !== item.defaultValue;
	}

	function openEdit(item: SystemConfigItem) {
		editingItem = item;
		editError = '';
		if (item.type === 'bytes' && typeof item.value === 'number') {
			const unit = bestUnit(item.value);
			editUnit = unit;
			editValue = String(item.value / unit);
		} else {
			editUnit = 1;
			editValue = String(item.value);
		}
		showDialog = true;
	}

	function closeEdit() {
		showDialog = false;
		editingItem = null;
		editValue = '';
		editUnit = 1;
		editError = '';
	}

	async function saveEdit() {
		if (!editingItem) return;
		const item = editingItem;

		let parsed: unknown;
		if (item.type === 'bytes' || item.type === 'number') {
			const n = Number(editValue);
			if (isNaN(n) || n < 0) {
				editError = m.admin_config_invalid_number();
				return;
			}
			parsed = item.type === 'bytes' ? Math.round(n * editUnit) : n;
		} else if (item.type === 'bool') {
			parsed = editValue === 'true';
		} else {
			parsed = editValue;
		}

		saving = true;
		try {
			items = await adminUpdateSystemConfig({ [item.key]: parsed });
			closeEdit();
			toast.success(m.admin_config_saved());
		} catch {
			toast.error(m.admin_update_failed());
		} finally {
			saving = false;
		}
	}

	async function handleReset(item: SystemConfigItem) {
		try {
			items = await adminResetSystemConfig(item.key);
			toast.success(m.admin_config_reset());
		} catch {
			toast.error(m.admin_update_failed());
		}
	}

	async function handleResetAll() {
		try {
			items = await adminResetSystemConfig();
			toast.success(m.admin_config_reset_all());
		} catch {
			toast.error(m.admin_update_failed());
		}
	}
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-ink">{m.admin_settings()}</h1>
			<p class="mt-1 text-sm text-ink-4">{m.admin_settings_desc()}</p>
		</div>
		<button
			class="flex items-center gap-2 rounded-lg border border-line bg-surface px-4 py-2 text-sm font-medium text-ink transition-colors hover:bg-surface-sunken disabled:opacity-50"
			onclick={handleResetAll}
			disabled={loading}
		>
			<RotateCw size={15} />
			{m.admin_reset_all()}
		</button>
	</div>

	{#if loading}
		<div class="flex justify-center py-16">
			<LoaderCircle size={24} class="animate-spin text-ink-4" />
		</div>
	{:else}
		<div class="overflow-hidden rounded-xl border border-line bg-surface">
			<table class="w-full text-left text-sm">
				<thead class="border-b border-line bg-surface-sunken text-xs text-ink-3">
					<tr>
						<th class="px-5 py-3 font-medium">{m.admin_config_col_setting()}</th>
						<th class="px-5 py-3 font-medium">{m.admin_config_col_current()}</th>
						<th class="px-5 py-3 font-medium">{m.admin_config_col_default()}</th>
						<th class="px-5 py-3 font-medium">{m.admin_config_col_type()}</th>
						<th class="px-5 py-3 text-right font-medium">{m.admin_config_col_actions()}</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-line">
					{#each items as item (item.key)}
						<tr class="transition-colors hover:bg-surface-sunken/50">
							<td class="px-5 py-4">
								<div class="font-medium text-ink">{item.description}</div>
								<div class="mt-0.5 font-mono text-xs text-ink-4">{item.key}</div>
							</td>
							<td class="px-5 py-4">
								<div class="font-mono text-sm text-ink">{fmtValue(item)}</div>
							</td>
							<td class="px-5 py-4">
								<div class="font-mono text-sm text-ink-4">{fmtDefault(item)}</div>
							</td>
							<td class="px-5 py-4">
								<span class="rounded-md bg-surface-sunken px-2 py-0.5 text-xs font-medium text-ink-3">
									{typeLabels[item.type] ?? item.type}
								</span>
							</td>
							<td class="px-5 py-4 text-right">
								<div class="flex items-center justify-end gap-1">
									<button
										class="rounded-lg p-2 text-ink-4 transition-colors hover:bg-primary-soft hover:text-primary"
										onclick={() => openEdit(item)}
										title={m.admin_config_edit()}
									>
										<Pencil size={15} />
									</button>
									<button
										class="rounded-lg p-2 text-ink-4 transition-colors hover:bg-danger-soft hover:text-danger disabled:opacity-30"
										onclick={() => handleReset(item)}
										disabled={!isModified(item)}
										title={m.admin_config_reset_to_default()}
									>
										<RotateCcw size={15} />
									</button>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

<Dialog bind:open={showDialog} title={m.admin_config_edit_title()} onCancel={closeEdit} footer={false} size="sm">
	{#if editingItem}
		<div class="space-y-4">
			<div>
				<label class="block text-sm font-medium text-ink-3">{editingItem.description}</label>
				<p class="mt-0.5 font-mono text-xs text-ink-4">{editingItem.key}</p>
			</div>

			{#if editingItem.type === 'bool'}
				<Dropdown triggerClass="flex w-full items-center justify-between gap-2 rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink outline-none transition-colors hover:bg-surface-sunken data-[state=open]:bg-surface-sunken" contentClass="w-full" align="end">
					{#snippet trigger()}
						{editValue}
						<ChevronDown size={14} class="text-ink-4 shrink-0" />
					{/snippet}
					<DropdownBase.Item class={editValue === "true" ? "bg-primary-soft text-primary font-medium" : ""} onSelect={() => editValue = "true"}>
						<span class="flex items-center gap-2">
							{#if editValue === "true"}<Check size={14} />{/if}
							true
						</span>
					</DropdownBase.Item>
					<DropdownBase.Item class={editValue === "false" ? "bg-primary-soft text-primary font-medium" : ""} onSelect={() => editValue = "false"}>
						<span class="flex items-center gap-2">
							{#if editValue === "false"}<Check size={14} />{/if}
							false
						</span>
					</DropdownBase.Item>
				</Dropdown>
			{:else if editingItem.type === 'bytes'}
				<div>
					<div class="flex gap-2">
						<input
							type="number"
							class="flex-1 rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
							class:border-danger={!!editError}
							bind:value={editValue}
							min="0"
							step="any"
						/>
						<Dropdown triggerClass="flex items-center gap-2 rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink outline-none transition-colors hover:bg-surface-sunken data-[state=open]:bg-surface-sunken" contentClass="min-w-[80px]" align="end">
							{#snippet trigger()}
								{BYTE_UNITS.find(u => u.value === editUnit)?.label}
								<ChevronDown size={14} class="text-ink-4 shrink-0" />
							{/snippet}
							{#each BYTE_UNITS as u}
								<DropdownBase.Item
									class={editUnit === u.value ? "bg-primary-soft text-primary font-medium" : ""}
									onSelect={() => editUnit = u.value}
								>
									<span class="flex items-center gap-2">
										{#if editUnit === u.value}<Check size={14} />{/if}
										{u.label}
									</span>
								</DropdownBase.Item>
							{/each}
						</Dropdown>
					</div>
					<p class="mt-1.5 font-mono text-xs text-ink-4">
						= {bytesPreview.toLocaleString()} bytes
					</p>
				</div>
			{:else}
				<div>
					<input
						type={editingItem.type === 'number' ? 'number' : 'text'}
						class="w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
						class:border-danger={!!editError}
						bind:value={editValue}
						placeholder={String(editingItem.defaultValue)}
						min="0"
					/>
				</div>
			{/if}

			{#if editError}
				<p class="text-xs text-danger">{editError}</p>
			{/if}

			<div class="flex justify-end gap-2 pt-1">
				<button
					class="rounded-lg border border-line bg-surface px-4 py-2 text-sm text-ink transition-colors hover:bg-surface-sunken"
					onclick={closeEdit}
				>
					{m.admin_config_cancel()}
				</button>
				<button
					class="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-on transition-colors hover:bg-primary-hover disabled:opacity-50"
					onclick={saveEdit}
					disabled={saving}
				>
					{#if saving}
						<LoaderCircle size={15} class="animate-spin" />
					{/if}
					{m.admin_config_save()}
				</button>
			</div>
		</div>
	{/if}
</Dialog>
