<script lang="ts">
	import { user, authReady } from '$lib/stores/auth';
	import { logout } from '$lib/api/auth';
	import { goto } from '$app/navigation';
	import { setUser } from '$lib/stores/auth';
	import { HardDrive, Upload, Star, Trash2, Film, User, Globe, ChevronDown, LogOut, Settings } from '@lucide/svelte';
	import { DropdownMenu } from 'bits-ui';
	import { page } from '$app/state';
	import * as m from '$lib/paraglide/messages';
	import { getLocale, setLocale, locales } from '$lib/paraglide/runtime';

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

	const localeLabels: Record<string, string> = { zh: '中文', en: 'English' };
</script>

<header class="sticky top-0 z-40 border-b border-gray-200 bg-white/80 backdrop-blur-sm">
	<div class="mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
		<a href="/files" class="flex items-center gap-2 font-semibold text-gray-900 transition-colors hover:text-blue-600">
			<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-blue-600 text-white">
				<HardDrive size={18} />
			</div>
			<span class="text-base">Netdisk</span>
		</a>
		<nav class="flex items-center gap-1 text-sm">
			{#if $authReady}
				{#if $user}
					<a
						href="/files"
						class="flex items-center gap-1.5 rounded-lg px-3 py-1.5 transition-colors {isActive('/files') && !isActive('/files/trash') && !isActive('/files/starred') ? 'bg-blue-50 text-blue-600 font-medium' : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'}"
					>
						<Upload size={15} /> Files
					</a>
					<a
						href="/files/starred"
						class="flex items-center gap-1.5 rounded-lg px-3 py-1.5 transition-colors {isActive('/files/starred') ? 'bg-blue-50 text-blue-600 font-medium' : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'}"
					>
						<Star size={15} /> Starred
					</a>
					<a
						href="/files/trash"
						class="flex items-center gap-1.5 rounded-lg px-3 py-1.5 transition-colors {isActive('/files/trash') ? 'bg-blue-50 text-blue-600 font-medium' : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'}"
					>
						<Trash2 size={15} /> Trash
					</a>
					<a
						href="/media"
						class="flex items-center gap-1.5 rounded-lg px-3 py-1.5 transition-colors {isActive('/media') ? 'bg-blue-50 text-blue-600 font-medium' : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'}"
					>
						<Film size={15} /> Media
					</a>
					<div class="mx-1 h-5 w-px bg-gray-200"></div>

					<!-- Account dropdown -->
					<DropdownMenu.Root>
						<DropdownMenu.Trigger
							class="flex items-center gap-1 rounded-lg px-2.5 py-1.5 text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 data-[state=open]:bg-gray-100 data-[state=open]:text-gray-900"
						>
							{#if $user?.profile?.avatar_url}
								<img src={$user.profile.avatar_url} alt="" class="h-5 w-5 rounded-full object-cover" />
							{:else}
								<User size={15} />
							{/if}
							<span class="hidden sm:inline">{$user?.profile?.display_name || $user?.username || 'Account'}</span>
							<ChevronDown size={12} class="text-gray-400" />
						</DropdownMenu.Trigger>
						<DropdownMenu.Portal>
							<DropdownMenu.Content
								class="z-50 min-w-[160px] rounded-xl border border-gray-100 bg-white p-1.5 shadow-lg data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0 data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95 data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2"
								sideOffset={8}
								align="end"
							>
								<DropdownMenu.Item
									class="flex cursor-pointer items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-gray-700 outline-none transition-colors hover:bg-gray-50 focus:bg-gray-50"
									onclick={() => goto('/account')}
								>
									<Settings size={14} class="text-gray-400" />
									Account
								</DropdownMenu.Item>
								<DropdownMenu.Separator class="my-1 h-px bg-gray-100" />
								<DropdownMenu.Item
									class="flex cursor-pointer items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-red-600 outline-none transition-colors hover:bg-red-50 focus:bg-red-50"
									onclick={handleLogout}
								>
									<LogOut size={14} />
									Logout
								</DropdownMenu.Item>
							</DropdownMenu.Content>
						</DropdownMenu.Portal>
					</DropdownMenu.Root>
				{:else}
					<a href="/login" class="rounded-lg px-3 py-1.5 text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900">Login</a>
					<a href="/register" class="rounded-lg bg-blue-600 px-3 py-1.5 text-white transition-colors hover:bg-blue-700">Register</a>
				{/if}
			{/if}

			<div class="mx-1 h-5 w-px bg-gray-200"></div>

			<!-- Language dropdown -->
			<DropdownMenu.Root>
				<DropdownMenu.Trigger
					class="flex items-center gap-1 rounded-lg px-2 py-1.5 text-xs text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 data-[state=open]:bg-gray-100 data-[state=open]:text-gray-700"
				>
					<Globe size={14} />
					<span>{localeLabels[getLocale()] ?? getLocale()}</span>
					<ChevronDown size={10} class="text-gray-400" />
				</DropdownMenu.Trigger>
				<DropdownMenu.Portal>
					<DropdownMenu.Content
						class="z-50 min-w-[120px] rounded-xl border border-gray-100 bg-white p-1.5 shadow-lg data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0 data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95 data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2"
						sideOffset={8}
						align="end"
					>
						{#each locales as locale}
							<DropdownMenu.Item
								class="flex cursor-pointer items-center gap-2 rounded-lg px-3 py-2 text-sm outline-none transition-colors {locale === getLocale() ? 'bg-blue-50 text-blue-600 font-medium' : 'text-gray-700 hover:bg-gray-50 focus:bg-gray-50'}"
								onclick={() => setLocale(locale)}
							>
								{localeLabels[locale] ?? locale}
							</DropdownMenu.Item>
						{/each}
					</DropdownMenu.Content>
				</DropdownMenu.Portal>
			</DropdownMenu.Root>
		</nav>
	</div>
</header>
