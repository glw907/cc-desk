<!--
@component
The sidebar: the next-session forecast, what waits for Geoff, and each release under way. It
draws nothing under the forecast until the first good snapshot arrives.
-->
<script lang="ts">
	import NextSession from './NextSession.svelte';
	import ReleaseBlock from './ReleaseBlock.svelte';
	import type { Snapshot } from './snapshot';
	import { count } from './sums';

	interface Props {
		/** The whole snapshot, so each release's sums come from its own rounds. */
		snapshot: Snapshot;
		/** The browser's clock, passed on to the forecast. */
		now: Date;
	}

	let { snapshot, now }: Props = $props();
</script>

<NextSession next={snapshot.next_session} {now} />

<section class="rail-section" aria-labelledby="waiting-heading">
	<h2 class="rail-heading" id="waiting-heading">Waiting for you</h2>
	{#each snapshot.waits_for_geoff as wait (`${wait.project}/${wait.round}`)}
		<a class="rail-row" href="/{wait.project}/{wait.round}">
			{wait.project}/{wait.round}
			<span class="sub">{count(wait.questions, 'question')}</span>
		</a>
	{:else}
		<p class="rail-row muted">Nothing</p>
	{/each}
</section>

<section class="rail-section" aria-labelledby="running-heading">
	<h2 class="rail-heading" id="running-heading">Running</h2>
	{#each snapshot.releases as release (`${release.project}/${release.name}`)}
		<ReleaseBlock {release} />
	{/each}
</section>
