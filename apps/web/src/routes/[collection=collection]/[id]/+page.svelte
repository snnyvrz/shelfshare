<script lang="ts">
    import { useI18n } from "$lib/i18n/context";
    const { t, n, date } = useI18n();
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
<a class="back-link" href="/{data.collection}"
    >{t(data.collection === "books" ? "← Back to books" : "← Back to authors")}</a
>
{#if form?.message}<div class="alert" role="alert">{t(form.message)}</div>{/if}
{#if form?.success}<p class="panel" role="status">
        {t(form.success)} <a href="/requests">{t("View requests →")}</a>
    </p>{/if}
<section class="detail-layout">
    <aside>
        {#if data.book}<BookCard book={data.book} />{:else}<div class="author-portrait" aria-hidden="true">
                {name.slice(0, 1)}
            </div>{/if}
    </aside>
    <div class="detail-copy">
        <p class="eyebrow">{t(data.book ? "FROM THE COLLECTION" : "BEHIND THE PAGES")}</p>
        <h1 dir="auto">{name}</h1>
        {#if data.book}<p class="byline">
                {t("by")} <a dir="auto" href="/authors/{data.book.author.id}">{data.book.author.name}</a>
            </p>
            <p class="small muted">
                {data.book.published_at
                    ? t("Published {date}", { date: date(data.book.published_at) })
                    : t("Publication date not listed")}
            </p>
            <hr />
            <h2>{t("About this book")}</h2>
            <p class="prose" dir="auto">
                {data.book.description ||
                    t("This story is still waiting for a description. Help the next reader discover it.")}
            </p>
        {:else if data.author}<h2>{t("A little about the author")}</h2>
            <p class="prose" dir="auto">{data.author.bio || t("A biography has not been added yet.")}</p>{/if}
        {#if data.user}<div class="actions detail-actions">
                <a class="button secondary" href="/{data.collection}/{data.book?.id ?? data.author?.id}/edit"
                    >{t(noun === "book" ? "Edit book" : "Edit author")}</a
                ><button class="text-button danger" onclick={() => (deleting = true)}
                    >{t(noun === "book" ? "Delete book" : "Delete author")}</button
                >
            </div>{:else}<p class="small muted">
                <a
                    href="/login?returnTo={encodeURIComponent(
                        `/${data.collection}/${data.book?.id ?? data.author?.id}`
                    )}">{t("Log in")}</a
                >
                {t("to contribute to this page.")}
            </p>{/if}
        {#if deleting}
            <div class="delete-confirm panel" role="region" aria-label={t("Confirm deletion")}>
                <h3>{t("Delete “{name}”?", { name })}</h3>
                <p>
                    {t(
                        noun === "book"
                            ? "This will remove the book from the shared collection."
                            : "This will remove the author from the shared collection."
                    )}{data.author ? t(" An author with books may need their books removed or reassigned first.") : ""}
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
                        <button class="danger-button" disabled={pending}
                            >{t(pending ? "Deleting…" : "Yes, delete")}</button
                        ><button type="button" class="button secondary" onclick={() => (deleting = false)}
                            >{t("Keep it")}</button
                        >
                    </div>
                </form>
            </div>
        {/if}
    </div>
</section>
{#if data.book}<section class="section">
        <div class="section-heading">
            <h2>{t("Readers' physical copies")}</h2>
            <a class="button" href="/my-shelf?book={data.book.id}">{t("Add my copy")}</a>
        </div>
        <div class="copy-grid">
            {#each data.copies as copy (copy.id)}<CopyCard
                    {copy}
                    owner={data.profiles[copy.ownerId]}
                    userId={data.user?.id}
                />{/each}
        </div>
        {#if !data.copies.length}<p class="empty-state">
                {t("No readers have listed a physical copy yet. Add yours to start sharing.")}
            </p>{/if}
    </section>{/if}
{#if data.author}<section class="section">
        <div class="section-heading">
            <h2>{t("On the shelf")}</h2>
            <span class="muted">{t("{count} books", { count: n(data.author.books?.length ?? 0) })}</span>
        </div>
        <div class="book-grid">
            {#each data.author.books ?? [] as book (book.id)}<BookCard book={{ ...book, author: data.author }} />{/each}
        </div>
        {#if !data.author.books?.length}<div class="empty-state">
                <p>{t("No books by this author yet.")}</p>
                <a class="button" href="/books/new">{t("Add a book")}</a>
            </div>{/if}
    </section>{/if}
