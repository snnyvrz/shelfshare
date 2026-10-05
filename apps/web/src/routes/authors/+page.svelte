<script lang="ts">
    import { useI18n } from "$lib/i18n/context";
    const { t, n } = useI18n();
    import type { PageProps } from "./$types";
    let { data }: PageProps = $props();
</script>

<svelte:head><title>{t("Authors · ShelfShare")}</title></svelte:head>
<div class="page-heading">
    <div>
        <p class="eyebrow">{t("THE PEOPLE BEHIND THE PAGES")}</p>
        <h1>{t("Meet the authors")}</h1>
        <p class="muted">{t("Every story starts with a voice. Discover the voices on our shelf.")}</p>
    </div>
    <a class="button" href="/authors/new">{t("Add an author +")}</a>
</div>
<form method="GET" class="panel author-search">
    <label for="q">{t("Find an author")}</label>
    <div class="actions">
        <input id="q" name="q" type="search" placeholder={t("Search names and biographies")} value={data.q} /><button
            >{t("Search")}</button
        >{#if data.q}<a href="/authors">{t("Clear")}</a>{/if}
    </div>
</form>
<p class="muted">
    {t(data.authors.length === 1 ? "{count} author" : "{count} authors", { count: n(data.authors.length) })}
</p>
<div class="author-grid">
    {#each data.authors as author (author.id)}<article class="panel author-card">
            <div class="author-initial" aria-hidden="true">{author.name.slice(0, 1)}</div>
            <h2><a dir="auto" href="/authors/{author.id}">{author.name}</a></h2>
            <p class="clamp muted" dir="auto">{author.bio || t("A voice in our growing collection.")}</p>
            <a href="/authors/{author.id}"
                >{t("{count} books · Meet the author →", { count: n(author.books?.length ?? 0) })}</a
            >
        </article>{/each}
</div>
{#if !data.authors.length}<div class="empty-state">
        <h2>{t(data.q ? "No matching authors" : "A new chapter awaits")}</h2>
        <p>
            {t(
                data.q
                    ? "Try a different name or clear your search."
                    : "Add the first author to start growing the collection."
            )}
        </p>
        <a class="button" href={data.q ? "/authors" : "/authors/new"}>{t(data.q ? "Clear search" : "Add an author")}</a>
    </div>{/if}
