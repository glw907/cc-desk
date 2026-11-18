<!--
@component
The frame of every page: the sidebar beside the page, and the browser tab's title. It starts the
poll that keeps the state fresh.
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import '../app.css';
	import { poll } from '#lib/poll.svelte.ts';
	import { tabTitle } from '#lib/session.ts';
	import Banners from '#lib/Banners.svelte';
	import Sidebar from '#lib/Sidebar.svelte';

	interface Props {
		/** The page for the current route. */
		children: Snippet;
	}

	let { children }: Props = $props();

	onMount(() => poll.start());
</script>

<svelte:head>
	<title>{tabTitle(poll.answer, poll.now)}</title>
</svelte:head>

<div class="frame">
	<nav class="rail" aria-label="Status">
		<a href="/" class="wordmark block px-3 pb-6 pt-2 text-xl">cc-desk</a>
		<a href="/rulings" class="rail-row">Rulings</a>
		{#if poll.answer?.snapshot}
			<Sidebar snapshot={poll.answer.snapshot} now={poll.now} />
		{/if}
	</nav>
	<div class="frame-content">
		<Banners
			answer={poll.answer}
			connected={poll.connected}
			answeredAt={poll.answeredAt}
			now={poll.now}
		/>
		{@render children()}
	</div>
</div>
