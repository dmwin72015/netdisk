<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { user, authReady } from '$lib/stores/auth';
	import { logout } from '$lib/api/auth';
	import { setUser } from '$lib/stores/auth';
	import {
		LayoutDashboard,
		Users,
		FileText,
		HardDrive,
		Settings,
		Shield,
		LogOut,
		ChevronLeft,
		Menu,
		X,
		Activity,
		Trash2,
		ScrollText,
	} from '@lucide/svelte';
	import * as m from '$lib/paraglide/messages';
	import LanguageDropdown from '$lib/components/LanguageDropdown.svelte';

	let { children } = $props();

	let authorized = $state(false);
	let sidebarOpen = $state(true);
	let mobileSidebarOpen = $state(false);

	const currentPath = $derived($page.url.pathname);

	interface NavItem {
		href: string;
		match: string;
		label: string;
		icon: typeof LayoutDashboard;
		key: string;
	}

	const navItems: NavItem[] = [
		{ href: '/admin', match: '/admin', label: m.admin_dashboard(), icon: LayoutDashboard, key: 'dashboard' },
		{ href: '/admin/users', match: '/admin/users', label: m.admin_users(), icon: Users, key: 'users' },
		{ href: '/admin/files', match: '/admin/files', label: m.admin_files(), icon: FileText, key: 'files' },
		{ href: '/admin/storage', match: '/admin/storage', label: m.admin_storage(), icon: HardDrive, key: 'storage' },
		{ href: '/admin/settings', match: '/admin/settings', label: m.admin_settings(), icon: Settings, key: 'settings' },
	];

	// Utility function items rendered as a separate section
	const utilityItems: NavItem[] = [
		{ href: '/admin/cleanup', match: '/admin/cleanup', label: m.admin_cleanup(), icon: Trash2, key: 'cleanup' },
		{ href: '/admin/logs', match: '/admin/logs', label: m.admin_logs(), icon: ScrollText, key: 'logs' },
	];

	function isActive(item: NavItem) {
		if (item.match === '/admin') return currentPath === '/admin';
		return currentPath.startsWith(item.match);
	}

	onMount(() => {
		if (!browser) return;
		if (!$user) {
			void goto('/login');
			return;
		}
		if ($user.role !== 'admin') {
			void goto('/');
			return;
		}
		authorized = true;

		sidebarOpen = window.innerWidth >= 1024;
	});

	async function handleLogout() {
		await logout();
		setUser(null);
		await goto('/login');
	}

	function handleBackToApp() {
		void goto('/');
	}

	const pageTitle = $derived.by(() => {
		const active = navItems.find((item) => isActive(item));
		return active ? active.label : 'Admin';
	});
</script>

{#if $authReady && authorized}
	<div class="flex h-screen overflow-hidden bg-surface-muted text-ink">
		<!-- Mobile sidebar backdrop -->
		{#if mobileSidebarOpen}
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div
				class="fixed inset-0 z-40 bg-overlay lg:hidden"
				onclick={() => (mobileSidebarOpen = false)}
				onkeydown={(e) => e.key === 'Escape' && (mobileSidebarOpen = false)}
				role="presentation"
			></div>
		{/if}

		<!-- Sidebar -->
		<aside
			class="fixed inset-y-0 left-0 z-50 flex w-52 flex-col border-r border-line bg-surface transition-transform duration-200 lg:static lg:translate-x-0"
			class:translate-x-0={mobileSidebarOpen}
			class:-translate-x-full={!mobileSidebarOpen}
		>
			<!-- Logo area -->
			<div class="flex h-16 items-center justify-between gap-3 border-b border-line px-5">
				<a href="/admin" class="flex items-center gap-3">
					<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-admin">
						<Shield size={16} class="text-primary-on" />
					</div>
					<div>
						<span class="text-sm font-semibold text-ink">NetDisk</span>
						<span class="block text-[10px] font-medium uppercase tracking-wider text-admin">Admin</span>
					</div>
				</a>
				<button
					class="rounded-lg p-1.5 text-ink-4 hover:bg-surface-sunken hover:text-ink-3 lg:hidden"
					onclick={() => (mobileSidebarOpen = false)}
				>
					<X size={18} />
				</button>
			</div>

			<!-- Navigation -->
			<nav class="flex-1 overflow-y-auto px-3 py-4">
				<ul class="space-y-1">
					{#each navItems as item}
						{@const active = isActive(item)}
						{@const Icon = item.icon}
						<li>
							<a
								href={item.href}
								class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors"
								class:bg-primary-soft={active}
								class:text-primary={active}
								class:text-ink-3={!active}
								class:pointer-events-none={active}
							>
								<Icon size={18} />
								{item.label}
							</a>
						</li>
					{/each}
				</ul>

				{#if utilityItems.length > 0}
					<div class="mt-4 pt-3 border-t border-line">
						<span class="px-3 text-[10px] font-medium uppercase tracking-wider text-ink-4">{m.admin_utility_section()}</span>
						<ul class="mt-2 space-y-1">
							{#each utilityItems as item}
								{@const active = isActive(item)}
								{@const Icon = item.icon}
								<li>
									<a
										href={item.href}
										class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors"
										class:bg-primary-soft={active}
										class:text-primary={active}
										class:text-ink-3={!active}
										class:pointer-events-none={active}
									>
										<Icon size={18} />
										{item.label}
									</a>
								</li>
							{/each}
						</ul>
					</div>
				{/if}
			</nav>

			<!-- Bottom section -->
			<div class="border-t border-line p-3">
				<button
					onclick={handleBackToApp}
					class="flex w-full items-center gap-2 rounded-lg px-3 py-2.5 text-sm text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink"
				>
					<ChevronLeft size={16} />
					{m.admin_back_to_app()}
				</button>
				<button
					onclick={handleLogout}
					class="mt-1 flex w-full items-center gap-2 rounded-lg px-3 py-2.5 text-sm text-ink-4 transition-colors hover:bg-danger-soft hover:text-danger"
				>
					<LogOut size={16} />
					{m.logout()}
				</button>
			</div>
		</aside>

		<!-- Main content area -->
		<div class="flex min-w-0 flex-1 flex-col">
			<!-- Top bar -->
			<header class="bg-surface/85 sticky top-0 z-30 flex h-16 items-center gap-4 border-b border-line px-4 backdrop-blur lg:px-6">
				<button
					class="rounded-lg p-1.5 text-ink-4 hover:bg-surface-sunken hover:text-ink lg:hidden"
					onclick={() => (mobileSidebarOpen = true)}
				>
					<Menu size={20} />
				</button>

				<div class="flex items-center gap-2 text-sm">
					<span class="text-ink-4">Admin</span>
					<span class="text-ink-5">/</span>
					<span class="font-medium text-ink-2">{pageTitle}</span>
				</div>

				<div class="ml-auto flex items-center gap-3">
					{#if $user}
						<span class="hidden text-sm text-ink-3 sm:inline">{$user.username}</span>
					{/if}
					<span class="flex items-center gap-1.5 rounded-full bg-admin/10 px-3 py-1 text-xs font-medium text-admin">
						<Activity size={12} />
						{m.admin_mode()}
					</span>
					<LanguageDropdown
						triggerClass="flex items-center gap-1 rounded-lg px-2.5 h-8 text-sm text-ink-3 transition-colors hover:bg-surface-sunken hover:text-ink data-[state=open]:bg-surface-sunken data-[state=open]:text-ink"
					/>
				</div>
			</header>

			<!-- Page content -->
			<main class="flex-1 overflow-y-auto bg-surface-muted p-4 lg:p-6">
				{@render children()}
			</main>
		</div>
	</div>
{:else if $authReady && $user && $user.role !== 'admin'}
	<div class="flex h-screen items-center justify-center bg-surface-muted">
		<div class="text-center">
			<Shield size={48} class="mx-auto mb-4 text-danger" />
			<p class="text-lg text-ink-3">{m.admin_only()}</p>
		</div>
	</div>
{/if}
