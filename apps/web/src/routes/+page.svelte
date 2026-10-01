<script lang="ts">
    import BookCard from "$lib/components/BookCard.svelte";
    import type { PageProps } from "./$types";
    let { data }: PageProps = $props();
</script>

<svelte:head
    ><title>ShelfShare · A world of stories</title><meta
        name="description"
        content="Discover books, meet their authors, and help grow a shared collection of stories."
    /></svelte:head
>
<section class="hero">
    <div class="hero-copy">
        <p class="eyebrow">YOUR NEXT CHAPTER STARTS HERE</p>
        <h1>A shared shelf.<br />A world of <em>stories.</em></h1>
        <p class="hero-description">
            Good books deserve good company. Explore the collection, discover an author, and leave something wonderful
            for the next reader.
        </p>
        <div class="actions">
            <a class="button" href="/books">Explore the collection <span aria-hidden="true">↗</span></a><a
                href="/authors">Meet the authors →</a
            >
        </div>
        <p class="hero-note">An open collection, thoughtfully grown together.</p>
    </div>
    <div class="hero-art" aria-hidden="true">
        <div class="art-book art-book-one">
            <span>THE ART<br />OF FINDING<br /><i>Your next read</i></span><small>SHELFSHARE</small>
        </div>
        <div class="art-book art-book-two">
            <span>Between<br /><i>the pages</i></span><small>A WORLD TO DISCOVER</small>
        </div>
        <div class="art-shelf"></div>
        <span class="art-star">✳</span>
    </div>
</section>
<section class="section">
    <div class="section-heading">
        <div>
            <p class="eyebrow">FRESH ON THE SHELF</p>
            <h2>Recently added</h2>
        </div>
        <a href="/books">View the collection →</a>
    </div>
    {#if data.unavailable}
        <div class="empty-state">
            <h3>The shelf is taking a moment</h3>
            <p>The books service is unavailable. Please try again shortly.</p>
            <a class="button secondary" href="/">Try again</a>
        </div>
    {:else if data.books.length}
        <div class="book-grid">
            {#each data.books as book (book.id)}<BookCard {book} />{/each}
        </div>
    {:else}
        <div class="empty-state">
            <h3>Every library starts with one book</h3>
            <p>Help us write the first chapter of this collection.</p>
            <a class="button" href="/books/new">Add the first book</a>
        </div>
    {/if}
</section>
<section class="community-banner">
    <div>
        <p class="eyebrow">A COLLECTION THAT BELONGS TO ALL OF US</p>
        <h2>Found a book worth sharing?</h2>
        <p>Add its story to the shelf. Someone's next favorite could be yours.</p>
    </div>
    <a class="button" href="/books/new">Add a book +</a>
</section>
