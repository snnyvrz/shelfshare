<script lang="ts">
    import { useI18n } from "$lib/i18n/context";
    const { t } = useI18n();
    import BookCard from "$lib/components/BookCard.svelte";
    import type { PageProps } from "./$types";
    let { data }: PageProps = $props();
</script>

<svelte:head
    ><title>{t("ShelfShare · A world of stories")}</title><meta
        name="description"
        content={t(
            "Share the physical books on your shelf, borrow from other readers, and arrange exchanges through real-time messages."
        )}
    /></svelte:head
>
<section class="hero">
    <div class="hero-copy">
        <p class="eyebrow">{t("YOUR NEXT CHAPTER STARTS HERE")}</p>
        <h1>{t("A shared shelf.")}<br />{t("A world of")} <em>{t("stories.")}</em></h1>
        <p class="hero-description">
            {t(
                "Your next favorite might be on a neighbor's shelf. List the physical books you own, borrow from other readers, and arrange an exchange through real-time messages."
            )}
        </p>
        <div class="actions">
            <a class="button" href="/shelves"
                >{t("Explore readers' shelves")} <span class="direction-arrow" aria-hidden="true">↗</span></a
            ><a href="/my-shelf">{t("Build my shelf →")}</a>
        </div>
        <p class="hero-note">{t("Request a copy. Meet its reader. Return it for the next chapter.")}</p>
    </div>
    <div class="hero-art" aria-hidden="true">
        <div class="art-book art-book-one">
            <span>{t("THE ART")}<br />{t("OF FINDING")}<br /><i>{t("Your next read")}</i></span><small>SHELFSHARE</small
            >
        </div>
        <div class="art-book art-book-two">
            <span>{t("Between")}<br /><i>{t("the pages")}</i></span><small>{t("A WORLD TO DISCOVER")}</small>
        </div>
        <div class="art-shelf"></div>
        <span class="art-star">✳</span>
    </div>
</section>
<section class="section">
    <div class="section-heading">
        <div>
            <p class="eyebrow">{t("FRESH ON THE SHELF")}</p>
            <h2>{t("Recently added")}</h2>
        </div>
        <a href="/books">{t("View the collection →")}</a>
    </div>
    {#if data.unavailable}
        <div class="empty-state">
            <h3>{t("The shelf is taking a moment")}</h3>
            <p>{t("The books service is unavailable. Please try again shortly.")}</p>
            <a class="button secondary" href="/">{t("Try again")}</a>
        </div>
    {:else if data.books.length}
        <div class="book-grid">
            {#each data.books as book (book.id)}<BookCard {book} />{/each}
        </div>
    {:else}
        <div class="empty-state">
            <h3>{t("Every library starts with one book")}</h3>
            <p>{t("Put your physical books on your profile and share them with another reader.")}</p>
            <a class="button" href="/my-shelf">{t("Add my first copy")}</a>
        </div>
    {/if}
</section>
<section class="community-banner">
    <div>
        <p class="eyebrow">{t("A COLLECTION THAT BELONGS TO ALL OF US")}</p>
        <h2>{t("Found a book worth sharing?")}</h2>
        <p>{t("Add your physical copy, choose whether to lend it, and connect with its next reader.")}</p>
    </div>
    <a class="button" href="/my-shelf">{t("Add my copy +")}</a>
</section>
