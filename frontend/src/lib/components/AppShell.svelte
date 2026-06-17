<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import {
		Bell,
		Film,
		Folder,
		HardDrive,
		Image as ImageIcon,
		ListRestart,
		LogOut,
			Search,
			Settings,
			Share2,
			Shield,
		Star,
		Trash2,
		User,
	} from '@lucide/svelte';
	import { logout } from '$lib/api/auth';
	import { setUser, user } from '$lib/stores/auth';
	import { Dropdown, DropdownBase } from '$lib/ui/dropdown';
	import LanguageDropdown from '$lib/components/LanguageDropdown.svelte';
	import FileSearchDialog from '$lib/components/FileSearchDialog.svelte';
	import { getFilesUrl } from '$lib/stores/last-section-url';
	import * as m from '$lib/paraglide/messages';

	let { children }: { children: import('svelte').Snippet } = $props();

	let currentPath = $derived(page.url.pathname);
	let accountOpen = $state(false);
	let searchOpen = $state(false);

	function handleSearchKeydown(e: KeyboardEvent) {
		if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
			e.preventDefault();
			searchOpen = true;
		}
	}

	type NavItem = { href: string; match: string; label: string; icon: typeof Folder };
	type MoreItem = { href: string; match: string; label: string; icon: typeof Star };

	const primaryNav: NavItem[] = [
		{ href: '/files/all', match: '/files', label: m.nav_files(), icon: Folder },
		{ href: '/photos', match: '/photos', label: m.nav_photos(), icon: ImageIcon },
		{ href: '/media', match: '/media', label: m.nav_media(), icon: Film },
	];

		const moreItems: MoreItem[] = [
			{ href: '/files/starred', match: '/files/starred', label: m.nav_starred(), icon: Star },
			{ href: '/files/trash', match: '/files/trash', label: m.nav_trash(), icon: Trash2 },
			{ href: '/shares', match: '/shares', label: '我的分享', icon: Share2 },
			{ href: '/tasks', match: '/tasks', label: m.upload_title(), icon: ListRestart },
		{ href: '/settings', match: '/settings', label: m.nav_settings(), icon: Settings },
		{ href: '/account', match: '/account', label: m.nav_account(), icon: User },
	];

	let pageTitle = $derived.by(() => {
		const all = [...primaryNav, ...moreItems];
		const active = all.find((item) => currentPath.startsWith(item.match));
		if (active) return active.label;
		if (currentPath.startsWith('/admin')) return m.admin_panel();
		return m.app_name();
	});

	function isActive(match: string) {
		return currentPath.startsWith(match);
	}

	// 点击「文件」tab 时跳到上次离开的子目录（仅普通点击，保留 ⌘/Ctrl+click 等原生行为）。
	function onFilesClick(event: MouseEvent) {
		if (event.defaultPrevented) return;
		if (event.button !== 0) return;
		if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
		const target = getFilesUrl();
		if (currentPath === target) return;
		event.preventDefault();
		void goto(target);
	}

	async function handleLogout() {
		await logout();
		setUser(null);
		await goto('/login');
	}
</script>

<svelte:window onkeydown={handleSearchKeydown} />

<div class="flex h-screen bg-[#f5f7fb] text-gray-900">
	<!-- Sidebar -->
	<aside class="sticky top-0 flex h-screen w-16 shrink-0 flex-col items-center border-r border-gray-100 bg-white py-4">
		<!-- Logo -->
		<a href="/" class="mb-6 flex flex-col items-center gap-1">
			<div class="flex h-9 w-9 items-center justify-center rounded-xl bg-blue-600 text-white shadow-sm shadow-blue-200">
				<HardDrive size={18} />
			</div>
		</a>

		<!-- Primary navigation (icon + text below) -->
		<nav class="flex w-full flex-col items-center gap-0.5">
			{#each primaryNav as item (item.href)}
				{@const Icon = item.icon}
				<a
					href={item.href}
					onclick={item.match === '/files' ? onFilesClick : undefined}
					class="relative flex w-13 flex-col items-center gap-1 rounded-xl py-2.5 text-xs transition-colors {isActive(item.match) ? 'bg-blue-50 font-medium text-blue-600' : 'text-gray-400 hover:bg-gray-50 hover:text-gray-600'}"
				>
					<Icon size={20} />
					<span>{item.label}</span>
				</a>
			{/each}
		</nav>

		<!-- Separator -->
		<div class="mt-auto mb-3 h-px w-8 rounded-full "></div>

		<!-- User menu at bottom -->
		<Dropdown
			bind:open={accountOpen}
			triggerClass="flex w-[52px] flex-col items-center gap-1 rounded-xl py-2.5 text-[14px] text-gray-400 transition-colors hover:bg-gray-50 hover:text-gray-600 data-[state=open]:bg-gray-50 data-[state=open]:text-gray-600"
			contentClass="min-w-[170px] mb-2"
			sideOffset={8}
			align="start"
		>
			{#snippet trigger()}
				{#if $user?.profile?.avatarUrl}
					<img src={$user.profile.avatarUrl} alt="" loading="lazy" class="h-7 w-7 rounded-full object-cover ring-2 ring-gray-100" />
				{:else}
					<span class="flex h-7 w-7 items-center justify-center rounded-full bg-blue-50 text-blue-600 ring-2 ring-gray-100"><User size={14} /></span>
				{/if}
				<span class="max-w-13 truncate text-xs">{$user?.profile?.displayName || $user?.username || ''}</span>
			{/snippet}

			<div class="px-2 pb-1 pt-1.5">
				<p class="truncate text-sm font-medium text-gray-900" title={$user?.profile?.displayName || $user?.username || ''}>
					{$user?.profile?.displayName || $user?.username || m.nav_account()}
				</p>
				<p class="truncate text-xs text-gray-400 max-w-46">{$user?.email || ''}</p>
			</div>

			{#each moreItems as item (item.href)}
				{@const Icon = item.icon}
				<DropdownBase.Item onSelect={() => goto(item.href)}>
					{#snippet icon()}<Icon size={14} class="text-gray-400" />{/snippet}
					{#snippet children()}{item.label}{/snippet}
				</DropdownBase.Item>
			{/each}

			{#if $user?.role === 'admin'}
				<DropdownBase.Item onSelect={() => goto('/admin')}>
					{#snippet icon()}<Shield size={14} class="text-amber-500" />{/snippet}
					{#snippet children()}{m.admin_panel()}{/snippet}
				</DropdownBase.Item>
			{/if}

			<DropdownBase.Separator />
			<DropdownBase.Item variant="destructive" onSelect={handleLogout}>
				{#snippet icon()}<LogOut size={14} />{/snippet}
				{#snippet children()}{m.logout()}{/snippet}
			</DropdownBase.Item>
		</Dropdown>
	</aside>

	<!-- Main area -->
	<div class="flex min-w-0 flex-1 flex-col">
		<header class="sticky top-0 z-30 flex h-16 items-center gap-4 border-b border-gray-200 bg-white/90 px-6 backdrop-blur">
			<div class="min-w-[8rem] text-lg font-semibold text-gray-950">{pageTitle}</div>
			<button type="button" onclick={() => { searchOpen = true; }}
				class="hidden h-9 max-w-xl flex-1 items-center gap-2 rounded-full border border-gray-200 bg-gray-50 px-4 text-sm text-gray-400 transition-colors hover:border-gray-300 hover:bg-gray-100 md:flex"
			>
				<Search size={16} />
				<span class="flex-1 text-left">{m.search_files()}…</span>
				<kbd class="rounded border border-gray-200 bg-white px-1.5 py-0.5 text-[10px] font-medium text-gray-500">⌘K</kbd>
			</button>
			<div class="ml-auto flex items-center gap-2">
				<LanguageDropdown
					triggerClass="flex items-center gap-1 rounded-lg px-2 py-1.5 text-sm text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 data-[state=open]:bg-gray-100 data-[state=open]:text-gray-700"
				/>
				<button type="button" class="flex h-9 w-9 items-center justify-center rounded-full text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-900" aria-label="通知">
					<Bell size={17} />
				</button>
			</div>
		</header>

		<main class="min-w-0 min-h-0 flex-1 overflow-y-auto px-6 py-5">
			{@render children()}
		</main>
	</div>
</div>

<FileSearchDialog bind:open={searchOpen} />
