<!--
@component
A map scale bar that draws progress: one segment per part, as wide as its share of the work.
A finished segment is inked, one in flight takes its state's color, and the rest stay paper.
-->
<script lang="ts">
	import type { State } from './mock.ts';
	let { parts, label, slim = false }: { parts: { min: number; state: State; asks?: unknown }[]; label: string; slim?: boolean } = $props();
</script>

<div class="scale" class:slim role="img" aria-label={label}>
	{#each parts as p, i (i)}
		<span class="seg {p.state}" class:asks={p.asks && p.state === 'todo'} style="flex-grow: {p.min}"></span>
	{/each}
</div>
