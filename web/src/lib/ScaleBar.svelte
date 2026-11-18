<!--
@component
A bar of segments, each as wide as its share of the planned minutes. A part still to do that
needs Geoff gets a small triangle beneath its segment. A part with no planned minutes draws no
segment, so give every part a planned figure.
-->
<script lang="ts">
	import type { Part } from './sums';

	interface Props {
		parts: Part[];
		/** Read aloud in place of the bar, since the segments carry no text. */
		label: string;
		/** A thinner bar, for the passes under a round. */
		slim?: boolean;
	}

	let { parts, label, slim = false }: Props = $props();
</script>

<div class="scale" class:slim role="img" aria-label={label}>
	{#each parts as part, i (i)}
		<span class="seg {part.state}" class:asks={part.asks} style="flex-grow: {part.minutes}"></span>
	{/each}
</div>
