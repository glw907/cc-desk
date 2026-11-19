<!--
@component
The sidebar: the next-session forecast, what waits for Geoff, the reports to read, and each
release under way. It draws nothing under the forecast until the first good snapshot arrives.
-->
<script lang="ts">
	import NextSession from './NextSession.svelte';
	import ReleaseBlock from './ReleaseBlock.svelte';
	import { terminalLink } from './link';
	import type { Snapshot } from './snapshot';
	import { count, countOrNo } from './sums';
	import { formatWhen } from './time';

	interface Props {
		/** The whole snapshot, so each release's sums come from its own rounds. */
		snapshot: Snapshot;
		/** The browser's clock, passed on to the forecast. */
		now: Date;
	}

	let { snapshot, now }: Props = $props();

	// The snapshot lists reports oldest first; the sidebar shows the newest on top.
	const unread = $derived(snapshot.reports.filter((r) => r.read === null).toReversed());
</script>

<NextSession next={snapshot.next_session} {now} />

<section class="rail-section" aria-labelledby="waiting-heading">
	<h2 class="rail-heading" id="waiting-heading">Waiting for you</h2>
	{#each snapshot.waits_for_geoff as wait (`${wait.project}/${wait.round}`)}
		{@const href = terminalLink(wait.project, wait.round)}
		<div class="rail-row waiting-row">
			<a href="/{wait.project}/{wait.round}">
				{wait.project}/{wait.round}
				<span class="sub">{count(wait.questions, 'question')}</span>
			</a>
			{#if href}
				<a class="term-link" {href}>Open {wait.project}/{wait.round} in a terminal</a>
			{/if}
		</div>
	{:else}
		<p class="rail-row muted">Nothing</p>
	{/each}
</section>

<section class="rail-section" aria-labelledby="reports-heading">
	<h2 class="rail-heading" id="reports-heading">Reports to read</h2>
	{#each unread as report (report.id)}
		<div class="rail-row">
			<a href="/reports/{report.id}">
				{report.project}: {report.title}
				<span class="sub">
					{countOrNo(report.calls.length, 'judgment call')}, {countOrNo(
						report.misses.length,
						'planning miss'
					)}. {formatWhen(report.sent, now)}
				</span>
			</a>
		</div>
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
