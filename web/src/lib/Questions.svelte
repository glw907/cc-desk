<!--
@component
The written questions of a round or a pass. Each shows its text, the background a reader needs,
its options as cards, and the reason for the recommendation. A card is plain text: the page
holds no control for answering, which happens at the desk. An answered question shows how it was
answered.
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
			{#if !q.answered}
				{#each q.background as paragraph, i (i)}
					<p class="muted">{paragraph}</p>
				{/each}
				{#if q.options.length > 0}
					<div class="option-cards">
						{#each q.options as option (option)}
							<div class="option-card">
								<p class="option-name">
									{option}
									{#if option === q.recommended}<span class="recommended">Recommended</span>{/if}
								</p>
								{#if q.causes[option]}<p class="muted">{q.causes[option]}</p>{/if}
							</div>
						{/each}
					</div>
				{/if}
				{#if q.reason}<p class="muted">{q.reason}</p>{/if}
			{/if}
		</li>
	{/each}
</ul>
