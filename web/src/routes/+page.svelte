<!--
@component
The front page: what the next decision session covers, any open guesses, and a board for each
release. It draws nothing until the first good snapshot arrives.
-->
<script lang="ts">
	import { poll } from '#lib/poll.svelte.ts';
	import ReleaseBoard from '#lib/ReleaseBoard.svelte';
	import { sessionWords } from '#lib/session.ts';
	import { count, findRound, list } from '#lib/sums.ts';
	import { DASH, NONE, formatRange, formatWhen } from '#lib/time.ts';

	const snapshot = $derived(poll.answer?.snapshot ?? null);
</script>

<main class="desk">
	{#if snapshot}
		{@const next = snapshot.next_session}
		<section class="next grid gap-3" aria-labelledby="next-heading">
			<h2 id="next-heading">
				Your next decision session: {sessionWords(next, poll.now) ?? 'none'}
			</h2>
			{#if next.state !== 'nothing_waiting'}
				<p class="muted" data-testid="front-range">
					Likely range:
					{#if next.range}{formatRange(next.range.low, next.range.high)}{:else}<span
							aria-hidden="true">{DASH}</span
						><span class="sr-only">{NONE}</span>{/if}.
				</p>
			{/if}
			{#if next.question_count !== null && next.question_count > 0}
				<p class="muted">
					{count(next.question_count, 'question')} from {count(next.rounds.length, 'round')}, in
					{list(next.projects)}.
				</p>
			{/if}
			<ul class="ahead">
				{#each next.rounds as name (name)}
					{@const round = findRound(snapshot, name)}
					{#if round}
						<li>
							<a href="/{name}"><strong>{name}</strong>, {round.summary}</a>
							{#each round.needs.questions.filter((q) => !q.answered) as q (q.id)}
								<span class="muted">{q.text}</span>
							{/each}
							{#if round.needed_time}
								<span class="muted">Needed by {formatWhen(round.needed_time, poll.now)}.</span>
							{/if}
						</li>
					{/if}
				{/each}
			</ul>
		</section>

		{#if snapshot.open_guesses.length > 0}
			<section class="next grid gap-3" aria-labelledby="guess-heading">
				<h2 id="guess-heading">Guesses waiting on an answer</h2>
				<ul class="ahead">
					{#each snapshot.open_guesses as guess (`${guess.project}/${guess.pass}/${guess.question}`)}
						<li data-guess="{guess.project}/{guess.pass}">
							<span><strong>{guess.project}/{guess.pass}</strong>, {guess.minutes} minutes</span>
						</li>
					{/each}
				</ul>
				<p class="muted">
					{snapshot.guess_minutes_total} of {snapshot.guess_minutes_limit} minutes
				</p>
			</section>
		{/if}

		<section class="grid gap-3" aria-labelledby="boards-heading">
			<h2 id="boards-heading">{count(snapshot.releases.length, 'release')} under way</h2>
			<div class="boards">
				{#each snapshot.releases as release (`${release.project}/${release.name}`)}
					<ReleaseBoard {release} {snapshot} />
				{/each}
			</div>
		</section>
	{/if}
</main>
