<!--
@component
The forecast of the next decision session, the one boxed thing in the sidebar: its state in
words, the likely range, and what it will ask. A missing range shows a dash that a screen
reader reads as "None".
-->
<script lang="ts">
	import { sessionWords } from './session';
	import type { NextSession } from './snapshot';
	import { count, list } from './sums';
	import { DASH, NONE, formatRange } from './time';

	interface Props {
		/** The snapshot's forecast. A `not_designed` entry names a round with no plan yet. */
		next: NextSession;
		/** The browser's clock, which "Today" and "Tomorrow" are measured from. */
		now: Date;
	}

	let { next, now }: Props = $props();

	const words = $derived(sessionWords(next, now));
</script>

<section class="session" aria-label="Next decision session">
	<span class="session-label">Next decision session</span>
	{#if words !== null}
		<span class="session-time">
			<span data-testid="session-state">{words}</span>
			{#if next.rough}<span class="rough">Rough</span>{/if}
		</span>
	{/if}
	{#if next.state !== 'nothing_waiting'}
		<span class="session-sub" data-testid="session-range">
			{#if next.range}{formatRange(next.range.low, next.range.high)}{:else}<span aria-hidden="true"
					>{DASH}</span
				><span class="sr-only">{NONE}</span>{/if}
		</span>
	{/if}
	{#if next.question_count !== null && next.question_count > 0}
		<span class="session-sub">
			{count(next.question_count, 'question')} from {count(next.rounds.length, 'round')}, in {list(
				next.projects
			)}
		</span>
	{/if}
	{#if next.not_designed.length > 0}
		<span class="session-sub">Not designed yet: {list(next.not_designed)}</span>
	{/if}
	{#if next.state === 'needs_you_now' && next.reasons.length > 0}
		<ul class="session-sub reasons">
			{#each next.reasons as reason (reason)}<li>{reason}</li>{/each}
		</ul>
	{/if}
</section>
