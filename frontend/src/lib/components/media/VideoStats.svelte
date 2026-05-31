<script lang="ts">
	import Hls from 'hls.js';

	let {
		hls,
		video,
		visible = $bindable(false),
	}: {
		hls: Hls | null;
		video: HTMLVideoElement | undefined;
		visible?: boolean;
	} = $props();

	let rows = $state<{ label: string; value: string }[]>([]);
	let timer: ReturnType<typeof setInterval> | undefined;

	$effect(() => {
		if (visible && hls && video) {
			update();
			timer = setInterval(update, 500);
			return () => { if (timer) clearInterval(timer); };
		} else {
			if (timer) clearInterval(timer);
		}
	});

	function update() {
		if (!hls || !video) return;

		const r: { label: string; value: string }[] = [];

		// Viewport / Frames
		const dpr = window.devicePixelRatio || 1;
		const vw = window.innerWidth;
		const vh = window.innerHeight;
		let totalFrames = 0;
		let droppedFrames = 0;
		if ('getVideoPlaybackQuality' in video) {
			const q = (video as { getVideoPlaybackQuality(): { totalVideoFrames: number; droppedVideoFrames: number } }).getVideoPlaybackQuality();
			totalFrames = q.totalVideoFrames;
			droppedFrames = q.droppedVideoFrames;
		} else if ('webkitDroppedFrameCount' in video) {
			droppedFrames = (video as { webkitDroppedFrameCount: number }).webkitDroppedFrameCount;
		}
		r.push({ label: 'Viewport / Frames', value: `${vw}x${vh}*${dpr.toFixed(2)} / ${droppedFrames} dropped of ${totalFrames}` });

		// Current / Optimal Res
		const levelIdx = hls.currentLevel >= 0 ? hls.currentLevel : hls.loadLevel;
		const level = levelIdx >= 0 ? hls.levels[levelIdx] : null;
		const curW = video.videoWidth || level?.width || 0;
		const curH = video.videoHeight || level?.height || 0;
		const fps = level?.frameRate ? Math.round(level.frameRate) : 0;
		const resLabel = curW && curH ? `${curW}x${curH}${fps ? '@' + fps : ''}` : '-';
		r.push({ label: 'Current / Optimal Res', value: `${resLabel} / ${resLabel}` });

		// Volume / Normalized
		const vol = Math.round(video.volume * 100);
		r.push({ label: 'Volume / Normalized', value: `${vol}% / ${vol}%` });

		// Codecs
		const videoCodec = level?.codecSet || level?.videoCodec || '-';
		const audioTrack = hls.audioTracks[hls.audioTrack];
		const audioCodec = audioTrack?.codec || audioTrack?.name || '-';
		r.push({ label: 'Codecs', value: `${videoCodec} / ${audioCodec}` });

		// Color
		r.push({ label: 'Color', value: 'bt709 / bt709' });

		// Connection Speed
		if (hls.bandwidthEstimate) {
			r.push({ label: 'Connection Speed', value: `${(hls.bandwidthEstimate / 1000).toFixed(0)} Kbps` });
		}

		// Network Activity
		r.push({ label: 'Network Activity', value: '0 KB' });

		// Buffer Health
		if (video.buffered.length > 0) {
			const current = video.currentTime;
			let maxBuffer = 0;
			for (let i = 0; i < video.buffered.length; i++) {
				if (video.buffered.start(i) <= current && current <= video.buffered.end(i)) {
					maxBuffer = video.buffered.end(i) - current;
					break;
				}
			}
			r.push({ label: 'Buffer Health', value: `${maxBuffer.toFixed(2)} s` });
		} else {
			r.push({ label: 'Buffer Health', value: '0 s' });
		}

		// Date
		r.push({ label: 'Date', value: new Date().toString() });

		rows = r;
	}
</script>

{#if visible}
	<div class="absolute right-3 top-3 z-10 max-h-[80%] overflow-y-auto rounded-lg bg-black/85 p-3 text-xs text-white shadow-2xl backdrop-blur-sm">
		<div class="mb-2 flex items-center justify-between">
			<span class="font-medium text-gray-300">Stats for nerds</span>
			<button type="button" onclick={() => visible = false}
				class="text-gray-400 hover:text-white">
				✕
			</button>
		</div>
		<div class="space-y-1.5">
			{#each rows as row}
				<div>
					<div class="text-gray-400">{row.label}</div>
					<span class="font-mono text-white">{row.value}</span>
				</div>
			{/each}
		</div>
	</div>
{/if}
