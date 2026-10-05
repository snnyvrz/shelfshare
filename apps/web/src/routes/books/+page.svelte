<script lang="ts">
    import { useI18n } from "$lib/i18n/context";
    import DateField from "$lib/components/DateField.svelte";
    const { t, n } = useI18n();
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

<svelte:head><title>{t("The collection · ShelfShare")}</title></svelte:head>
<div class="page-heading">
    <div>
        <p class="eyebrow">{t("FIND YOUR NEXT CHAPTER")}</p>
        <h1>{t("The collection")}</h1>
        <p class="muted">{t("Stories to get lost in. Ideas to take with you.")}</p>
    </div>
    <a class="button" href="/books/new">{t("Add a book +")}</a>
</div>
<form class="filter-panel panel" method="GET">
    <div class="search-field">
        <label for="q">{t("Search the shelf")}</label><input
            id="q"
            name="q"
            type="search"
            placeholder={t("A title, a phrase, a new discovery…")}
            value={data.filters.q ?? ""}
        />
    </div>
    <div>
        <label for="sort">{t("Sort by")}</label><select
            id="sort"
            name="sort"
            value={data.filters.sort ?? "created_at_desc"}
            ><option value="created_at_desc">{t("Recently added")}</option><option value="created_at_asc"
                >{t("Oldest additions")}</option
            ><option value="title_asc">{t("Title: A–Z")}</option><option value="title_desc">{t("Title: Z–A")}</option
            ><option value="published_at_desc">{t("Newest publication")}</option><option value="published_at_asc"
                >{t("Oldest publication")}</option
            ></select
        >
    </div>
    <button class="filter-submit">{t("Search")}</button>
    <details
        class="advanced-filters"
        open={!!(data.filters.author_id || data.filters.published_after || data.filters.published_before)}
    >
        <summary>{t("Author & publication filters")}</summary>
        <div class="filter-row">
            <div>
                <label for="author_id">{t("Author")}</label><select
                    id="author_id"
                    name="author_id"
                    value={data.filters.author_id ?? ""}
                    ><option value="">{t("All authors")}</option>{#each data.authors as author (author.id)}<option
                            value={author.id}>{author.name}</option
                        >{/each}</select
                >
            </div>
            <div>
                <label for="published_after">{t("Published from")}</label><DateField
                    id="published_after"
                    name="published_after"
                    value={data.filters.published_after ?? ""}
                />
            </div>
            <div>
                <label for="published_before">{t("Published through")}</label><DateField
                    id="published_before"
                    name="published_before"
                    value={data.filters.published_before ?? ""}
                />
            </div>
        </div>
    </details>
</form>
<div class="results-heading">
    <p class="muted">
        {t(data.books.pagination.total === 1 ? "{count} book on this shelf" : "{count} books on this shelf", {
            count: n(data.books.pagination.total),
        })}
    </p>
    {#if data.filters.q || data.filters.author_id || data.filters.published_after || data.filters.published_before}<a
            href="/books">{t("Clear filters")}</a
        >{/if}
</div>
{#if data.books.data.length}<div class="book-grid">
        {#each data.books.data as book (book.id)}<BookCard {book} />{/each}
    </div>
{:else}<div class="empty-state">
        <h2>{t("No books on this shelf yet")}</h2>
        <p>{t("Try a different search, or add a book to the collection.")}</p>
        <div class="actions">
            <a class="button secondary" href="/books">{t("Reset search")}</a><a class="button" href="/books/new"
                >{t("Add a book")}</a
            >
        </div>
    </div>{/if}
{#if data.books.pagination.total_pages > 1}
    <nav class="pagination" aria-label={t("Book pagination")}>
        {#if data.books.pagination.page > 1}<a class="button secondary" href={pageLink(data.books.pagination.page - 1)}
                >{t("← Previous")}</a
            >{/if}<span
            >{t("Page {page} of {total}", {
                page: n(data.books.pagination.page),
                total: n(data.books.pagination.total_pages),
            })}</span
        >{#if data.books.pagination.page < data.books.pagination.total_pages}<a
                class="button secondary"
                href={pageLink(data.books.pagination.page + 1)}>{t("Next →")}</a
            >{/if}
    </nav>
{/if}
