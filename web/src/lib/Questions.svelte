<!--
@component
The written questions of a round or a pass, each with how it was answered when it has been.
-->
<script lang="ts">
	import type { Question } from './snapshot';
	import { word } from './words.ts';

	interface Props {
		/** The questions to list, in the snapshot's order. */
		questions: Question[];
	}

	let { questions }: Props = $props();
</script>

<ul class="calls">
	{#each questions as q (q.id)}
		<li>
			{q.text}
			{#if q.answered && q.answer_kind}<span class="muted">{word('answerKind', q.answer_kind)}</span>{/if}
		</li>
	{/each}
</ul>
