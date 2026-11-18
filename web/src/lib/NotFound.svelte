<!--
@component
The page for an address that names nothing: a heading, the address, and a link home. The item
pages draw it only once a snapshot is in, so it never flashes before the first poll answers.
-->
<script lang="ts">
	interface Props {
		/** The address that matched nothing, shown as written. */
		address: string;
	}

	let { address }: Props = $props();
</script>

<main class="page">
	<div class="page-main">
		<h1 class="doc-title">Not found</h1>
		<div class="prose-doc mt-6">
			<p>Nothing is at <code>{address}</code>.</p>
			<p><a href="/">Go to the front page</a></p>
		</div>
	</div>
</main>
