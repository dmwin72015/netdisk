<script lang="ts">
	import { user, authReady } from '$lib/stores/auth';
	import { logout } from '$lib/api/auth';
	import { goto } from '$app/navigation';
	import { setUser } from '$lib/stores/auth';
	import { HardDrive, Folder, Star, Trash2, Film, Image, User, ChevronDown, LogOut, Settings, ListRestart, Shield } from '@lucide/svelte';
	import { Dropdown, DropdownBase } from '$lib/ui/dropdown';
	import LanguageDropdown from '$lib/components/LanguageDropdown.svelte';
	import { page } from '$app/state';
	import * as m from '$lib/paraglide/messages';

	async function handleLogout() {
		await logout();
		setUser(null);
		await goto('/login');
	}

	let currentPath = $derived(page.url.pathname);

	function isActive(path: string): boolean {
		if (path === '/') return currentPath === '/';
		return currentPath.startsWith(path);
	}

	let accountOpen = $state(false);
</script>

<header class="sticky top-0 z-40 border-b border-gray-200 bg-white/80 backdrop-blur-sm">
	<div class="mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
		<a href="/" class="flex items-center gap-2 font-semibold text-gray-900 transition-colors hover:text-blue-600">
			<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-blue-600 text-white">
				<HardDrive size={18} />
			</div>
			<span class="text-base">Netdisk</span>
		</a>
		<nav class="flex items-center gap-1 text-sm">
			{#if $authReady}
				{#if $user}
					<a
						href="/files/all"
						class="flex items-center gap-1.5 rounded-lg px-3 py-1.5 transition-colors {isActive('/files') ? 'bg-blue-50 text-blue-600 font-medium' : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'}"
					>
						<Folder size={15} /> {m.nav_files()}
					</a>
					<a
						href="/media"
						class="flex items-center gap-1.5 rounded-lg px-3 py-1.5 transition-colors {isActive('/media') ? 'bg-blue-50 text-blue-600 font-medium' : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'}"
					>
						<Film size={15} /> {m.nav_media()}
					</a>
					<a
						href="/photos"
						class="flex items-center gap-1.5 rounded-lg px-3 py-1.5 transition-colors {isActive('/photos') ? 'bg-blue-50 text-blue-600 font-medium' : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'}"
					>
						<Image size={15} /> {m.nav_photos()}
					</a>
					<div class="mx-1 h-5 w-px bg-gray-200"></div>

					<!-- Account dropdown -->
					<Dropdown
						bind:open={accountOpen}
						triggerClass="flex items-center gap-1 rounded-lg px-2.5 py-1.5 text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 data-[state=open]:bg-gray-100 data-[state=open]:text-gray-900"
						contentClass="min-w-[160px]"
						sideOffset={8}
						align="end"
					>
						{#snippet trigger()}
							{#if $user?.profile?.avatarUrl}
								<img src={$user.profile.avatarUrl} alt="" class="h-5 w-5 rounded-full object-cover" />
							{:else}
								<User size={15} />
							{/if}
							<span class="hidden sm:inline">{$user?.profile?.displayName || $user?.username || m.nav_account()}</span>
							<ChevronDown size={12} class="text-gray-400" />
						{/snippet}

						<DropdownBase.Item onSelect={() => goto('/files/starred')}>
							{#snippet icon()}
								<Star size={14} class="text-gray-400" />
							{/snippet}
							{m.nav_starred()}
						</DropdownBase.Item>
						<DropdownBase.Item onSelect={() => goto('/files/trash')}>
							{#snippet icon()}
								<Trash2 size={14} class="text-gray-400" />
							{/snippet}
							{m.nav_trash()}
						</DropdownBase.Item>
						<DropdownBase.Separator />
						{#if $user.role === 'admin'}
							<DropdownBase.Item onSelect={() => goto('/admin')}>
								{#snippet icon()}
									<Shield size={14} class="text-amber-500" />
								{/snippet}
								{m.admin_panel()}
							</DropdownBase.Item>
							<DropdownBase.Separator />
						{/if}
						<DropdownBase.Item onSelect={() => goto('/tasks')}>
							{#snippet icon()}
								<ListRestart size={14} class="text-gray-400" />
							{/snippet}
							{m.upload_title()}
						</DropdownBase.Item>
						<DropdownBase.Item onSelect={() => goto('/account')}>
							{#snippet icon()}
								<Settings size={14} class="text-gray-400" />
							{/snippet}
							{m.nav_account()}
						</DropdownBase.Item>
						<DropdownBase.Separator />
						<DropdownBase.Item variant="destructive" onSelect={handleLogout}>
							{#snippet icon()}
								<LogOut size={14} />
							{/snippet}
							{m.logout()}
						</DropdownBase.Item>
					</Dropdown>
				{:else}
					<a href="/login" class="rounded-lg px-3 py-1.5 text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900">{m.nav_login()}</a>
					<a href="/register" class="rounded-lg bg-blue-600 px-3 py-1.5 text-white transition-colors hover:bg-blue-700">{m.nav_register()}</a>
				{/if}
			{/if}

			<div class="mx-1 h-5 w-px bg-gray-200"></div>

			<!-- Language dropdown -->
			<LanguageDropdown />
		</nav>
	</div>
</header>
