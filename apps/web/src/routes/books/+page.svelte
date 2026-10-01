<script lang="ts">
    import BookCard from "$lib/components/BookCard.svelte";
    import type { PageProps } from "./$types";
    let { data }: PageProps = $props();
    function pageLink(number: number) {
        const query = new URLSearchParams(data.filters);
        query.set("page", String(number));
        query.delete("page_size");
        return `/books?${query}`;
    }
</script>

<svelte:head><title>The collection · ShelfShare</title></svelte:head>
<div class="page-heading">
    <div>
        <p class="eyebrow">FIND YOUR NEXT CHAPTER</p>
        <h1>The collection</h1>
        <p class="muted">Stories to get lost in. Ideas to take with you.</p>
    </div>
    <a class="button" href="/books/new">Add a book +</a>
</div>
<form class="filter-panel panel" method="GET">
    <div class="search-field">
        <label for="q">Search the shelf</label><input
            id="q"
            name="q"
            type="search"
            placeholder="A title, a phrase, a new discovery…"
            value={data.filters.q ?? ""}
        />
    </div>
    <div>
        <label for="sort">Sort by</label><select id="sort" name="sort" value={data.filters.sort ?? "created_at_desc"}
            ><option value="created_at_desc">Recently added</option><option value="created_at_asc"
                >Oldest additions</option
            ><option value="title_asc">Title: A–Z</option><option value="title_desc">Title: Z–A</option><option
                value="published_at_desc">Newest publication</option
            ><option value="published_at_asc">Oldest publication</option></select
        >
    </div>
    <button class="filter-submit">Search</button>
    <details
        class="advanced-filters"
        open={!!(data.filters.author_id || data.filters.published_after || data.filters.published_before)}
    >
        <summary>Author & publication filters</summary>
        <div class="filter-row">
            <div>
                <label for="author_id">Author</label><select
                    id="author_id"
                    name="author_id"
                    value={data.filters.author_id ?? ""}
                    ><option value="">All authors</option>{#each data.authors as author (author.id)}<option
                            value={author.id}>{author.name}</option
                        >{/each}</select
                >
            </div>
            <div>
                <label for="published_after">Published from</label><input
                    id="published_after"
                    name="published_after"
                    type="date"
                    value={data.filters.published_after ?? ""}
                />
            </div>
            <div>
                <label for="published_before">Published through</label><input
                    id="published_before"
                    name="published_before"
                    type="date"
                    value={data.filters.published_before ?? ""}
                />
            </div>
        </div>
    </details>
</form>
<div class="results-heading">
    <p class="muted">
        {data.books.pagination.total}
        {data.books.pagination.total === 1 ? "book" : "books"} on this shelf
    </p>
    {#if data.filters.q || data.filters.author_id || data.filters.published_after || data.filters.published_before}<a
            href="/books">Clear filters</a
        >{/if}
</div>
{#if data.books.data.length}<div class="book-grid">
        {#each data.books.data as book (book.id)}<BookCard {book} />{/each}
    </div>
{:else}<div class="empty-state">
        <h2>No books on this shelf yet</h2>
        <p>Try a different search, or add a book to the collection.</p>
        <div class="actions">
            <a class="button secondary" href="/books">Reset search</a><a class="button" href="/books/new">Add a book</a>
        </div>
    </div>{/if}
{#if data.books.pagination.total_pages > 1}
    <nav class="pagination" aria-label="Book pagination">
        {#if data.books.pagination.page > 1}<a class="button secondary" href={pageLink(data.books.pagination.page - 1)}
                >← Previous</a
            >{/if}<span>Page {data.books.pagination.page} of {data.books.pagination.total_pages}</span
        >{#if data.books.pagination.page < data.books.pagination.total_pages}<a
                class="button secondary"
                href={pageLink(data.books.pagination.page + 1)}>Next →</a
            >{/if}
    </nav>
{/if}
