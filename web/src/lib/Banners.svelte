<!--
@component
The notices above the page: the state is stale, there is none yet, or the server has stopped
answering. Each keeps the last good page below it. Shows nothing while the state is current.
-->
<script lang="ts">
	import type { StateAnswer } from './snapshot';
	import { formatWhen } from './time';

	interface Props {
		/** The last answer the server gave, or null before the first one. */
		answer: StateAnswer | null;
		/** False when the last request got no answer. */
		connected: boolean;
		/** When the last good answer arrived, by the browser's clock. */
		answeredAt: Date | null;
		/** The browser's clock, for "Today". */
		now: Date;
	}

	let { answer, connected, answeredAt, now }: Props = $props();
</script>

{#if answer && !connected}
	<div class="banner" role="status" data-banner="not-connected">
		<strong>Not connected</strong>
		<span>
			The page can't reach cc-desk.
			{#if answeredAt}Last answer: {formatWhen(answeredAt.toISOString(), now)}.{/if}
		</span>
	</div>
{/if}

{#if answer?.noState}
	<div class="banner" role="status" data-banner="no-state">
		<strong>No state yet</strong>
		{#if answer.error}<span>{answer.error}</span>{/if}
	</div>
{:else if answer?.error}
	<div class="banner" role="status" data-banner="stale">
		<strong>Stale</strong>
		<span>
			{answer.error}
			{#if answer.lastGood}Last good snapshot: {formatWhen(answer.lastGood, now)}.{/if}
		</span>
	</div>
{/if}
