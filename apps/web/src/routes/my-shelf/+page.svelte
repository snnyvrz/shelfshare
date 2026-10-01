<script lang="ts">
    import { enhance } from "$app/forms";
    import type { PageProps } from "./$types";
    let { data, form }: PageProps = $props();
</script>

<svelte:head><title>My shelf · ShelfShare</title></svelte:head>
<section class="section">
    <h1>My physical books</h1>
    <p>List the books you own, choose which to share, and keep track of your copies.</p>
    <div class="actions">
        <a href="/shelves/{data.user?.id}">My public shelf →</a><a href="/account/profile">Edit profile</a><a
            href="/requests">Requests and loans</a
        >
    </div>
    {#if form?.message}<p class="alert" role="alert">{form.message}</p>{/if}{#if form?.success}<p role="status">
            {form.success}
        </p>{/if}
    <div class="panel stack">
        <h2>Add my copy</h2>
        <form method="GET" class="actions">
            <label>Find a catalog title<input name="q" value={data.q} placeholder="Search title" /></label><button
                class="button secondary">Search</button
            >
        </form>
        <form method="POST" action="?/add" use:enhance class="stack">
            <label
                >Book<select name="bookId" required
                    ><option value="">Choose a title</option>{#each data.books as book (book.id)}<option
                            value={book.id}
                            selected={book.id === data.selectedBook}>{book.title} — {book.author.name}</option
                        >{/each}</select
                ></label
            >
            <p class="small">
                Missing a title? <a href="/books/new?returnTo=/my-shelf">Add catalog details first</a>.
                <a href="/authors/new">Add a missing author</a>.
            </p>
            <label
                >Condition<input
                    name="condition"
                    required
                    maxlength="120"
                    placeholder="Good, with a few pencil notes"
                /></label
            >
            <label
                >Copy notes<textarea
                    name="notes"
                    maxlength="2000"
                    placeholder="Edition, language, or anything another reader should know"></textarea></label
            >
            <label class="check-label"><input type="checkbox" name="visible" checked />Show on my public shelf</label>
            <label class="check-label"><input type="checkbox" name="lendable" checked />Offer to lend</label>
            <button class="button">Add my copy</button>
        </form>
    </div>
    <h2 class="section">On my shelf</h2>
    <div class="copy-grid">
        {#each data.copies as copy (copy.id)}<article class="panel stack">
                <h3><a href="/books/{copy.bookId}">{copy.book.Title}</a></h3>
                <p class="eyebrow">{copy.availability}{!copy.visible ? " · hidden" : ""}</p>
                <form method="POST" action="?/edit" use:enhance class="stack">
                    <input type="hidden" name="id" value={copy.id} />
                    <label>Condition<input name="condition" value={copy.condition} required maxlength="120" /></label>
                    <label>Notes<textarea name="notes" maxlength="2000">{copy.notes}</textarea></label>
                    <label class="check-label"
                        ><input type="checkbox" name="visible" checked={copy.visible} />Public</label
                    >
                    <label class="check-label"
                        ><input type="checkbox" name="lendable" checked={copy.lendable} />Offer to lend</label
                    >
                    <button class="button secondary">Save copy</button>
                </form>
                <form method="POST" action="?/archive" use:enhance>
                    <input type="hidden" name="id" value={copy.id} /><button class="text-button danger"
                        >Archive copy</button
                    >
                </form>
            </article>{/each}
    </div>
    {#if !data.copies.length}<p class="empty-state">Your shelf is ready for its first physical book.</p>{/if}
    <div class="actions">
        {#if data.page > 1}<a href="?page={data.page - 1}">← Previous</a>{/if}{#if data.page * 50 < data.total}<a
                href="?page={data.page + 1}">Next →</a
            >{/if}
    </div>
</section>
