<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { user } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { Trash2, Users, ChevronLeft, ChevronRight, Loader2 } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import { fmtSize } from '$lib/utils/format';
	import {
		adminListUsers,
		adminUpdateRole,
		adminUpdateStorageBase,
		adminDeleteUser,
		type AdminUser
	} from '$lib/api/admin';
	import { confirmDelete, promptInput } from '$lib/dialog';
	import * as m from '$lib/paraglide/messages';

	const PAGE_SIZE = 20;

	let users = $state<AdminUser[]>([]);
	let total = $state(0);
	let offset = $state(0);
	let loading = $state(true);

	let currentPage = $derived(Math.floor(offset / PAGE_SIZE) + 1);
	let totalPages = $derived(Math.ceil(total / PAGE_SIZE));

	onMount(() => {
		if (!browser) return;
		if (!$user || $user.role !== 'admin') {
			void goto('/');
			return;
		}
		loadUsers();
	});

	async function loadUsers() {
		loading = true;
		try {
			const res = await adminListUsers(PAGE_SIZE, offset);
			users = res.items;
			total = res.total;
		} catch {
			toast.error(m.load_failed());
		} finally {
			loading = false;
		}
	}

	function goPage(page: number) {
		offset = (page - 1) * PAGE_SIZE;
		loadUsers();
	}

	async function handleRoleChange(u: AdminUser, newRole: string) {
		if (newRole === u.role) return;
		try {
			await adminUpdateRole(u.id, newRole);
			u.role = newRole;
		} catch {
			toast.error(m.load_failed());
		}
	}

	async function handleSetBaseStorage(u: AdminUser) {
		const val = await promptInput(m.set_base_storage(), m.enter_storage_bytes(), String(u.baseBytes));
		if (val === null) return;
		const bytes = parseInt(val, 10);
		if (isNaN(bytes) || bytes < 0) return;
		try {
			const updated = await adminUpdateStorageBase(u.id, bytes);
			u.baseBytes = updated.baseBytes;
			u.totalBytes = updated.totalBytes;
		} catch {
			toast.error(m.load_failed());
		}
	}

	async function handleDelete(u: AdminUser) {
		const ok = await confirmDelete(m.confirm_delete_user({ name: u.username }));
		if (!ok) return;
		try {
			await adminDeleteUser(u.id);
			users = users.filter((x) => x.id !== u.id);
			total--;
		} catch {
			toast.error(m.delete_failed());
		}
	}

	function fmtDate(ts: number): string {
		return new Date(ts * 1000).toLocaleDateString();
	}
</script>

<div class="space-y-4">
	<div class="flex items-center gap-2">
		<Users size={20} class="text-slate-600" />
		<h1 class="text-xl font-semibold">{m.user_management()}</h1>
		<span class="ml-auto text-sm text-slate-400">{m.total_items({ total: String(total) })}</span>
	</div>

	<div class="overflow-hidden rounded-lg border bg-white">
		<table class="w-full text-left text-sm">
			<thead class="border-b bg-slate-50 text-xs text-slate-500">
				<tr>
					<th class="px-4 py-3 font-medium">{m.username()}</th>
					<th class="px-4 py-3 font-medium">{m.email()}</th>
					<th class="px-4 py-3 font-medium">{m.col_role()}</th>
					<th class="px-4 py-3 font-medium">{m.col_storage_limit()}</th>
					<th class="px-4 py-3 font-medium">{m.col_registered()}</th>
					<th class="px-4 py-3 font-medium">{m.col_actions()}</th>
				</tr>
			</thead>
			<tbody class="divide-y">
				{#if loading}
					<tr>
						<td colspan="6" class="px-4 py-10 text-center text-slate-400">
							<Loader2 size={20} class="mx-auto animate-spin" />
						</td>
					</tr>
				{:else if users.length === 0}
					<tr>
						<td colspan="6" class="px-4 py-10 text-center text-slate-400">{m.no_users()}</td>
					</tr>
				{:else}
					{#each users as u (u.id)}
						<tr class="hover:bg-slate-50 transition-colors">
							<td class="px-4 py-3 font-medium text-slate-900">{u.username}</td>
							<td class="px-4 py-3 text-slate-600">{u.email}</td>
							<td class="px-4 py-3">
								<select
									class="rounded border border-slate-200 bg-white px-2 py-1 text-xs focus:border-blue-400 focus:outline-none"
									value={u.role}
									onchange={(e) => handleRoleChange(u, (e.target as HTMLSelectElement).value)}
								>
									<option value="user">user</option>
									<option value="admin">admin</option>
								</select>
							</td>
							<td class="px-4 py-3">
								<button
									class="text-left underline decoration-dashed underline-offset-2 hover:text-blue-600 transition-colors"
									onclick={() => handleSetBaseStorage(u)}
									title={m.set_base_storage()}
								>
									<div class="text-xs text-slate-500">{fmtSize(u.usedBytes)} / {fmtSize(u.totalBytes)}</div>
									<div class="text-xs text-slate-400">{m.storage_base()}: {fmtSize(u.baseBytes)}</div>
								</button>
							</td>
							<td class="px-4 py-3 text-slate-500">{fmtDate(u.createdAt)}</td>
							<td class="px-4 py-3">
								<button
									class="rounded p-1.5 text-slate-400 transition-colors hover:bg-red-50 hover:text-red-600"
									onclick={() => handleDelete(u)}
									title={m.delete_btn()}
								>
									<Trash2 size={15} />
								</button>
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>

	{#if totalPages > 1}
		<div class="flex items-center justify-center gap-2">
			<button
				class="rounded-lg border px-3 py-1.5 text-sm transition-colors hover:bg-slate-50 disabled:opacity-40"
				disabled={currentPage <= 1}
				onclick={() => goPage(currentPage - 1)}
			>
				<ChevronLeft size={14} />
			</button>
			<span class="text-sm text-slate-500">{currentPage} / {totalPages}</span>
			<button
				class="rounded-lg border px-3 py-1.5 text-sm transition-colors hover:bg-slate-50 disabled:opacity-40"
				disabled={currentPage >= totalPages}
				onclick={() => goPage(currentPage + 1)}
			>
				<ChevronRight size={14} />
			</button>
		</div>
	{/if}
</div>
