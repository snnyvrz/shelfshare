<script lang="ts">
    import { page } from "$app/state";
</script>

<svelte:head><title>{page.status} · ShelfShare</title></svelte:head>
<div class="empty-state error-page">
    <p class="eyebrow">{page.status}</p>
    <h1>{page.status === 404 ? "This page has left the shelf." : "A little pause in the story."}</h1>
    <p>{page.error?.message ?? "Something went wrong. Please try again."}</p>
    <div class="actions">
        <a class="button" href="/">Back home</a><a class="button secondary" href={page.url.pathname + page.url.search}
            >Try again</a
        >
    </div>
</div>
