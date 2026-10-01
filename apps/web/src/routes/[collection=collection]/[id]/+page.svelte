<script lang="ts">
    import { enhance } from "$app/forms";
    import BookCard from "$lib/components/BookCard.svelte";
    import CopyCard from "$lib/components/CopyCard.svelte";
    import type { PageProps } from "./$types";
    let { data, form }: PageProps = $props();
    let deleting = $state(false);
    let pending = $state(false);
    let name = $derived(data.book?.title ?? data.author?.name ?? "");
    let noun = $derived(data.book ? "book" : "author");
</script>

<svelte:head><title>{name} · ShelfShare</title></svelte:head>
<a class="back-link" href="/{data.collection}">← Back to {data.collection}</a>
{#if form?.message}<div class="alert" role="alert">{form.message}</div>{/if}
{#if form?.success}<p class="panel" role="status">{form.success} <a href="/requests">View requests →</a></p>{/if}
<section class="detail-layout">
    <aside>
        {#if data.book}<BookCard book={data.book} />{:else}<div class="author-portrait" aria-hidden="true">
                {name.slice(0, 1)}
            </div>{/if}
    </aside>
    <div class="detail-copy">
        <p class="eyebrow">{data.book ? "FROM THE COLLECTION" : "BEHIND THE PAGES"}</p>
        <h1>{name}</h1>
        {#if data.book}<p class="byline">by <a href="/authors/{data.book.author.id}">{data.book.author.name}</a></p>
            <p class="small muted">
                {data.book.published_at ? `Published ${data.book.published_at}` : "Publication date not listed"}
            </p>
            <hr />
            <h2>About this book</h2>
            <p class="prose">
                {data.book.description ||
                    "This story is still waiting for a description. Help the next reader discover it."}
            </p>
        {:else if data.author}<h2>A little about the author</h2>
            <p class="prose">{data.author.bio || "A biography has not been added yet."}</p>{/if}
        {#if data.user}<div class="actions detail-actions">
                <a class="button secondary" href="/{data.collection}/{data.book?.id ?? data.author?.id}/edit"
                    >Edit {noun}</a
                ><button class="text-button danger" onclick={() => (deleting = true)}>Delete {noun}</button>
            </div>{:else}<p class="small muted">
                <a
                    href="/login?returnTo={encodeURIComponent(
                        `/${data.collection}/${data.book?.id ?? data.author?.id}`
                    )}">Log in</a
                > to contribute to this page.
            </p>{/if}
        {#if deleting}
            <div class="delete-confirm panel" role="region" aria-label="Confirm deletion">
                <h3>Delete “{name}”?</h3>
                <p>
                    This will remove the {noun} from the shared collection.{data.author
                        ? " An author with books may need their books removed or reassigned first."
                        : ""}
                </p>
                <form
                    method="POST"
                    action="?/delete"
                    use:enhance={() => {
                        pending = true;
                        return async ({ update }) => {
                            try {
                                await update();
                            } finally {
                                pending = false;
                            }
                        };
                    }}
                >
                    <div class="actions">
                        <button class="danger-button" disabled={pending}>{pending ? "Deleting…" : "Yes, delete"}</button
                        ><button type="button" class="button secondary" onclick={() => (deleting = false)}
                            >Keep it</button
                        >
                    </div>
                </form>
            </div>
        {/if}
    </div>
</section>
{#if data.book}<section class="section">
        <div class="section-heading">
            <h2>Readers' physical copies</h2>
            <a class="button" href="/my-shelf?book={data.book.id}">Add my copy</a>
        </div>
        <div class="copy-grid">
            {#each data.copies as copy (copy.id)}<CopyCard
                    {copy}
                    owner={data.profiles[copy.ownerId]}
                    userId={data.user?.id}
                />{/each}
        </div>
        {#if !data.copies.length}<p class="empty-state">
                No readers have listed a physical copy yet. Add yours to start sharing.
            </p>{/if}
    </section>{/if}
{#if data.author}<section class="section">
        <div class="section-heading">
            <h2>On the shelf</h2>
            <span class="muted">{data.author.books?.length ?? 0} books</span>
        </div>
        <div class="book-grid">
            {#each data.author.books ?? [] as book (book.id)}<BookCard book={{ ...book, author: data.author }} />{/each}
        </div>
        {#if !data.author.books?.length}<div class="empty-state">
                <p>No books by this author yet.</p>
                <a class="button" href="/books/new">Add a book</a>
            </div>{/if}
    </section>{/if}
