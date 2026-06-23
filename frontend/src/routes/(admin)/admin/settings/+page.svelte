<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { LoaderCircle, Server, Upload, Shield, Clock, HardDrive, Trash2 } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import { adminSystemInfo, type AdminSystemInfo } from '$lib/api/admin';
	import { fmtSize } from '$lib/utils/format';
	import * as m from '$lib/paraglide/messages';

	let info = $state<AdminSystemInfo | null>(null);
	let loading = $state(true);

	type SettingGroup = {
		title: string;
		icon: typeof Server;
		color: string;
		items: { label: string; value: string }[];
	};

	const groups = $derived<SettingGroup[]>(info ? [
		{
			title: 'Server',
			icon: Server,
			color: 'text-primary',
			items: [
				{ label: 'Port', value: String(info.server.port) },
			],
		},
		{
			title: 'Upload',
			icon: Upload,
			color: 'text-info',
			items: [
				{ label: 'Chunk Size', value: fmtSize(info.upload.chunkSize) },
				{ label: 'Max Upload Size', value: fmtSize(info.upload.maxUploadSize) },
			],
		},
		{
			title: 'Limits',
			icon: Shield,
			color: 'text-success',
			items: [
				{ label: 'Default Storage Quota', value: fmtSize(info.limits.defaultStorageQuota) },
				{ label: 'Avatar Max Size', value: fmtSize(info.limits.avatarMaxSize) },
			],
		},
		{
			title: 'Trash',
			icon: Trash2,
			color: 'text-danger',
			items: [
				{ label: 'Retention Days', value: `${info.trash.retentionDays} days` },
			],
		},
		{
			title: 'JWT',
			icon: Clock,
			color: 'text-warning',
			items: [
				{ label: 'Access Token TTL', value: `${info.jwt.accessTTLMin} min` },
				{ label: 'Refresh Token TTL', value: `${info.jwt.refreshTTLHour} hours` },
			],
		},
	] : []);

	onMount(() => {
		if (!browser) return;
		loadInfo();
	});

	async function loadInfo() {
		loading = true;
		try {
			info = await adminSystemInfo();
		} catch {
			toast.error(m.admin_load_failed());
		} finally {
			loading = false;
		}
	}
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-ink">{m.admin_settings()}</h1>
		<p class="mt-1 text-sm text-ink-4">System configuration and server settings</p>
	</div>

	{#if loading}
		<div class="flex justify-center py-16">
			<LoaderCircle size={24} class="animate-spin text-ink-4" />
		</div>
	{:else if groups.length > 0}
		<div class="grid gap-6 sm:grid-cols-2">
			{#each groups as group}
				{@const Icon = group.icon}
				<div class="rounded-xl border border-line bg-surface p-5">
					<div class="mb-4 flex items-center gap-2">
						<div class="rounded-lg bg-surface-sunken p-1.5">
							<Icon size={16} class={group.color} />
						</div>
						<h2 class="text-sm font-semibold text-ink-2">{group.title}</h2>
					</div>
					<dl class="space-y-2 text-sm">
						{#each group.items as item}
							<div class="flex items-center justify-between">
								<dt class="text-ink-3">{item.label}</dt>
								<dd class="font-mono text-xs text-ink">{item.value}</dd>
							</div>
						{/each}
					</dl>
				</div>
			{/each}
		</div>
	{/if}
</div>
