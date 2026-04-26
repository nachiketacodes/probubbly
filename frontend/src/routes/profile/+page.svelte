<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let user: any = null;
	let transactions: any[] = [];
	let loading = true;
	let error = '';

	let totalWins = 0;
	let totalLosses = 0;
	let totalEarned = 0;
	let totalSpent = 0;
	let winRate = 0;

	onMount(async () => {
		const token = localStorage.getItem('token');
		const storedUser = localStorage.getItem('user');
		if (!token) {
			goto('/');
			return;
		}
		if (storedUser) user = JSON.parse(storedUser);
		await loadProfile();
	});

	async function loadProfile() {
		loading = true;
		try {
			const res = await api.getWallet();
			if (res.ok) {
				const data = await res.json();
				user = data.user;
				transactions = data.transactions;

				totalWins = transactions.filter((t) => t.type === 'payout').length;
				totalLosses = transactions.filter((t) => t.type === 'loss').length;
				totalEarned = transactions
					.filter((t) => t.amount > 0 && t.type === 'payout')
					.reduce((s, t) => s + t.amount, 0);
				totalSpent = transactions
					.filter((t) => t.type === 'predict')
					.reduce((s, t) => s + Math.abs(t.amount), 0);

				const resolved = totalWins + totalLosses;
				winRate = resolved > 0 ? Math.round((totalWins / resolved) * 100) : 0;
			} else {
				error = 'Failed to load profile';
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
			month: 'long',
			year: 'numeric'
		});
	}

	function initials(name: string) {
		return name?.charAt(0).toUpperCase() || '?';
	}
</script>

<div class="min-h-screen bg-gray-50">
	<header class="bg-white border-b border-gray-200">
		<div class="max-w-7xl mx-auto px-6 py-4 flex items-center gap-4">
			<a href="/dashboard" class="text-gray-500 hover:text-gray-900 text-sm">← Dashboard</a>
			<h1 class="text-lg font-bold text-gray-900">Profile</h1>
		</div>
	</header>

	<div class="max-w-3xl mx-auto px-6 py-8">
		{#if loading}
			<div class="text-center py-12"><p class="text-gray-500">Loading profile...</p></div>
		{:else if error}
			<div class="p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">{error}</div>
		{:else}
			<div class="bg-white border border-gray-200 rounded-xl p-6 mb-6">
				<div class="flex items-center gap-4 mb-6">
					<div
						class="w-14 h-14 rounded-full bg-blue-100 flex items-center justify-center text-2xl font-bold text-blue-700"
					>
						{initials(user?.username)}
					</div>
					<div>
						<h2 class="text-xl font-bold text-gray-900">{user?.username}</h2>
						<p class="text-sm text-gray-500">Member since {formatDate(user?.joined_at)}</p>
						{#if user?.is_admin}
							<span
								class="inline-block mt-1 px-2 py-0.5 bg-purple-100 text-purple-700 text-xs font-semibold rounded-full"
								>Admin</span
							>
						{/if}
					</div>
				</div>

				<div class="grid grid-cols-4 gap-3">
					<div class="bg-gray-50 rounded-lg p-3 text-center">
						<p class="text-xs text-gray-500 uppercase tracking-wide font-medium">Balance</p>
						<p class="text-xl font-bold text-gray-900 mt-1 font-mono">🪙 {user?.balance}</p>
					</div>
					<div class="bg-gray-50 rounded-lg p-3 text-center">
						<p class="text-xs text-gray-500 uppercase tracking-wide font-medium">Win Rate</p>
						<p class="text-xl font-bold text-gray-900 mt-1">{winRate}%</p>
					</div>
					<div class="bg-gray-50 rounded-lg p-3 text-center">
						<p class="text-xs text-gray-500 uppercase tracking-wide font-medium">Wins</p>
						<p class="text-xl font-bold text-green-700 mt-1">{totalWins}</p>
					</div>
					<div class="bg-gray-50 rounded-lg p-3 text-center">
						<p class="text-xs text-gray-500 uppercase tracking-wide font-medium">Losses</p>
						<p class="text-xl font-bold text-red-700 mt-1">{totalLosses}</p>
					</div>
				</div>
			</div>

			<div class="grid grid-cols-2 gap-4 mb-6">
				<div class="bg-white border border-gray-200 rounded-xl p-5">
					<p class="text-xs text-gray-500 uppercase tracking-wide font-medium mb-1">Total Earned</p>
					<p class="text-2xl font-bold text-green-700 font-mono">+{totalEarned} 🪙</p>
					<p class="text-xs text-gray-400 mt-1">from winning predictions</p>
				</div>
				<div class="bg-white border border-gray-200 rounded-xl p-5">
					<p class="text-xs text-gray-500 uppercase tracking-wide font-medium mb-1">Total Spent</p>
					<p class="text-2xl font-bold text-red-700 font-mono">{totalSpent} 🪙</p>
					<p class="text-xs text-gray-400 mt-1">on predictions placed</p>
				</div>
			</div>

			<div class="bg-white border border-gray-200 rounded-xl overflow-hidden">
				<div class="px-5 py-4 border-b border-gray-200">
					<h3 class="font-semibold text-gray-900">Recent Activity</h3>
				</div>
				{#if transactions.length === 0}
					<div class="text-center py-12">
						<p class="text-gray-500 text-sm">No activity yet</p>
						<a href="/events" class="inline-block mt-3 text-blue-600 text-sm hover:underline"
							>Browse events to start predicting</a
						>
					</div>
				{:else}
					<div class="divide-y divide-gray-100">
						{#each transactions.slice(0, 10) as tx}
							<div class="px-5 py-3 flex items-center justify-between">
								<div>
									<p class="text-sm text-gray-900">{tx.description}</p>
									<p class="text-xs text-gray-400 mt-0.5 capitalize">{tx.type}</p>
								</div>
								<p
									class="text-sm font-semibold font-mono {tx.amount > 0
										? 'text-green-700'
										: tx.amount < 0
											? 'text-red-700'
											: 'text-gray-500'}"
								>
									{tx.amount > 0 ? '+' : ''}{tx.amount} 🪙
								</p>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>
