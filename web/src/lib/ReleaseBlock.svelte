<!--
@component
One release in the sidebar: a bar of its rounds with the percent of the work done, then each
building round hung beneath it on a rule, with a bar of that round's passes.
-->
<script lang="ts">
	import ScaleBar from './ScaleBar.svelte';
	import type { Release } from './snapshot';
	import { buildingRounds, doneCount, passParts, percentDone, roundParts } from './sums';
	import { word } from './words';

	interface Props {
		/** The release to draw: its scale bar, its rounds in order, and its "Rough" mark. */
		release: Release;
	}

	let { release }: Props = $props();

	const roundsDone = $derived(
		`${doneCount(release.rounds)} of ${release.rounds.length} rounds done`
	);
</script>

<article class="release" data-release="{release.project}/{release.name}">
	<p class="release-head">
		<span><strong>{release.project}</strong> to {release.name}</span>
		{#if release.rough}<span class="rough">Rough</span>{/if}
	</p>
	<ScaleBar parts={roundParts(release.rounds)} label={roundsDone} />
	<p class="figures">
		<span>{roundsDone}</span>
		<span>{percentDone(release.rounds)}% of the work</span>
	</p>

	<ul class="branch">
		{#each buildingRounds(release) as round (round.name)}
			{@const passesDone = `${doneCount(round.passes)} of ${round.passes.length} passes done`}
			<li data-round="{release.project}/{round.name}">
				<p class="branch-head">
					<a class="branch-title" href="/{release.project}/{round.name}">{round.name}</a>
				</p>
				<ScaleBar parts={passParts(round.passes)} label={passesDone} slim />
				<p class="figures">
					<span>{passesDone}</span>
					<span class="state {round.state}">{word('roundState', round.state)}</span>
				</p>
			</li>
		{/each}
	</ul>
</article>
