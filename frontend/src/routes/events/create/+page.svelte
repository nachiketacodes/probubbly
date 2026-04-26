<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let title = '';
	let description = '';
	let eventDate = '';
	let eventTime = '';
	let error = '';
	let loading = false;

	onMount(() => {
		const token = localStorage.getItem('token');
		if (!token) goto('/');
	});

	async function handleCreate() {
		error = '';
		if (!title.trim()) {
			error = 'Title is required';
			return;
		}
		if (!eventDate) {
			error = 'Date is required';
			return;
		}

		loading = true;
		try {
			const res = await api.createEvent({
				title: title.trim(),
				description: description.trim(),
				event_date: eventDate,
				event_time: eventTime || undefined
			});
			if (res.ok) {
				const event = await res.json();
				goto(`/events/${event.id}`);
			} else {
				error = 'Failed to create event';
			}
		} catch (err) {
			error = 'Connection failed';
		} finally {
			loading = false;
		}
	}
</script>

<div class="min-h-screen bg-gray-50">
	<header class="bg-white border-b border-gray-200">
		<div class="max-w-7xl mx-auto px-6 py-4 flex items-center gap-4">
			<a href="/events" class="text-gray-500 hover:text-gray-900 text-sm">← Events</a>
			<h1 class="text-lg font-bold text-gray-900">Create Event</h1>
		</div>
	</header>

	<div class="max-w-2xl mx-auto px-6 py-8">
		<div class="bg-white border border-gray-200 rounded-xl p-6">
			<h2 class="text-base font-semibold text-gray-900 mb-1">New Prediction Event</h2>
			<p class="text-sm text-gray-500 mb-6">
				Create a binary outcome event for others to forecast.
			</p>

			{#if error}
				<div class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700">
					{error}
				</div>
			{/if}

			<div class="space-y-4">
				<div>
					<label class="block text-xs font-medium text-gray-700 uppercase tracking-wide mb-1.5"
						>Event Title</label
					>
					<input
						type="text"
						bind:value={title}
						placeholder="Will X happen by date Y?"
						class="w-full px-3 py-2.5 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
					/>
					<p class="text-xs text-gray-400 mt-1">Frame as a clear yes or no question.</p>
				</div>

				<div>
					<label class="block text-xs font-medium text-gray-700 uppercase tracking-wide mb-1.5"
						>Description</label
					>
					<textarea
						bind:value={description}
						placeholder="Describe the event and how the outcome will be determined..."
						rows="3"
						class="w-full px-3 py-2.5 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
					></textarea>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-xs font-medium text-gray-700 uppercase tracking-wide mb-1.5"
							>Date of Occurrence</label
						>
						<input
							type="date"
							bind:value={eventDate}
							class="w-full px-3 py-2.5 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
						/>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-700 uppercase tracking-wide mb-1.5"
							>Time (optional)</label
						>
						<input
							type="time"
							bind:value={eventTime}
							class="w-full px-3 py-2.5 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
						/>
					</div>
				</div>

				<div class="p-3 bg-blue-50 border border-blue-200 rounded-lg text-xs text-blue-700">
					Events have binary outcomes: <strong>YES</strong> or <strong>NO</strong>. Return ratios
					adjust dynamically as predictions come in. House takes a 3% cut on all winnings.
				</div>

				<button
					on:click={handleCreate}
					disabled={loading}
					class="w-full py-2.5 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
				>
					{loading ? 'Creating...' : 'Create Event'}
				</button>
			</div>
		</div>
	</div>
</div>
