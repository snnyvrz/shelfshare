<script lang="ts">
    import { enhance } from "$app/forms";
    import CopyCard from "$lib/components/CopyCard.svelte";
    import type { PageProps } from "./$types";
    let { data, form }: PageProps = $props();
</script>

<svelte:head><title>{data.profile.displayName}'s shelf · ShelfShare</title></svelte:head>
<section class="section">
    <a class="back-link" href="/shelves">← Readers' shelves</a>
    <h1>{data.profile.displayName}'s shelf</h1>
    {#if data.profile.location}<p class="muted">{data.profile.location}</p>{/if}
    <p class="message-body">{data.profile.bio || "A reader with stories to share."}</p>
    {#if data.user?.id === data.profile.id}<a class="button secondary" href="/account/profile">Edit profile</a>
    {:else if data.user}<form method="POST" action="?/message" use:enhance>
            <button class="button">Message {data.profile.displayName}</button>
        </form>
    {:else}<a class="button" href="/login?returnTo={encodeURIComponent(`/shelves/${data.profile.id}`)}"
            >Log in to message</a
        >{/if}
    {#if form?.message}<p class="alert" role="alert">{form.message}</p>{/if}
    {#if form?.success}<p class="panel" role="status">{form.success} <a href="/requests">View requests →</a></p>{/if}
    <div class="copy-grid">
        {#each data.copies as copy (copy.id)}<CopyCard {copy} owner={data.profile} userId={data.user?.id} />{/each}
    </div>
    {#if !data.copies.length}<p class="empty-state">No public copies on this shelf yet.</p>{/if}
    <div class="actions">
        {#if data.page > 1}<a href="?page={data.page - 1}">← Previous</a>{/if}{#if data.page * 50 < data.total}<a
                href="?page={data.page + 1}">Next →</a
            >{/if}
    </div>
</section>
