<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let events: any[] = [];
	let loading = true;
	let filter = 'open';
	let error = '';

	onMount(async () => {
		const token = localStorage.getItem('token');
		if (!token) {
			goto('/');
			return;
		}
		await loadEvents();
	});

	async function loadEvents() {
		loading = true;
		try {
			const res = await api.getEvents(filter);
			if (res.ok) {
				events = await res.json();
			} else {
				error = 'Failed to load events';
			}
		} catch (err) {
			error = 'Connection failed';
		} finally {
			loading = false;
		}
	}

	function formatDate(dateStr: string) {
		return new Date(dateStr).toLocaleDateString('en-GB', {
			day: 'numeric',
			month: 'short',
			year: 'numeric'
		});
	}

	function computeRatios(yesCoins: number, noCoins: number) {
		const total = yesCoins + noCoins;
		if (total === 0) return { yes: 1.94, no: 1.94, yesPct: 50, noPct: 50 };
		const yesFrac = yesCoins / total;
		const noFrac = noCoins / total;
		const base = 1.05;
		let yesRatio = Math.min(Math.max(((base + noFrac * 3.5) / yesFrac) * 0.97, 1.02), 9.6);
		let noRatio = Math.min(Math.max(((base + yesFrac * 3.5) / noFrac) * 0.97, 1.02), 9.6);
		return {
			yes: Math.round(yesRatio * 100) / 100,
			no: Math.round(noRatio * 100) / 100,
			yesPct: Math.round(yesFrac * 100),
			noPct: Math.round(noFrac * 100)
		};
	}
</script>

<div class="min-h-screen bg-gray-50">
	<header class="bg-white border-b border-gray-200">
		<div class="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
			<div class="flex items-center gap-4">
				<a href="/dashboard" class="text-gray-500 hover:text-gray-900 text-sm">← Dashboard</a>
				<h1 class="text-lg font-bold text-gray-900">Events</h1>
			</div>
			<a
				href="/events/create"
				class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
				>+ Create Event</a
			>
		</div>
	</header>

	<div class="max-w-4xl mx-auto px-6 py-8">
		<div class="flex border-b border-gray-200 mb-6">
			{#each [['open', 'Open'], ['resolved', 'Resolved'], ['', 'All']] as [f, label]}
				<button
					on:click={() => {
						filter = f;
						loadEvents();
					}}
					class="px-4 py-2 text-sm font-medium border-b-2 transition-colors {filter === f
						? 'border-blue-600 text-blue-600'
						: 'border-transparent text-gray-500 hover:text-gray-900'}">{label}</button
				>
			{/each}
		</div>

		{#if loading}
			<div class="text-center py-12"><p class="text-gray-500">Loading events...</p></div>
		{:else if error}
			<div class="p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">{error}</div>
		{:else if events.length === 0}
			<div class="text-center py-16">
				<p class="text-2xl mb-3">📋</p>
				<p class="text-gray-700 font-medium">No {filter} events</p>
				<p class="text-gray-500 text-sm mt-1">Create one to get started</p>
				<a
					href="/events/create"
					class="inline-block mt-4 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
					>Create Event</a
				>
			</div>
		{:else}
			<div class="space-y-4">
				{#each events as event}
					<div
						on:click={() => goto(`/events/${event.id}`)}
						on:keypress={() => goto(`/events/${event.id}`)}
						role="button"
						tabindex="0"
						class="bg-white border border-gray-200 rounded-xl p-5 hover:border-gray-300 hover:shadow-sm transition-all cursor-pointer"
					>
						<div class="flex items-start justify-between mb-2">
							<h2 class="font-semibold text-gray-900 text-base flex-1 pr-4">{event.title}</h2>
							<span
								class="px-2 py-1 rounded-full text-xs font-medium flex-shrink-0 {event.status ===
								'open'
									? 'bg-blue-50 text-blue-700'
									: 'bg-gray-100 text-gray-600'}"
							>
								{event.status === 'open' ? 'Open' : 'Resolved'}
							</span>
						</div>
						<p class="text-xs text-gray-500 mb-3">
							{formatDate(event.event_date)} · Created by {event.creator_name}
						</p>
						{#if event.status === 'open'}
							{@const ratios = computeRatios(event.yes_coins, event.no_coins)}
							<div class="grid grid-cols-2 gap-3 mb-3">
								<div class="bg-green-50 rounded-lg p-3">
									<p class="text-xs text-green-700 font-medium uppercase tracking-wide">
										YES return
									</p>
									<p class="text-xl font-bold text-green-800 font-mono">{ratios.yes}x</p>
									<p class="text-xs text-green-600 mt-1">{event.yes_coins} 🪙 · {ratios.yesPct}%</p>
								</div>
								<div class="bg-red-50 rounded-lg p-3">
									<p class="text-xs text-red-700 font-medium uppercase tracking-wide">NO return</p>
									<p class="text-xl font-bold text-red-800 font-mono">{ratios.no}x</p>
									<p class="text-xs text-red-600 mt-1">{event.no_coins} 🪙 · {ratios.noPct}%</p>
								</div>
							</div>
							<div class="h-1.5 bg-gray-100 rounded-full overflow-hidden">
								<div class="h-full bg-green-500 rounded-full" style="width: {ratios.yesPct}%"></div>
							</div>
							<div class="flex justify-between text-xs text-gray-400 mt-1">
								<span>YES {ratios.yesPct}%</span>
								<span>{ratios.noPct}% NO</span>
							</div>
						{:else}
							<div class="flex items-center gap-2">
								<span class="text-sm text-gray-500">Outcome:</span>
								<span
									class="px-3 py-1 rounded-full text-sm font-semibold {event.outcome === 'yes'
										? 'bg-green-100 text-green-800'
										: 'bg-red-100 text-red-800'}">{event.outcome?.toUpperCase()}</span
								>
							</div>
						{/if}
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>
