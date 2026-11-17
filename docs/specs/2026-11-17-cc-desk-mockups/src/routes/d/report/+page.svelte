<svelte:head><title>cairn has one list of checks and one runner</title></svelte:head>

<main class="page">
	<div class="page-main">
		<p class="doc-meta muted">A report from cairn. Round 1 of 17 on the route to the release. Sent 1 hour ago.</p>
		<h1 class="doc-title">cairn has one list of checks and one runner</h1>

		<div class="prose-doc mt-6">
			<p>Every check cairn runs is now named in one file, and one command runs them, on your machine and in CI alike. The three passes merged in 560 minutes against a plan of 600. Nothing a site owner sees changed.</p>

			<h2>Calls made along the way</h2>
			<ul class="calls">
				<li>The slow browser checks run only in CI by default. Running them locally added eleven minutes to every commit.</li>
				<li>A check that fails prints its own name and the command to rerun it, and nothing else.</li>
				<li>The old <code>npm run verify</code> stays for one release as a pointer to the new command, then goes.</li>
			</ul>

			<h2>What the plan missed</h2>
			<ul class="calls">
				<li>The plan counted 14 checks. There were 19, because five lived in a workflow file nobody had listed. Pass 2 ran 70 minutes over.</li>
				<li>The plan assumed CI could call the runner unchanged. CI needed a flag to skip the checks that need a display.</li>
			</ul>
			<p>Both misses came from planning off the docs and never listing the workflow files. The planner's checklist now lists them.</p>
		</div>

		<section class="decide">
			<button class="btn btn-primary">Mark as read</button>
			<p class="muted mt-3">Round 2 has already started. A note here reaches the desk.</p>
		</section>

		<details class="prose-doc muted">
			<summary>The detail: each pass, its pull request, and its minutes</summary>
			<p>Folded away unless you open it.</p>
		</details>
	</div>
</main>
