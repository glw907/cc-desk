<!--
@component
A pass's page at `/<project>/<round>/<pass>`: its name, summary, state, figures, what it needs,
and the command of its round. A stopped pass shows its reason and no time, because the snapshot
holds no stop time. The pass's own needs are its questions: the snapshot gives a pass no needs
kind or reason, so a pass with none borrows its round's reason when the round needs nothing.
-->
<script lang="ts">
	import { page } from '$app/state';
	import Command from '#lib/Command.svelte';
	import Figures from '#lib/Figures.svelte';
	import NotFound from '#lib/NotFound.svelte';
	import Questions from '#lib/Questions.svelte';
	import { poll } from '#lib/poll.svelte.ts';
	import { findPassAt } from '#lib/sums.ts';
	import { formatMinutes, formatWhen } from '#lib/time.ts';
	import { word } from '#lib/words.ts';

	const snapshot = $derived(poll.answer?.snapshot ?? null);
	const found = $derived(
		snapshot
			? findPassAt(
					snapshot,
					page.params.project ?? '',
					page.params.round ?? '',
					page.params.pass ?? ''
				)
			: undefined
	);
	const minutes = (n: number | null) => (n === null ? null : formatMinutes(n));
</script>

<svelte:head>
	{#if found}<title>{found.pass.name} · {found.release.project} · cc-desk</title>{/if}
</svelte:head>

{#if snapshot && !found}
	<NotFound address={page.url.pathname} />
{:else if found}
	{@const { release, round, pass } = found}
	<main class="page">
		<div class="page-main">
			<p class="doc-meta muted">
				<a href="/{release.project}/{round.name}">{release.project}/{round.name}</a>
			</p>
			<h1 class="doc-title">{pass.name}</h1>

			<div class="prose-doc mt-6">
				<p>{pass.summary}</p>
				<p class="state-line">
					<span class="state {pass.state}">{word('passState', pass.state)}</span>
					{#if pass.state === 'stopped' && pass.stop_reason}
						<span class="why">{word('stopReason', pass.stop_reason)}</span>
					{/if}
				</p>

				<h2>Figures</h2>
				<Figures
					figures={[
						{ label: 'Planned', value: minutes(pass.planned_minutes) },
						{ label: 'Actual', value: minutes(pass.actual_minutes) },
						{ label: 'Work before it', value: minutes(pass.minutes_before) },
						{
							label: 'Needed by',
							value: pass.needed_time === null ? null : formatWhen(pass.needed_time, poll.now)
						}
					]}
				/>

				<h2>What it needs from you</h2>
				{#if pass.questions.length > 0}
					<Questions questions={pass.questions} />
				{:else}
					<p>
						Nothing. {round.needs.kind === 'nothing' && round.needs.reason
							? round.needs.reason
							: 'No reason recorded'}
					</p>
				{/if}

				<Command project={release.project} round={round.name} />
			</div>
		</div>
	</main>
{/if}
