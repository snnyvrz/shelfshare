<script lang="ts">
    import { useI18n } from "$lib/i18n/context";
    const { t } = useI18n();
    import { enhance } from "$app/forms";
    import type { PageProps } from "./$types";
    let { data, form }: PageProps = $props();
</script>

<svelte:head><title>{t("My shelf · ShelfShare")}</title></svelte:head>
<section class="section">
    <h1>{t("My physical books")}</h1>
    <p>{t("List the books you own, choose which to share, and keep track of your copies.")}</p>
    <div class="actions">
        <a href="/shelves/{data.user?.id}">{t("My public shelf →")}</a><a href="/account/profile">{t("Edit profile")}</a
        ><a href="/requests">{t("Requests and loans")}</a>
    </div>
    {#if form?.message}<p class="alert" role="alert">{t(form.message)}</p>{/if}{#if form?.success}<p role="status">
            {t(form.success)}
        </p>{/if}
    <div class="panel stack">
        <h2>{t("Add my copy")}</h2>
        <form method="GET" class="actions">
            <label
                >{t("Find a catalog title")}<input
                    dir="auto"
                    name="q"
                    value={data.q}
                    placeholder={t("Search title")}
                /></label
            ><button class="button secondary">{t("Search")}</button>
        </form>
        <form method="POST" action="?/add" use:enhance class="stack">
            <label
                >{t("Book")}<select name="bookId" required
                    ><option value="">{t("Choose a title")}</option>{#each data.books as book (book.id)}<option
                            value={book.id}
                            selected={book.id === data.selectedBook}>{book.title} — {book.author.name}</option
                        >{/each}</select
                ></label
            >
            <p class="small">
                {t("Missing a title?")} <a href="/books/new?returnTo=/my-shelf">{t("Add catalog details first")}</a>.
                <a href="/authors/new">{t("Add a missing author")}</a>.
            </p>
            <label
                >{t("Condition")}<input
                    dir="auto"
                    name="condition"
                    required
                    maxlength="120"
                    placeholder={t("Good, with a few pencil notes")}
                /></label
            >
            <label
                >{t("Copy notes")}<textarea
                    dir="auto"
                    name="notes"
                    maxlength="2000"
                    placeholder={t("Edition, language, or anything another reader should know")}></textarea></label
            >
            <label class="check-label"
                ><input type="checkbox" name="visible" checked />{t("Show on my public shelf")}</label
            >
            <label class="check-label"><input type="checkbox" name="lendable" checked />{t("Offer to lend")}</label>
            <button class="button">{t("Add my copy")}</button>
        </form>
    </div>
    <h2 class="section">{t("On my shelf")}</h2>
    <div class="copy-grid">
        {#each data.copies as copy (copy.id)}<article class="panel stack">
                <h3><a dir="auto" href="/books/{copy.bookId}">{copy.book.Title}</a></h3>
                <p class="eyebrow">{t(copy.availability)}{!copy.visible ? t(" · hidden") : ""}</p>
                <form method="POST" action="?/edit" use:enhance class="stack">
                    <input type="hidden" name="id" value={copy.id} />
                    <label
                        >{t("Condition")}<input
                            dir="auto"
                            name="condition"
                            value={copy.condition}
                            required
                            maxlength="120"
                        /></label
                    >
                    <label>{t("Notes")}<textarea dir="auto" name="notes" maxlength="2000">{copy.notes}</textarea></label
                    >
                    <label class="check-label"
                        ><input type="checkbox" name="visible" checked={copy.visible} />{t("Public")}</label
                    >
                    <label class="check-label"
                        ><input type="checkbox" name="lendable" checked={copy.lendable} />{t("Offer to lend")}</label
                    >
                    <button class="button secondary">{t("Save copy")}</button>
                </form>
                <form method="POST" action="?/archive" use:enhance>
                    <input type="hidden" name="id" value={copy.id} /><button class="text-button danger"
                        >{t("Archive copy")}</button
                    >
                    >
                </form>
            </article>{/each}
    </div>
    {#if !data.copies.length}<p class="empty-state">{t("Your shelf is ready for its first physical book.")}</p>{/if}
    <div class="actions">
        {#if data.page > 1}<a href="?page={data.page - 1}">{t("← Previous")}</a>{/if}{#if data.page * 50 < data.total}<a
                href="?page={data.page + 1}">{t("Next →")}</a
            >
            >{/if}
    </div>
</section>
