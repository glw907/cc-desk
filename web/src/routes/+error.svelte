<!--
@component
The page for an address no route matches, such as `/a/b/c/d`. A 404 reads as not found; any other
status shows its message. Both link to the front page.
-->
<script lang="ts">
	import { page } from '$app/state';
</script>

<main class="page">
	<div class="page-main">
		<h1 class="doc-title">{page.status === 404 ? 'Not found' : `Error ${page.status}`}</h1>
		<div class="prose-doc mt-6">
			{#if page.status === 404}
				<p>Nothing is at <code>{page.url.pathname}</code>.</p>
			{:else}
				<p>{page.error?.message}</p>
			{/if}
			<p><a href="/">Go to the front page</a></p>
		</div>
	</div>
</main>
