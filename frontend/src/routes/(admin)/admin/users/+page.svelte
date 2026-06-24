<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import {
		Trash2, Eye, ChevronLeft, ChevronRight, LoaderCircle,
		Search, Plus, X, Check, ChevronDown, Pencil, HardDrive,
	} from '@lucide/svelte';
	import { Select } from 'bits-ui';
	import { toast } from 'svelte-sonner';
	import Dialog from '$lib/ui/dialog/Dialog.svelte';
	import DateRangePicker from '$lib/ui/date-range-picker/DateRangePicker.svelte';
	import { fmtSize } from '$lib/utils/format';
	import {
		adminListUsers,
		adminUpdateRole,
		adminUpdateStorageBase,
		adminDeleteUser,
		adminCreateUser,
		type AdminUser,
	} from '$lib/api/admin';
	import { confirmDelete, promptInput } from '$lib/dialog';
	import * as m from '$lib/paraglide/messages';

	const PAGE_SIZE = 20;

	let users = $state<AdminUser[]>([]);
	let total = $state(0);
	let offset = $state(0);
	let loading = $state(true);
	let searchQuery = $state('');
	let roleFilter = $state('');
	let sortBy = $state('-created_at');
	let dateRange = $state<{ start: Date | null; end: Date | null }>({ start: null, end: null });

	let currentPage = $derived(Math.floor(offset / PAGE_SIZE) + 1);
	let totalPages = $derived(Math.ceil(total / PAGE_SIZE));

	// Create user dialog
	let createOpen = $state(false);
	let createUsername = $state('');
	let createEmail = $state('');
	let createPassword = $state('');
	let createRole = $state('user');
	let creating = $state(false);

	// Edit user dialog
	let editOpen = $state(false);
	let editUser = $state<AdminUser | null>(null);
	let editRoleValue = $state('user');
	let editStorageBytes = $state('');
	let editing = $state(false);

	const roleOptions = [
		{ value: 'user', label: 'user' },
		{ value: 'admin', label: 'admin' },
	];

	const roleFilterOptions = [
		{ value: '', label: m.admin_all_roles() },
		{ value: 'user', label: 'user' },
		{ value: 'admin', label: 'admin' },
	];

	const sortOptions = [
		{ value: '-created_at', label: m.admin_newest_first() },
		{ value: 'created_at', label: m.admin_oldest_first() },
		{ value: 'username', label: m.admin_name_az() },
		{ value: '-username', label: m.admin_name_za() },
		{ value: 'storage', label: m.admin_most_storage() },
	];

	onMount(() => {
		if (!browser) return;
		loadUsers();
	});

	async function loadUsers() {
		loading = true;
		try {
			const res = await adminListUsers(PAGE_SIZE, offset, searchQuery || undefined, roleFilter || undefined, sortBy, dateRange.start || undefined, dateRange.end || undefined);
			users = res.items;
			total = res.total;
		} catch {
			toast.error(m.admin_load_failed());
		} finally {
			loading = false;
		}
	}

	function goPage(page: number) {
		offset = (page - 1) * PAGE_SIZE;
		loadUsers();
	}

	function handleSearch() {
		offset = 0;
		loadUsers();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') handleSearch();
	}

	// --- Create user dialog ---
	function openCreateDialog() {
		createUsername = '';
		createEmail = '';
		createPassword = '';
		createRole = 'user';
		createOpen = true;
	}

	async function handleCreateUser() {
		if (!createUsername.trim() || !createEmail.trim() || !createPassword.trim()) {
			toast.error(m.admin_please_fill());
			return;
		}
		creating = true;
		try {
			await adminCreateUser(createUsername.trim(), createEmail.trim(), createPassword, createRole);
			toast.success(m.admin_user_created());
			createOpen = false;
			await loadUsers();
		} catch {
			toast.error(m.admin_create_failed());
		} finally {
			creating = false;
		}
	}

	// --- Edit user dialog ---
	function openEditUser(u: AdminUser) {
		editUser = u;
		editRoleValue = u.role;
		editStorageBytes = String(u.baseBytes);
		editOpen = true;
	}

	async function handleEditUser() {
		if (!editUser) return;
		editing = true;
		try {
			const roleChanged = editRoleValue !== editUser.role;
			if (roleChanged) {
				await adminUpdateRole(editUser.id, editRoleValue);
				editUser.role = editRoleValue;
			}
			const bytes = parseInt(editStorageBytes, 10);
			if (!isNaN(bytes) && bytes >= 0 && bytes !== editUser.baseBytes) {
				const updated = await adminUpdateStorageBase(editUser.id, bytes);
				editUser.baseBytes = updated.baseBytes;
				editUser.totalBytes = updated.totalBytes;
			}
			if (roleChanged) {
				toast.success(m.admin_role_updated({ role: editRoleValue }));
			} else {
				toast.success(m.admin_storage_updated());
			}
			editOpen = false;
		} catch {
			toast.error(m.admin_update_failed());
		} finally {
			editing = false;
		}
	}

	// --- Delete user ---
	async function handleDelete(u: AdminUser) {
		const ok = await confirmDelete(m.admin_confirm_delete_user({ name: u.username }));
		if (!ok) return;
		try {
			await adminDeleteUser(u.id);
			users = users.filter((x) => x.id !== u.id);
			total--;
			toast.success(m.admin_user_deleted());
		} catch {
			toast.error(m.admin_delete_failed());
		}
	}

	function fmtDate(ts: number): string {
		return new Date(ts * 1000).toLocaleString();
	}
</script>

<div class="space-y-5">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-ink">{m.admin_user_management()}</h1>
			<p class="mt-0.5 text-sm text-ink-4">{m.admin_user_management_desc()}</p>
		</div>
		<button
			onclick={openCreateDialog}
			class="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-on transition-colors hover:bg-primary-hover"
		>
			<Plus size={16} />
			{m.admin_create_user()}
		</button>
	</div>

	<!-- Filters -->
	<div class="flex flex-wrap items-center gap-3">
		<div class="relative flex-1 min-w-[200px] max-w-xs">
			<Search size={16} class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-ink-4" />
			<input
				bind:value={searchQuery}
				onkeydown={handleKeydown}
				placeholder={m.admin_search_users()}
				class="w-full rounded-lg border border-line bg-surface py-2 pl-9 pr-3 text-sm text-ink placeholder:text-ink-4 focus:border-primary focus:outline-none"
			/>
		</div>
		<Select.Root type="single" bind:value={roleFilter} onValueChange={() => { offset = 0; loadUsers(); }}>
			<Select.Trigger class="flex items-center justify-between gap-2 rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink-3 min-w-[130px] data-[placeholder]:text-ink-4 focus:border-primary focus:outline-none">
				<Select.Value placeholder={m.admin_all_roles()} />
				<ChevronDown size={14} class="text-ink-4" />
			</Select.Trigger>
			<Select.Content class="z-50 overflow-hidden rounded-lg border border-line bg-surface p-1 shadow-pop" sideOffset={4} align="start">
				{#each roleFilterOptions as opt}
					<Select.Item value={opt.value} class="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm text-ink outline-none transition-colors hover:bg-primary/5 data-[highlighted]:bg-primary-soft data-[highlighted]:text-primary data-[state=checked]:bg-primary-soft data-[state=checked]:text-primary data-[state=checked]:font-semibold">
						{opt.label}
						{#if opt.value === roleFilter}<Check size={14} class="ml-auto" />{/if}
					</Select.Item>
				{/each}
			</Select.Content>
		</Select.Root>
		<Select.Root type="single" bind:value={sortBy} onValueChange={loadUsers}>
			<Select.Trigger class="flex items-center justify-between gap-2 rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink-3 min-w-[150px] data-[placeholder]:text-ink-4 focus:border-primary focus:outline-none">
				<Select.Value />
				<ChevronDown size={14} class="text-ink-4" />
			</Select.Trigger>
			<Select.Content class="z-50 overflow-hidden rounded-lg border border-line bg-surface p-1 shadow-pop" sideOffset={4} align="start">
				{#each sortOptions as opt}
					<Select.Item value={opt.value} class="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm text-ink outline-none transition-colors hover:bg-primary/5 data-[highlighted]:bg-primary-soft data-[highlighted]:text-primary data-[state=checked]:bg-primary-soft data-[state=checked]:text-primary data-[state=checked]:font-semibold">
						{opt.label}
						{#if opt.value === sortBy}<Check size={14} class="ml-auto" />{/if}
					</Select.Item>
				{/each}
			</Select.Content>
		</Select.Root>
			<DateRangePicker
				value={dateRange}
				onValueChange={(range) => {
					dateRange = range;
					offset = 0;
					loadUsers();
				}}
				class="min-w-[240px]"
			/>
		<button
			onclick={handleSearch}
			class="flex items-center gap-1.5 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-on transition-colors hover:bg-primary-hover"
		>
			<Search size={14} />
			{m.admin_search()}
		</button>
		<span class="ml-auto text-sm text-ink-4">{m.total_items({ total: String(total) })}</span>
	</div>

	<!-- User table -->
	<div class="overflow-hidden rounded-xl border border-line bg-surface">
		<table class="w-full table-fixed text-left text-sm">
		<colgroup>
		  <col style="width: 140px" />
		  <col style="width: 200px" />
		  <col style="width: 120px" />
		  <col style="width: 90px" />
		  <col style="width: 150px" />
		  <col style="width: 160px" />
		  <col style="width: 100px" />
		</colgroup>
			<thead class="border-b border-line bg-surface-sunken text-xs text-ink-3">
				<tr>
					<th class="px-4 py-3 font-medium">{m.username()}</th>
					<th class="px-4 py-3 font-medium">{m.email()}</th>
					<th class="px-4 py-3 font-medium">{m.admin_register_method()}</th>
					<th class="px-4 py-3 font-medium">{m.col_role()}</th>
					<th class="px-4 py-3 font-medium">{m.col_storage_limit()}</th>
					<th class="px-4 py-3 font-medium">{m.col_registered()}</th>
					<th class="px-4 py-3 font-medium">{m.col_actions()}</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-line-soft">
				{#if loading}
					<tr>
						<td colspan="7" class="px-4 py-12 text-center text-ink-4">
							<LoaderCircle size={20} class="mx-auto animate-spin" />
						</td>
					</tr>
				{:else if users.length === 0}
					<tr>
						<td colspan="7" class="px-4 py-12 text-center text-ink-4">{m.no_users()}</td>
					</tr>
				{:else}
					{#each users as u (u.id)}
						<tr class="transition-colors hover:bg-surface-sunken">
							<td class="truncate px-4 py-3 font-medium text-ink">{u.username}</td>
							<td class="truncate px-4 py-3 text-ink-3">{u.email}</td>
							<td class="truncate px-4 py-3 text-xs text-ink-4">{u.registerMethod}</td>
							<td class="px-4 py-3">
								<span
									class="rounded px-2 py-0.5 text-xs font-medium {u.role === 'admin'
										? 'bg-primary-soft text-primary'
										: 'bg-surface-sunken text-ink-3'}"
								>{u.role}</span>
							</td>
							<td class="px-4 py-3">
								{#if u.totalBytes > 0}
									{@const pct = Math.round((u.usedBytes / u.totalBytes) * 100)}
									<div class="mb-1 flex items-center justify-between text-xs">
										<span class="text-ink-3">{fmtSize(u.usedBytes)} / {fmtSize(u.totalBytes)}</span>
										<span class="text-ink-4">{pct}%</span>
									</div>
									<div class="h-1.5 overflow-hidden rounded-full bg-line">
										<div
											class="h-full rounded-full transition-all {pct > 90 ? 'bg-danger' : pct > 70 ? 'bg-warning' : 'bg-primary'}"
											style="width: {pct}%"
										></div>
									</div>
								{/if}
							</td>
							<td class="px-4 py-3 text-xs text-ink-4">{fmtDate(u.createdAt)}</td>
							<td class="px-4 py-3">
								<div class="flex items-center gap-1">
									<button
										class="rounded-lg p-1.5 text-ink-4 transition-colors hover:bg-primary-soft hover:text-primary"
										onclick={() => openEditUser(u)}
										title={m.admin_edit_role()}
									>
										<Pencil size={15} />
									</button>
									<button
										class="rounded-lg p-1.5 text-ink-4 transition-colors hover:bg-primary-soft hover:text-primary"
										onclick={() => goto(`/admin/users/${u.id}`)}
										title={m.admin_view_details()}
									>
										<Eye size={15} />
									</button>
									<button
										class="rounded-lg p-1.5 text-ink-4 transition-colors hover:bg-danger-soft hover:text-danger"
										onclick={() => handleDelete(u)}
										title={m.admin_delete()}
									>
										<Trash2 size={15} />
									</button>
								</div>
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>

	<!-- Pagination -->
	{#if totalPages > 1}
		<div class="flex items-center justify-center gap-2">
			<button
				class="rounded-lg border border-line px-3 py-1.5 text-sm text-ink-2 transition-colors hover:border-ink-2 hover:bg-surface hover:text-ink disabled:opacity-40"
				disabled={currentPage <= 1}
				onclick={() => goPage(currentPage - 1)}
			>
				<ChevronLeft size={14} />
			</button>
			<span class="text-sm text-ink-4">{currentPage} / {totalPages}</span>
			<button
				class="rounded-lg border border-line px-3 py-1.5 text-sm text-ink-2 transition-colors hover:border-ink-2 hover:bg-surface hover:text-ink disabled:opacity-40"
				disabled={currentPage >= totalPages}
				onclick={() => goPage(currentPage + 1)}
			>
				<ChevronRight size={14} />
			</button>
		</div>
	{/if}
</div>

<!-- Create User Dialog -->
<Dialog bind:open={createOpen} title={m.admin_create_user()} onConfirm={handleCreateUser} confirmText={creating ? m.admin_creating() : m.admin_create_user()} showCancel size="sm">
	<form autocomplete="off" onsubmit={(e) => e.preventDefault()}>
		<input type="text" name="prevent_autofill" autocomplete="username" value="" class="hidden" tabindex="-1" aria-hidden="true" />
		<input type="password" name="prevent_autofill_pw" autocomplete="current-password" value="" class="hidden" tabindex="-1" aria-hidden="true" />
		<div class="space-y-4">
			<div>
				<label class="mb-1 block text-xs text-ink-3">{m.username()}</label>
				<input
					bind:value={createUsername}
					name="new-admin-username"
					autocomplete="off"
					autocorrect="off"
					autocapitalize="off"
					spellcheck="false"
					placeholder="Username"
					class="w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink placeholder:text-ink-4 focus:border-primary focus:outline-none"
				/>
			</div>
			<div>
				<label class="mb-1 block text-xs text-ink-3">{m.email()}</label>
				<input
					bind:value={createEmail}
					type="email"
					name="new-admin-email"
					autocomplete="off"
					autocorrect="off"
					autocapitalize="off"
					spellcheck="false"
					placeholder="Email"
					class="w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink placeholder:text-ink-4 focus:border-primary focus:outline-none"
				/>
			</div>
			<div>
				<label class="mb-1 block text-xs text-ink-3">{m.password()}</label>
				<input
					bind:value={createPassword}
					type="password"
					name="new-admin-password"
					autocomplete="new-password"
					placeholder="Password"
					class="w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink placeholder:text-ink-4 focus:border-primary focus:outline-none"
				/>
			</div>
		<div>
			<label class="mb-1 block text-xs text-ink-3">{m.col_role()}</label>
			<Select.Root type="single" bind:value={createRole}>
				<Select.Trigger class="flex w-full items-center justify-between gap-1 rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink data-[placeholder]:text-ink-4">
					<Select.Value placeholder="user" />
					<ChevronDown size={14} class="text-ink-4" />
				</Select.Trigger>
				<Select.Portal>
				<Select.Content class="z-[60] overflow-hidden rounded-lg border border-line bg-surface p-1 shadow-pop" sideOffset={4} side="top" align="start">
				{#each roleOptions as opt}
					<Select.Item value={opt.value} class="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm text-ink outline-none transition-colors hover:bg-primary/5 data-[highlighted]:bg-primary-soft data-[highlighted]:text-primary data-[state=checked]:bg-primary-soft data-[state=checked]:text-primary data-[state=checked]:font-semibold">
						{opt.label}
						{#if opt.value === createRole}<Check size={14} class="ml-auto" />{/if}
					</Select.Item>
				{/each}
				</Select.Content>
				</Select.Portal>
			</Select.Root>
		</div>
	</form>
</Dialog>

<!-- Edit User Dialog -->
<Dialog bind:open={editOpen} title="Edit User" onConfirm={handleEditUser} confirmText={editing ? 'Saving...' : 'Save'} showCancel size="sm">
	{#if editUser}
		<p class="mb-4 text-sm font-medium text-ink">{editUser.username}</p>
		<div class="space-y-4">
			<div>
				<label class="mb-1 block text-xs text-ink-3">{m.col_role()}</label>
			<Select.Root type="single" bind:value={editRoleValue}>
				<Select.Trigger class="flex w-full items-center justify-between gap-1 rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink data-[placeholder]:text-ink-4">
					<Select.Value />
					<ChevronDown size={14} class="shrink-0 text-ink-4" />
				</Select.Trigger>
				<Select.Content class="z-50 w-[var(--bits-select-trigger-width)] overflow-hidden rounded-lg border border-line bg-surface p-1 shadow-pop" sideOffset={4} align="start">
					{#each roleOptions as opt}
						<Select.Item value={opt.value} class="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm text-ink outline-none transition-colors hover:bg-primary/5 data-[highlighted]:bg-primary-soft data-[highlighted]:text-primary data-[state=checked]:bg-primary-soft data-[state=checked]:text-primary data-[state=checked]:font-semibold">
							{opt.label}
							{#if opt.value === editRoleValue}<Check size={14} class="ml-auto" />{/if}
						</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>
			</div>
			<div>
				<label class="mb-1 block text-xs text-ink-3">{m.admin_set_base_storage()}</label>
				<input
					bind:value={editStorageBytes}
					type="number"
					class="w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink placeholder:text-ink-4 focus:border-primary focus:outline-none"
				/>
				<p class="mt-1 text-xs text-ink-4">{fmtSize(parseInt(editStorageBytes) || 0)}</p>
			</div>
		</div>
	{/if}
</Dialog>
