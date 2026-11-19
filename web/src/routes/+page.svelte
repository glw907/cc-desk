<!--
@component
The front page: what the next decision session covers, any open guesses, what the session check
did lately (read from the check log, since `check.json` is written only by `cc-desk check`), each
session's state, and a board for each release. It draws nothing until the first good snapshot
arrives.
-->
<script lang="ts">
	import { poll } from '#lib/poll.svelte.ts';
	import ReleaseBoard from '#lib/ReleaseBoard.svelte';
	import { sessionWords } from '#lib/session.ts';
	import { count, findRound, list } from '#lib/sums.ts';
	import type { CheckLogLine, SessionEntry } from '#lib/snapshot.ts';
	import { DASH, NONE, formatRange, formatWhen } from '#lib/time.ts';
	import { word } from '#lib/words.ts';

	const snapshot = $derived(poll.answer?.snapshot ?? null);

	/** One check-log line as a sentence. The outcome says what happened, which may not be what was chosen. */
	function checkSentence(line: CheckLogLine): string {
		const choice = word('checkDecision', line.decision);
		const why = line.reason ? `, because ${word('checkReason', line.reason)}` : '';
		return `${formatWhen(line.time, poll.now)}: on ${line.project}, ${line.record}, the check chose to ${choice}${why}. Result: ${line.outcome}.`;
	}

	/** The reason of the newest restart the check chose for a session's record, in words. */
	function lastRelaunchReason(s: SessionEntry): string | null {
		const record = s.pass ? `pass:${s.pass}` : s.kind;
		const lines = (snapshot?.check_log ?? []).filter(
			(l) => l.project === s.project && l.record === record && l.decision.includes('relaunch') && l.reason
		);
		const reason = lines.at(-1)?.reason;
		return reason ? word('checkReason', reason) : null;
	}
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

		<section class="next grid gap-3" aria-labelledby="check-heading">
			<h2 id="check-heading">Since the last check</h2>
			{#if snapshot.check_log.length === 0}
				<p class="muted">The check has done nothing yet.</p>
			{:else}
				<ul class="ahead">
					{#each snapshot.check_log.slice(-10).toReversed() as line (`${line.time}/${line.record}/${line.decision}`)}
						<li>
							<span>{checkSentence(line)}</span>
						</li>
					{/each}
				</ul>
			{/if}
		</section>

		<section class="next grid gap-3" aria-labelledby="sessions-heading">
			<h2 id="sessions-heading">Sessions</h2>
			{#if snapshot.sessions.length === 0}
				<p class="muted">No session is recorded.</p>
			{:else}
				<ul class="ahead">
					{#each snapshot.sessions as s (`${s.project}/${s.name}`)}
						<li data-session={s.name}>
							{#if s.pass && s.round}
								<a href="/{s.project}/{s.round}/{s.pass}"><strong>{s.name}</strong></a>
							{:else if s.round}
								<a href="/{s.project}/{s.round}"><strong>{s.name}</strong></a>
							{:else}
								<strong>{s.name}</strong>
							{/if}
							<span class="muted">
								{s.last_state ? word('sessionState', s.last_state) : 'No state seen yet'}.
								{#if s.relaunches > 0}
									Restarted {s.relaunches === 1 ? 'once' : `${s.relaunches} times`}{#if lastRelaunchReason(s)}, last because {lastRelaunchReason(s)}{/if}.
								{/if}
								{#if s.stop_reason}
									Stopped: {word('stopReason', s.stop_reason).toLowerCase()}{#if s.stopped_at}, {formatWhen(s.stopped_at, poll.now)}{/if}.
								{/if}
							</span>
						</li>
					{/each}
				</ul>
			{/if}
		</section>

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
