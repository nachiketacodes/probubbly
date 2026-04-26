<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	let user: any = null;
	let transactions: any[] = [];
	let loading = true;
	let borrowing = false;
	let canBorrow = false;
	let error = '';

	onMount(async () => {
		const token = localStorage.getItem('token');
		if (!token) {
			goto('/');
			return;
		}
		await loadWallet();
	});

	async function loadWallet() {
		loading = true;
		try {
			const res = await api.getWallet();
			if (res.ok) {
				const data = await res.json();
				user = data.user;
				transactions = data.transactions;
				const today = new Date().toISOString().split('T')[0];
				const lastBorrowDate = user.last_borrow ? user.last_borrow.split('T')[0] : null;
				canBorrow = !lastBorrowDate || lastBorrowDate !== today;
			} else {
				error = 'Failed to load wallet';
			}
		} catch (err) {
			error = 'Connection failed';
		} finally {
			loading = false;
		}
	}

	async function handleBorrow() {
		borrowing = true;
		try {
			const res = await api.borrowCoins();
			if (res.ok) {
				await loadWallet();
			} else {
				const text = await res.text();
				alert(text || 'Borrow failed');
			}
		} catch (err) {
			alert('Connection failed');
		} finally {
			borrowing = false;
		}
	}

	function formatDate(dateStr: string) {
		return new Date(dateStr).toLocaleDateString('en-GB', {
			day: 'numeric',
			month: 'short',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function formatCoins(amount: number): string {
		return Number(amount).toFixed(4);
	}

	function txColour(type: string): string {
		if (type === 'payout' || type === 'signup' || type === 'borrow') return 'text-green-700';
		if (type === 'predict' || type === 'loss') return 'text-red-700';
		return 'text-gray-700';
	}

	function txPrefix(amount: number): string {
		if (amount > 0) return '+';
		return '';
	}
</script>

<div class="min-h-screen bg-gray-50">
	<header class="bg-white border-b border-gray-200">
		<div class="max-w-7xl mx-auto px-6 py-4 flex items-center gap-4">
			<a href="/dashboard" class="text-gray-500 hover:text-gray-900 text-sm">← Dashboard</a>
			<h1 class="text-lg font-bold text-gray-900">Wallet</h1>
		</div>
	</header>

	<div class="max-w-3xl mx-auto px-6 py-8">
		{#if loading}
			<div class="text-center py-12"><p class="text-gray-500">Loading wallet...</p></div>
		{:else if error}
			<div class="p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">{error}</div>
		{:else}
			<div class="grid grid-cols-3 gap-4 mb-6">
				<div class="bg-gray-100 rounded-lg p-4">
					<p class="text-xs text-gray-500 uppercase tracking-wide font-medium">Balance</p>
					<p class="text-2xl font-semibold text-gray-900 mt-1 font-mono">
						🪙 {formatCoins(user?.balance)}
					</p>
				</div>
				<div class="bg-gray-100 rounded-lg p-4">
					<p class="text-xs text-gray-500 uppercase tracking-wide font-medium">Total Borrowed</p>
					<p class="text-2xl font-semibold text-gray-900 mt-1 font-mono">
						{formatCoins(user?.borrowed)}
					</p>
				</div>
				<div class="bg-gray-100 rounded-lg p-4">
					<p class="text-xs text-gray-500 uppercase tracking-wide font-medium">Transactions</p>
					<p class="text-2xl font-semibold text-gray-900 mt-1">{transactions.length}</p>
				</div>
			</div>

			<div class="bg-white border border-gray-200 rounded-xl p-5 mb-6">
				<div class="flex items-center justify-between">
					<div>
						<p class="font-semibold text-gray-900 text-sm">Daily House Loan</p>
						<p class="text-xs text-gray-500 mt-0.5">
							400 coins · once per day · resets at midnight UTC
						</p>
					</div>
					{#if canBorrow}
						<button
							on:click={handleBorrow}
							disabled={borrowing}
							class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
						>
							{borrowing ? 'Borrowing...' : 'Borrow 400 coins'}
						</button>
					{:else}
						<span class="px-3 py-1.5 bg-gray-100 text-gray-500 text-sm rounded-lg">Used today</span>
					{/if}
				</div>
			</div>

			<div class="bg-white border border-gray-200 rounded-xl overflow-hidden">
				<div class="px-5 py-4 border-b border-gray-200">
					<h2 class="font-semibold text-gray-900">Transaction History</h2>
				</div>
				{#if transactions.length === 0}
					<div class="text-center py-12">
						<p class="text-gray-500 text-sm">No transactions yet</p>
					</div>
				{:else}
					<div class="divide-y divide-gray-100">
						{#each transactions as tx}
							<div class="px-5 py-3 flex items-center justify-between">
								<div>
									<p class="text-sm text-gray-900">{tx.description}</p>
									<p class="text-xs text-gray-400 mt-0.5">{formatDate(tx.created_at)}</p>
								</div>
								<div class="text-right">
									<p class="text-sm font-semibold font-mono {txColour(tx.type)}">
										{txPrefix(tx.amount)}{formatCoins(tx.amount)} 🪙
									</p>
									<p class="text-xs text-gray-400 capitalize">{tx.type}</p>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>
