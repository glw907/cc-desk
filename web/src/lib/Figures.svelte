<!--
@component
The figures of one item as a definition list. Each value carries `aria-labelledby` for its own
term, so a reader or a test finds a figure by its label. A null value shows the en dash with
"None" for a screen reader.
-->
<script lang="ts">
	import { DASH, NONE } from './time.ts';

	interface Figure {
		label: string;
		/** The text to show, already formatted. Null when the snapshot holds no value. */
		value: string | null;
	}

	interface Props {
		/** The figures in the order they read. */
		figures: Figure[];
	}

	let { figures }: Props = $props();

	const uid = $props.id();
</script>

<dl class="figure-list">
	{#each figures as figure, i (figure.label)}
		<div>
			<dt id="{uid}-{i}">{figure.label}</dt>
			<dd aria-labelledby="{uid}-{i}">
				{#if figure.value === null}<span aria-hidden="true">{DASH}</span><span class="sr-only"
						>{NONE}</span
					>{:else}{figure.value}{/if}
			</dd>
		</div>
	{/each}
</dl>
