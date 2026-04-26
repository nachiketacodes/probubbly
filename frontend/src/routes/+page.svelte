<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';

	let mode: 'login' | 'signup' = 'login';
	let loginId = '';
	let pin = '';
	let username = '';
	let error = '';
	let loading = false;

	onMount(() => {
		const token = localStorage.getItem('token');
		const storedUser = localStorage.getItem('user');
		if (token && storedUser) {
			goto('/dashboard');
		}
	});

	async function handleLogin() {
		error = '';
		loading = true;

		try {
			const response = await api.login(loginId, pin);
			const data = await response.json();

			if (response.ok) {
				auth.login(data.token, data.user);
				goto('/dashboard');
			} else {
				error = data.error || 'Invalid Login ID or PIN';
			}
		} catch (err) {
			error = 'Connection failed. Make sure the server is running.';
		} finally {
			loading = false;
		}
	}

	async function handleSignup() {
		error = '';

		// Validate Login ID format
		const loginIdPattern = /^[A-Za-z][0-9]{4}$/;
		if (!loginIdPattern.test(loginId)) {
			error = 'Login ID must be 1 letter followed by 4 digits (e.g. A1234)';
			return;
		}

		// Validate PIN format
		if (!/^[0-9]{4}$/.test(pin)) {
			error = 'PIN must be exactly 4 digits';
			return;
		}

		if (!username.trim()) {
			error = 'Username is required';
			return;
		}

		loading = true;

		try {
			const response = await api.signup(loginId, pin, username.trim());
			const data = await response.json();

			if (response.ok) {
				auth.login(data.token, data.user);
				goto('/dashboard');
			} else {
				error = data.error || 'Signup failed';
			}
		} catch (err) {
			error = 'Connection failed. Make sure the server is running.';
		} finally {
			loading = false;
		}
	}

	function handleSubmit() {
		if (mode === 'login') {
			handleLogin();
		} else {
			handleSignup();
		}
	}
</script>

<div class="min-h-screen bg-gray-50 flex items-center justify-center p-6">
	<div class="w-full max-w-md">
		<div class="bg-white rounded-2xl border border-gray-200 p-8">
			<!-- Logo -->
			<div class="text-center mb-8">
				<h1 class="text-3xl font-bold text-gray-900 tracking-tight">Probubbly</h1>
				<p class="text-sm text-gray-500 mt-1">Educational Probability Platform</p>
			</div>

			<!-- Tabs -->
			<div class="flex bg-gray-100 rounded-lg p-1 mb-6">
				<button
					class="flex-1 py-2 px-4 rounded-md text-sm font-medium transition-all {mode === 'login'
						? 'bg-white text-gray-900 shadow-sm'
						: 'text-gray-600 hover:text-gray-900'}"
					on:click={() => {
						mode = 'login';
						error = '';
					}}
				>
					Sign in
				</button>
				<button
					class="flex-1 py-2 px-4 rounded-md text-sm font-medium transition-all {mode === 'signup'
						? 'bg-white text-gray-900 shadow-sm'
						: 'text-gray-600 hover:text-gray-900'}"
					on:click={() => {
						mode = 'signup';
						error = '';
					}}
				>
					Create account
				</button>
			</div>

			<!-- Error Message -->
			{#if error}
				<div class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700">
					{error}
				</div>
			{/if}

			<!-- Form -->
			<form on:submit|preventDefault={handleSubmit} class="space-y-4">
				<div>
					<label class="block text-xs font-medium text-gray-700 uppercase tracking-wide mb-1.5">
						Login ID
					</label>
					<input
						type="text"
						bind:value={loginId}
						placeholder="e.g. A1234"
						maxlength="5"
						class="w-full px-3 py-2.5 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						required
					/>
					{#if mode === 'signup'}
						<p class="text-xs text-gray-500 mt-1">
							1 letter + 4 digits. Private — not shown to others.
						</p>
					{/if}
				</div>

				<div>
					<label class="block text-xs font-medium text-gray-700 uppercase tracking-wide mb-1.5">
						PIN
					</label>
					<input
						type="password"
						bind:value={pin}
						placeholder="4-digit PIN"
						maxlength="4"
						class="w-full px-3 py-2.5 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						required
					/>
				</div>

				{#if mode === 'signup'}
					<div>
						<label class="block text-xs font-medium text-gray-700 uppercase tracking-wide mb-1.5">
							Username (public)
						</label>
						<input
							type="text"
							bind:value={username}
							placeholder="Display name"
							class="w-full px-3 py-2.5 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
							required
						/>
					</div>
				{/if}

				<button
					type="submit"
					disabled={loading}
					class="w-full bg-blue-600 text-white py-2.5 rounded-lg font-medium text-sm hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed mt-6"
				>
					{#if loading}
						Processing...
					{:else if mode === 'login'}
						Sign in
					{:else}
						Create account — get 500 coins
					{/if}
				</button>
			</form>

			{#if mode === 'login'}
				<div class="mt-4 p-3 bg-gray-50 rounded-lg text-xs text-gray-600">
					<strong class="text-gray-700">Demo:</strong> Admin login — ID: A0000, PIN: 0000
				</div>
			{/if}
		</div>
	</div>
</div>
