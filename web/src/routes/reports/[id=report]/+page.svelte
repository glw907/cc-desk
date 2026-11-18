<!--
@component
A report's page at `/reports/<id>`: who sent it and when, its title and summary, the calls made
along the way, what the plan missed, and a folded table of the round's passes. It draws nothing
until the first good snapshot arrives, then the not-found view when no report has the id.
-->
<script lang="ts">
	import { page } from '$app/state';
	import NotFound from '#lib/NotFound.svelte';
	import { poll } from '#lib/poll.svelte.ts';
	import { formatMinutes, formatWhen } from '#lib/time.ts';
	import { word } from '#lib/words.ts';

	const snapshot = $derived(poll.answer?.snapshot ?? null);
	const report = $derived(snapshot?.reports.find((r) => r.id === page.params.id));
</script>

{#if snapshot && !report}
	<NotFound address={page.url.pathname} />
{:else if report}
	<main class="page">
		<div class="page-main">
			<p class="doc-meta muted">
				A report from {report.project}{#if report.round}, round {report.round}{/if}. Sent
				{formatWhen(report.sent, poll.now)}.
			</p>
			<h1 class="doc-title">{report.title}</h1>

			<div class="prose-doc mt-6">
				<p>{report.summary}</p>

				<h2>Calls made along the way</h2>
				{#if report.calls.length > 0}
					<ul class="calls">
						{#each report.calls as call (call.call)}
							<li>
								<strong>{call.call}.</strong> Chose {call.chosen}, not {call.other}. {call.why}
							</li>
						{/each}
					</ul>
				{:else}
					<p>No judgment calls were made along the way.</p>
				{/if}

				<h2>What the plan missed</h2>
				{#if report.misses.length > 0}
					<ul class="calls">
						{#each report.misses as miss (miss.miss)}
							<li>{miss.miss} {miss.cost}</li>
						{/each}
					</ul>
				{:else}
					<p>The plan missed nothing.</p>
				{/if}

				<details>
					<summary>The detail: each pass, its pull request, and its minutes</summary>
					<table class="table">
						<thead>
							<tr>
								<th scope="col">Pass</th>
								<th scope="col">State</th>
								<th scope="col">Pull request</th>
								<th scope="col">Minutes</th>
							</tr>
						</thead>
						<tbody>
							{#each report.passes as pass (pass.name)}
								<tr>
									<td>{pass.name}</td>
									<td><span class="state {pass.state}">{word('passState', pass.state)}</span></td>
									<td>{pass.pull_request === null ? 'None' : `#${pass.pull_request}`}</td>
									<td>
										{pass.actual_minutes === null
											? `${formatMinutes(pass.planned_minutes)} planned`
											: `${formatMinutes(pass.actual_minutes)} of ${formatMinutes(pass.planned_minutes)} planned`}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</details>
			</div>
		</div>
	</main>
{/if}
