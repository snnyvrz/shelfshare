<script lang="ts">
    import CopyCard from "$lib/components/CopyCard.svelte";
    import type { PageProps } from "./$types";
    let { data, form }: PageProps = $props();
</script>

<svelte:head><title>Readers' shelves · ShelfShare</title></svelte:head>
<section class="section">
    <p class="eyebrow">REAL BOOKS. REAL CONNECTIONS.</p>
    <h1>Readers' shelves</h1>
    <p>
        Discover physical books shared by other readers. Request a copy, arrange your exchange, and return it for the
        next reader.
    </p>
    <div class="actions">
        <a class="button" href="/my-shelf">Add my books</a><a href="/books">Browse catalog titles →</a>
    </div>
    <h2 class="section">Find a reader</h2>
    <form method="GET" class="actions">
        <label
            >Display name or general location<input
                name="q"
                value={data.q}
                maxlength="100"
                placeholder="Find someone nearby"
            /></label
        ><button class="button secondary">Find readers</button>
    </form>
    <div class="copy-grid">
        {#each data.readers as reader (reader.id)}<article class="panel">
                <h3><a href="/shelves/{reader.id}">{reader.displayName}</a></h3>
                <p class="small muted">{reader.location}</p>
                <p class="clamp">{reader.bio || "A reader with stories to share."}</p>
                <a href="/shelves/{reader.id}">View shelf and message →</a>
            </article>{/each}
    </div>
    {#if !data.readers.length}<p>No readers match this search.</p>{/if}
    <div class="actions">
        {#if data.readersPage > 1}<a href="?q={encodeURIComponent(data.q)}&readersPage={data.readersPage - 1}"
                >← Previous readers</a
            >{/if}{#if data.readersPage * 50 < data.readerTotal}<a
                href="?q={encodeURIComponent(data.q)}&readersPage={data.readersPage + 1}">Next readers →</a
            >{/if}
    </div>
    <h2 class="section">Physical books to discover</h2>
    {#if form?.message}<p class="alert" role="alert">{form.message}</p>{/if}
    {#if form?.success}<p class="panel" role="status">{form.success} <a href="/requests">View requests →</a></p>{/if}
    <div class="copy-grid">
        {#each data.copies as copy (copy.id)}<CopyCard
                {copy}
                owner={data.profiles[copy.ownerId]}
                userId={data.user?.id}
            />{/each}
    </div>
    {#if !data.copies.length}<div class="empty-state">
            <h2>Start a shelf worth sharing</h2>
            <p>Add your physical books and invite another reader into their next chapter.</p>
        </div>{/if}
    <div class="actions">
        {#if data.page > 1}<a href="?page={data.page - 1}">← Previous</a>{/if}{#if data.page * 50 < data.total}<a
                href="?page={data.page + 1}">Next →</a
            >{/if}
    </div>
</section>
