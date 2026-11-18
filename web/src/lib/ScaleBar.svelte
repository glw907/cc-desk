<!--
@component
A bar of segments, each as wide as its share of the planned minutes. A part still to do that
needs Geoff gets a small triangle beneath its segment. A part with no planned minutes draws no
segment, so give every part a planned figure. Each segment holds its state's words as hidden
text, and "Needs you" where it has a triangle, so no state reaches a reader by color alone.
-->
<script lang="ts">
	import type { Part } from './sums';
	import { word, type Field } from './words';

	interface Props {
		/** The segments in order, each sized by its share of the planned minutes. */
		parts: Part[];
		/** Names the bar's list, ahead of the segments' own words. */
		label: string;
		/** The table that words the parts' state codes: `passState` for passes, `roundState` for rounds. */
		words: Extract<Field, 'passState' | 'roundState'>;
		/** A thinner bar, for the passes under a round. */
		slim?: boolean;
	}

	let { parts, label, words, slim = false }: Props = $props();
</script>

<div class="scale" class:slim role="list" aria-label={label}>
	{#each parts as part, i (i)}
		<span class="seg {part.state}" class:asks={part.asks} role="listitem" style="flex-grow: {part.minutes}"
			><span class="sr-only">{word(words, part.state)}{part.asks ? ', Needs you' : ''}</span></span
		>
	{/each}
</div>
