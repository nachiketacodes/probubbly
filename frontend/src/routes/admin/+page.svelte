<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let stats: any = null;
	let users: any[] = [];
	let houseLedger: any[] = [];
	let loading = true;
	let activeTab = 'stats';
	let error = '';

	onMount(async () => {
		const token = localStorage.getItem('token');
		const storedUser = localStorage.getItem('user');
		if (!token) {
			goto('/');
			return;
		}
		if (storedUser) {
			const user = JSON.parse(storedUser);
			if (!user.is_admin) {
				goto('/dashboard');
				return;
			}
		}
		await loadAll();
	});

	async function loadAll() {
		loading = true;
		try {
			const [statsRes, usersRes, ledgerRes] = await Promise.all([
				api.getAdminStats(),
				api.getAllUsers(),
				api.getHouseLedger()
			]);
			if (statsRes.ok) stats = await statsRes.json();
			if (usersRes.ok) users = await usersRes.json();
			if (ledgerRes.ok) houseLedger = await ledgerRes.json();
		} catch (err) {
			error = 'Failed to load admin data';
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
</script>

<div class="min-h-screen bg-gray-50">
	<header class="bg-white border-b border-gray-200">
		<div class="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
			<div class="flex items-center gap-4">
				<a href="/dashboard" class="text-gray-500 hover:text-gray-900 text-sm">← Dashboard</a>
				<h1 class="text-lg font-bold text-gray-900">Admin Panel</h1>
				<span class="px-2 py-0.5 bg-purple-100 text-purple-700 text-xs font-semibold rounded-full"
					>Admin</span
				>
			</div>
			<button
				on:click={loadAll}
				class="px-3 py-1.5 border border-gray-300 text-gray-600 text-sm rounded-lg hover:bg-gray-50 transition-colors"
			>
				Refresh
			</button>
		</div>
	</header>

	<div class="max-w-6xl mx-auto px-6 py-8">
		{#if loading}
			<div class="text-center py-12"><p class="text-gray-500">Loading admin data...</p></div>
		{:else if error}
			<div class="p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">{error}</div>
		{:else}
			<div class="grid grid-cols-4 gap-4 mb-8">
				<div class="bg-gray-100 rounded-lg p-4">
					<p class="text-xs text-gray-500 uppercase tracking-wide font-medium">Total Users</p>
					<p class="text-2xl font-semibold text-gray-900 mt-1">{stats?.total_users}</p>
				</div>
				<div class="bg-gray-100 rounded-lg p-4">
					<p class="text-xs text-gray-500 uppercase tracking-wide font-medium">Open Events</p>
					<p class="text-2xl font-semibold text-gray-900 mt-1">{stats?.open_events}</p>
				</div>
				<div class="bg-gray-100 rounded-lg p-4">
					<p class="text-xs text-gray-500 uppercase tracking-wide font-medium">Total Forecasts</p>
					<p class="text-2xl font-semibold text-gray-900 mt-1">{stats?.total_predictions}</p>
				</div>
				<div class="bg-gray-100 rounded-lg p-4">
					<p class="text-xs text-gray-500 uppercase tracking-wide font-medium">House Earnings</p>
					<p class="text-2xl font-semibold text-gray-900 mt-1 font-mono">
						🪙 {stats?.total_house_earnings}
					</p>
				</div>
			</div>

			<div class="flex border-b border-gray-200 mb-6">
				{#each [['stats', 'Platform'], ['users', 'Users'], ['ledger', 'House Ledger']] as [tab, label]}
					<button
						on:click={() => (activeTab = tab)}
						class="px-4 py-2 text-sm font-medium border-b-2 transition-colors {activeTab === tab
							? 'border-blue-600 text-blue-600'
							: 'border-transparent text-gray-500 hover:text-gray-900'}">{label}</button
					>
				{/each}
			</div>

			{#if activeTab === 'stats'}
				<div class="grid grid-cols-2 gap-6">
					<div class="bg-white border border-gray-200 rounded-xl p-5">
						<h3 class="font-semibold text-gray-900 mb-4">Event Overview</h3>
						<div class="space-y-3">
							<div class="flex justify-between">
								<span class="text-sm text-gray-600">Total Events</span>
								<span class="text-sm font-semibold">{stats?.total_events}</span>
							</div>
							<div class="flex justify-between">
								<span class="text-sm text-gray-600">Open Events</span>
								<span class="text-sm font-semibold text-blue-600">{stats?.open_events}</span>
							</div>
							<div class="flex justify-between">
								<span class="text-sm text-gray-600">Resolved Events</span>
								<span class="text-sm font-semibold text-gray-600">{stats?.resolved_events}</span>
							</div>
							<div class="flex justify-between">
								<span class="text-sm text-gray-600">Total Predictions</span>
								<span class="text-sm font-semibold">{stats?.total_predictions}</span>
							</div>
						</div>
					</div>
					<div class="bg-white border border-gray-200 rounded-xl p-5">
						<h3 class="font-semibold text-gray-900 mb-4">Coin Economy</h3>
						<div class="space-y-3">
							<div class="flex justify-between">
								<span class="text-sm text-gray-600">Coins in Circulation</span>
								<span class="text-sm font-semibold font-mono">🪙 {stats?.total_coins_in_play}</span>
							</div>
							<div class="flex justify-between">
								<span class="text-sm text-gray-600">House Earnings</span>
								<span class="text-sm font-semibold text-green-700 font-mono"
									>🪙 {stats?.total_house_earnings}</span
								>
							</div>
							<div class="flex justify-between">
								<span class="text-sm text-gray-600">House Cut Rate</span>
								<span class="text-sm font-semibold">3%</span>
							</div>
							<div class="flex justify-between">
								<span class="text-sm text-gray-600">Daily Loan Amount</span>
								<span class="text-sm font-semibold font-mono">🪙 400</span>
							</div>
						</div>
					</div>
				</div>
			{:else if activeTab === 'users'}
				<div class="bg-white border border-gray-200 rounded-xl overflow-hidden">
					<div class="px-5 py-4 border-b border-gray-200">
						<h3 class="font-semibold text-gray-900">All Users ({users.length})</h3>
					</div>
					<div class="divide-y divide-gray-100">
						{#each users as user}
							<div class="px-5 py-3 flex items-center justify-between">
								<div class="flex items-center gap-3">
									<div
										class="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center text-sm font-bold text-blue-700"
									>
										{user.username?.charAt(0).toUpperCase()}
									</div>
									<div>
										<div class="flex items-center gap-2">
											<p class="text-sm font-medium text-gray-900">{user.username}</p>
											{#if user.is_admin}
												<span
													class="px-1.5 py-0.5 bg-purple-100 text-purple-700 text-xs font-semibold rounded"
													>Admin</span
												>
											{/if}
										</div>
										<p class="text-xs text-gray-400">Joined {formatDate(user.joined_at)}</p>
									</div>
								</div>
								<div class="text-right">
									<p class="text-sm font-semibold font-mono">🪙 {user.balance}</p>
									<p class="text-xs text-gray-400">{user.prediction_count} predictions</p>
								</div>
							</div>
						{/each}
					</div>
				</div>
			{:else if activeTab === 'ledger'}
				<div class="bg-white border border-gray-200 rounded-xl overflow-hidden">
					<div class="px-5 py-4 border-b border-gray-200">
						<h3 class="font-semibold text-gray-900">House Ledger</h3>
						<p class="text-xs text-gray-500 mt-0.5">3% cut collected from winning predictions</p>
					</div>
					{#if houseLedger.length === 0}
						<div class="text-center py-12">
							<p class="text-gray-500 text-sm">No house earnings yet</p>
							<p class="text-gray-400 text-xs mt-1">Earnings appear when events are resolved</p>
						</div>
					{:else}
						<div class="divide-y divide-gray-100">
							{#each houseLedger as entry}
								<div class="px-5 py-3 flex items-center justify-between">
									<div>
										<p class="text-sm text-gray-900">{entry.event_title}</p>
										<p class="text-xs text-gray-400 mt-0.5">{formatDate(entry.created_at)}</p>
									</div>
									<p class="text-sm font-semibold text-green-700 font-mono">
										+{entry.cut_amount} 🪙
									</p>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			{/if}
		{/if}
	</div>
</div>
