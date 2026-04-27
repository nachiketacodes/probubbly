<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	let event: any = null;
	let ratios: any = null;
	let loading = true;
	let error = '';
	let user: any = null;
	let predictSide: string | null = null;
	let predictAmount = 10;
	let predicting = false;
	let predictError = '';
	let predictSuccess = '';
	let resolving = false;

	$: eventId = $page.params.id;

	onMount(async () => {
		const token = localStorage.getItem('token');
		const storedUser = localStorage.getItem('user');
		if (!token) {
			goto('/');
			return;
		}
		if (storedUser) {
			user = JSON.parse(storedUser);
		}
		await loadEvent();
	});

	async function loadEvent() {
		loading = true;
		try {
			const res = await api.getEvent(eventId);
			if (res.ok) {
				const data = await res.json();
				event = data.event;
				ratios = data.ratios;
			} else {
				error = 'Event not found';
			}
		} catch (err) {
			error = 'Connection failed';
		} finally {
			loading = false;
		}
	}

	async function handlePredict() {
		if (!predictSide) return;
		predictError = '';
		predictSuccess = '';

		if (predictAmount < 2) {
			predictError = 'Minimum prediction is 2 coins';
			return;
		}
		if (predictAmount > 80) {
			predictError = 'Maximum prediction is 80 coins';
			return;
		}
		if (user.balance < predictAmount) {
			predictError = 'Insufficient balance';
			return;
		}

		predicting = true;
		try {
			const res = await api.predict(eventId, predictSide, predictAmount);
			const data = await res.json();
			if (res.ok) {
				user.balance = data.new_balance;
				localStorage.setItem('user', JSON.stringify(user));
				ratios = data.new_ratios;
				await loadEvent();
				predictSuccess = `Successfully predicted ${predictSide.toUpperCase()} with ${predictAmount} coins at ${predictSide === 'yes' ? ratios.yes : ratios.no}x return!`;
				predictSide = null;
				predictAmount = 10;
			} else {
				predictError = data || 'Prediction failed';
			}
		} catch (err) {
			predictError = 'Connection failed';
		} finally {
			predicting = false;
		}
	}


	async function handleDelete() {
    if (!confirm('Delete this event? If open, all predictions will be refunded.')) return;
    resolving = true;
    try {
        const res = await api.deleteEvent(eventId);
        if (res.ok) {
            goto('/events');
        } else {
            const data = await res.json();
            alert(data.error || 'Failed to delete event');
        }
    } catch (err) {
        alert('Connection failed');
    } finally {
        resolving = false;
    }
}

	async function handleResolve(outcome: string) {
		if (!confirm(`Resolve this event as ${outcome.toUpperCase()}? This cannot be undone.`)) return;
		resolving = true;
		try {
			const res = await api.resolveEvent(eventId, outcome);
			if (res.ok) {
				await loadEvent();
			} else {
				alert('Failed to resolve event');
			}
		} catch (err) {
			alert('Connection failed');
		} finally {
			resolving = false;
		}
	}

	function formatDate(dateStr: string) {
		return new Date(dateStr).toLocaleDateString('en-GB', {
			day: 'numeric',
			month: 'long',
			year: 'numeric'
		});
	}

	$: estimatedPayout =
		predictSide && ratios
			? Math.floor(predictAmount * (predictSide === 'yes' ? ratios.yes : ratios.no) * 0.97)
			: 0;
</script>

<div class="min-h-screen bg-gray-50">
	<header class="bg-white border-b border-gray-200">
		<div class="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
			<div class="flex items-center gap-4">
				<a href="/events" class="text-gray-500 hover:text-gray-900 text-sm">← Events</a>
				<h1 class="text-lg font-bold text-gray-900">Event Detail</h1>
			</div>
			<div class="text-sm text-gray-500 font-mono">🪙 {user?.balance || 0}</div>
		</div>
	</header>

	<div class="max-w-3xl mx-auto px-6 py-8">
		{#if loading}
			<div class="text-center py-12"><p class="text-gray-500">Loading event...</p></div>
		{:else if error}
			<div class="p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">{error}</div>
		{:else if event}
			<div class="bg-white border border-gray-200 rounded-xl p-6 mb-6">
				<div class="flex items-start justify-between mb-3">
					<h2 class="text-xl font-bold text-gray-900 flex-1 pr-4">{event.title}</h2>
					<span
						class="px-3 py-1 rounded-full text-sm font-medium {event.status === 'open'
							? 'bg-blue-50 text-blue-700'
							: 'bg-gray-100 text-gray-600'}"
					>
						{event.status === 'open' ? 'Open' : 'Resolved'}
					</span>
				</div>
				{#if event.description}
					<p class="text-gray-600 text-sm mb-4 leading-relaxed">{event.description}</p>
				{/if}
				<div class="flex flex-wrap gap-4 text-xs text-gray-500">
					<span
						>📅 {formatDate(event.event_date)}{event.event_time
							? ` at ${event.event_time}`
							: ''}</span
					>
					<span>👤 Created by {event.creator_name}</span>
					<span>🪙 Total pool: {event.yes_coins + event.no_coins} coins</span>
				</div>
				{#if event.status === 'resolved'}
					<div
						class="mt-4 p-3 rounded-lg {event.outcome === 'yes'
							? 'bg-green-50 border border-green-200'
							: 'bg-red-50 border border-red-200'}"
					>
						<p class="font-semibold {event.outcome === 'yes' ? 'text-green-800' : 'text-red-800'}">
							Outcome: {event.outcome?.toUpperCase()}
						</p>
					</div>
				{/if}
			</div>

			{#if event.status === 'open' && ratios}
				<div class="bg-white border border-gray-200 rounded-xl p-6 mb-6">
					<h3 class="font-semibold text-gray-900 mb-4">Live Return Ratios</h3>
					<div class="grid grid-cols-2 gap-4 mb-4">
						<button
							on:click={() => (predictSide = predictSide === 'yes' ? null : 'yes')}
							class="p-4 rounded-xl border-2 transition-all text-left {predictSide === 'yes'
								? 'border-green-500 bg-green-50'
								: 'border-gray-200 bg-green-50 hover:border-green-300'}"
						>
							<p class="text-xs font-semibold text-green-700 uppercase tracking-wide mb-1">YES</p>
							<p class="text-3xl font-bold text-green-800 font-mono">{ratios.yes}x</p>
							<p class="text-xs text-green-600 mt-1">
								{ratios.yes_pct}% backing YES · {event.yes_coins} 🪙
							</p>
						</button>
						<button
							on:click={() => (predictSide = predictSide === 'no' ? null : 'no')}
							class="p-4 rounded-xl border-2 transition-all text-left {predictSide === 'no'
								? 'border-red-500 bg-red-50'
								: 'border-gray-200 bg-red-50 hover:border-red-300'}"
						>
							<p class="text-xs font-semibold text-red-700 uppercase tracking-wide mb-1">NO</p>
							<p class="text-3xl font-bold text-red-800 font-mono">{ratios.no}x</p>
							<p class="text-xs text-red-600 mt-1">
								{ratios.no_pct}% backing NO · {event.no_coins} 🪙
							</p>
						</button>
					</div>

					<div class="h-2 bg-gray-100 rounded-full overflow-hidden mb-1">
						<div class="h-full bg-green-500 rounded-full" style="width: {ratios.yes_pct}%"></div>
					</div>
					<div class="flex justify-between text-xs text-gray-400 mb-4">
						<span>YES {ratios.yes_pct}%</span>
						<span>{ratios.no_pct}% NO</span>
					</div>

					{#if predictSide}
						<div class="border border-gray-200 rounded-xl p-4 bg-gray-50">
							<p class="text-sm font-semibold text-gray-900 mb-3">
								Predict {predictSide.toUpperCase()} — your balance: {user?.balance} 🪙
							</p>

							{#if predictError}
								<div
									class="mb-3 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700"
								>
									{predictError}
								</div>
							{/if}

							{#if predictSuccess}
								<div
									class="mb-3 p-3 bg-green-50 border border-green-200 rounded-lg text-sm text-green-700"
								>
									{predictSuccess}
								</div>
							{/if}

							<div class="mb-3">
								<label
									class="block text-xs font-medium text-gray-700 uppercase tracking-wide mb-1.5"
									>Amount (2–80 coins)</label
								>
								<input
									type="number"
									min="2"
									max="80"
									bind:value={predictAmount}
									class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
								/>
							</div>

							<div
								class="flex items-center justify-between p-3 bg-white border border-gray-200 rounded-lg mb-3"
							>
								<span class="text-sm text-gray-600">Estimated return if correct:</span>
								<span class="text-base font-bold text-green-700 font-mono"
									>+{estimatedPayout} 🪙</span
								>
							</div>

							<div class="text-xs text-gray-400 mb-3">
								House cut: 3% applied · Ratio locked at time of prediction
							</div>

							<div class="flex gap-3">
								<button
									on:click={handlePredict}
									disabled={predicting}
									class="flex-1 py-2.5 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
								>
									{predicting ? 'Processing...' : `Confirm — predict ${predictSide.toUpperCase()}`}
								</button>
								<button
									on:click={() => {
										predictSide = null;
										predictError = '';
									}}
									class="px-4 py-2.5 border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50 transition-colors"
								>
									Cancel
								</button>
							</div>
						</div>
					{/if}
				</div>
			{/if}

			{#if user?.is_admin || event.creator_id === user?.id}
    {#if event.status === 'open' && user?.is_admin}
        <div class="bg-amber-50 border border-amber-200 rounded-xl p-5 mb-6">
            <p class="text-sm font-semibold text-amber-800 mb-3">⚙ Resolve Event</p>
            <p class="text-xs text-amber-700 mb-4">Once resolved, payouts are calculated and distributed automatically.</p>
            <div class="flex gap-3">
                <button
                    on:click={() => handleResolve('yes')}
                    disabled={resolving}
                    style="background-color: #16a34a; color: white;"
                    class="flex-1 py-2.5 text-sm font-medium rounded-lg disabled:opacity-50"
                >
                    Resolve as YES
                </button>
                <button
                    on:click={() => handleResolve('no')}
                    disabled={resolving}
                    style="background-color: #dc2626; color: white;"
                    class="flex-1 py-2.5 text-sm font-medium rounded-lg disabled:opacity-50"
                >
                    Resolve as NO
                </button>
            </div>
        </div>
    {/if}

    <div class="bg-red-50 border border-red-200 rounded-xl p-5 mb-6">
        <p class="text-sm font-semibold text-red-800 mb-1">🗑 Delete Event</p>
        <p class="text-xs text-red-600 mb-3">
            {event.status === 'open' ? 'All predictions will be refunded automatically.' : 'Admin only — resolved event deletion.'}
        </p>
        <button
            on:click={handleDelete}
            disabled={resolving}
            style="background-color: #dc2626; color: white;"
            class="w-full py-2.5 text-sm font-medium rounded-lg disabled:opacity-50"
        >
            Delete Event
        </button>
    </div>
{/if}
		{/if}
	</div>
</div>
