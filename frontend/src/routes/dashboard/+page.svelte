<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let user: any = null;
	let events: any[] = [];
	let loading = true;
	let canBorrow = false;
	let borrowing = false;
	let error = '';

	onMount(async () => {
		// Read directly from localStorage — simple and reliable
		const token = localStorage.getItem('token');
		const storedUser = localStorage.getItem('user');

		if (!token || !storedUser) {
			goto('/');
			return;
		}

		user = JSON.parse(storedUser);

		// Now load dashboard data
		try {
			const eventsRes = await api.getEvents('open');
			if (eventsRes.ok) {
				events = await eventsRes.json();
			}

			const walletRes = await api.getWallet();
			if (walletRes.ok) {
				const walletData = await walletRes.json();
				// Update balance from server (source of truth)
				user.balance = walletData.user.balance;
				// Check borrow eligibility
				const today = new Date().toISOString().split('T')[0];
				const lastBorrowDate = walletData.user.last_borrow
					? walletData.user.last_borrow.split('T')[0]
					: null;
				canBorrow = !lastBorrowDate || lastBorrowDate !== today;
			}
		} catch (err) {
			error = 'Failed to load dashboard data';
			console.error(err);
		} finally {
			loading = false;
		}
	});

	async function handleBorrow() {
		borrowing = true;
		try {
			const res = await api.borrowCoins();
			if (res.ok) {
				const data = await res.json();
				user.balance = data.new_balance;
				user = { ...user };
				canBorrow = false;
				localStorage.setItem('user', JSON.stringify(user));
			} else {
				const data = await res.json();
				alert(data || 'Borrow failed');
			}
		} catch (err) {
			alert('Connection failed');
		} finally {
			borrowing = false;
		}
	}

	function logout() {
		localStorage.removeItem('token');
		localStorage.removeItem('user');
		goto('/');
	}
</script>

<div class="min-h-screen bg-gray-50">
	<header class="bg-white border-b border-gray-200">
		<div class="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
			<div>
				<h1 class="text-xl font-bold text-gray-900">Probubbly</h1>
				<p class="text-xs text-gray-500">Educational Probability Platform</p>
			</div>
			<div class="flex items-center gap-4">
				<div class="text-right">
					<p class="text-sm font-medium text-gray-900">{user?.username || ''}</p>
					<p class="text-xs text-gray-500 font-mono">🪙 {user?.balance?.toFixed(4) || '0.0000'}</p>
				</div>
				<button
					on:click={logout}
					class="px-4 py-2 text-sm text-gray-600 hover:text-gray-900 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
				>
					Sign out
				</button>
			</div>
		</div>
	</header>

	<div class="max-w-7xl mx-auto px-6 py-8">
		{#if loading}
			<div class="text-center py-12">
				<p class="text-gray-500">Loading your dashboard...</p>
			</div>
		{:else if error}
			<div class="p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
				{error}
			</div>
		{:else}
			<div class="grid grid-cols-4 gap-4 mb-8">
				<div class="bg-gray-100 rounded-lg p-4">
					<p class="text-xs text-gray-500 uppercase tracking-wide font-medium">Balance</p>
					<p class="text-2xl font-semibold text-gray-900 mt-1 font-mono">
						🪙 {user?.balance?.toFixed(4) || '0.0000'}
					</p>
					<p class="text-xs text-gray-500 mt-1">coins available</p>
				</div>

				<div class="bg-gray-100 rounded-lg p-4">
					<p class="text-xs text-gray-500 uppercase tracking-wide font-medium">Open Events</p>
					<p class="text-2xl font-semibold text-gray-900 mt-1">{events.length}</p>
					<p class="text-xs text-gray-500 mt-1">available to predict</p>
				</div>

				<div class="bg-gray-100 rounded-lg p-4">
					<p class="text-xs text-gray-500 uppercase tracking-wide font-medium">Daily Loan</p>
					<div class="mt-2">
						{#if canBorrow}
							<button
								on:click={handleBorrow}
								disabled={borrowing}
								class="px-3 py-1.5 bg-blue-600 text-white text-xs font-medium rounded hover:bg-blue-700 transition-colors disabled:opacity-50"
							>
								{borrowing ? 'Borrowing...' : 'Borrow 400'}
							</button>
						{:else}
							<span class="text-sm text-gray-500">Used today</span>
						{/if}
					</div>
					<p class="text-xs text-gray-500 mt-1">400 coins · resets daily</p>
				</div>

				<div class="bg-gray-100 rounded-lg p-4">
					<p class="text-xs text-gray-500 uppercase tracking-wide font-medium">Admin</p>
					{#if user?.is_admin}
						<a
							href="/admin"
							class="inline-block mt-2 px-3 py-1.5 bg-purple-600 text-white text-xs font-medium rounded hover:bg-purple-700 transition-colors"
							>Admin Panel</a
						>
					{:else}
						<p class="text-sm text-gray-500 mt-2">Not admin</p>
					{/if}
				</div>
			</div>

			<div class="flex gap-3 mb-6">
				<a
					href="/events"
					class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
					>View All Events</a
				>
				<a
					href="/events/create"
					class="px-4 py-2 bg-white border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50 transition-colors"
					>+ Create Event</a
				>
				<a
					href="/wallet"
					class="px-4 py-2 bg-white border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50 transition-colors"
					>Wallet</a
				>
				<a
					href="/profile"
					class="px-4 py-2 bg-white border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50 transition-colors"
					>Profile</a
				>
			</div>

			<div class="bg-white rounded-xl border border-gray-200 p-6">
				<h2 class="text-lg font-semibold text-gray-900 mb-4">Open Events</h2>
				{#if events.length === 0}
					<p class="text-center text-gray-500 py-8">No open events yet. Create one!</p>
				{:else}
					<div class="space-y-3">
						{#each events.slice(0, 5) as event}
							<a
								href="/events/{event.id}"
								class="block p-4 border border-gray-200 rounded-lg hover:border-gray-300 hover:shadow-sm transition-all"
							>
								<h3 class="font-medium text-gray-900 text-sm mb-1">{event.title}</h3>
								<p class="text-xs text-gray-500">
									{new Date(event.event_date).toLocaleDateString('en-GB')} · Created by {event.creator_name}
								</p>
								<div class="mt-2 flex gap-4 text-xs">
									<span class="text-green-700 font-medium">YES pool: {event.yes_coins} 🪙</span>
									<span class="text-red-700 font-medium">NO pool: {event.no_coins} 🪙</span>
								</div>
							</a>
						{/each}
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>
