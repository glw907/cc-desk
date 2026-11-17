<svelte:head><title>Approve the dubplate intake design?</title></svelte:head>

<main class="page">
	<div class="page-main">
		<p class="doc-meta muted">A design for dubplate. Sent 20 minutes ago.</p>
		<h1 class="doc-title">Approve the dubplate intake design?</h1>

		<div class="prose-doc mt-6">
			<p>Intake is how a record gets into dubplate. You drop a folder of files, and the server reads the tags, matches the release, and files it. Today that code is spread over three commands and trusts whatever tags it finds. This round makes it one path, with one place that decides what a file is.</p>
			<p>The round is five passes, about 2,100 minutes. Two things in it change what you see, and both are asked below.</p>

			<h2>Yours to decide</h2>

			<p>When intake can't match a folder to a release, it has to put the folder somewhere you'll find it.</p>
			<fieldset class="question">
				<label class="question-option"><input type="radio" name="q1" class="radio radio-sm radio-primary" checked /><span><strong>A "Needs a look" shelf in the app.</strong> The folder shows in the library with a marker, and you fix it there. <span class="badge badge-sm badge-outline">recommended</span></span></label>
				<label class="question-option"><input type="radio" name="q1" class="radio radio-sm radio-primary" /><span><strong>A folder on disk.</strong> Unmatched folders move to <code>unsorted/</code> and nothing shows in the app.</span></label>
			</fieldset>

			<p>The round can spend from this week's allowance or wait for the reset on November 20.</p>
			<fieldset class="question">
				<label class="question-option"><input type="radio" name="q2" class="radio radio-sm radio-primary" checked /><span><strong>Start now.</strong> It stops at 85 percent, like every other round. <span class="badge badge-sm badge-outline">recommended</span></span></label>
				<label class="question-option"><input type="radio" name="q2" class="radio radio-sm radio-primary" /><span><strong>Wait for the reset.</strong></span></label>
			</fieldset>

			<h2>Already decided</h2>
			<p>These three could have gone either way. Say so in a note if you would have chosen differently.</p>
			<ul class="calls">
				<li>A file whose tags disagree with its folder name follows the tags.</li>
				<li>A second copy of a release you already have is kept beside the first and never merged into it.</li>
				<li>Intake of a large folder shows one count and no per-file progress.</li>
			</ul>
		</div>

		<section class="decide">
			<div class="flex flex-wrap gap-3">
				<button class="btn btn-primary">Approve</button>
				<button class="btn">Send back with a note</button>
			</div>
			<p class="muted mt-3">On approve, the round starts. You hear from it next in its report.</p>
		</section>

		<details class="prose-doc muted">
			<summary>The full design: the five passes, the minutes, and the technical calls</summary>
			<p>Folded away unless you open it.</p>
		</details>
	</div>
</main>
