<script lang="ts">
    import BookCard from "$lib/components/BookCard.svelte";
    import type { PageProps } from "./$types";
    let { data }: PageProps = $props();
</script>

<svelte:head
    ><title>ShelfShare · A world of stories</title><meta
        name="description"
        content="Share the physical books on your shelf, borrow from other readers, and arrange exchanges through real-time messages."
    /></svelte:head
>
<section class="hero">
    <div class="hero-copy">
        <p class="eyebrow">YOUR NEXT CHAPTER STARTS HERE</p>
        <h1>A shared shelf.<br />A world of <em>stories.</em></h1>
        <p class="hero-description">
            Your next favorite might be on a neighbor's shelf. List the physical books you own, borrow from other
            readers, and arrange an exchange through real-time messages.
        </p>
        <div class="actions">
            <a class="button" href="/shelves">Explore readers' shelves <span aria-hidden="true">↗</span></a><a
                href="/my-shelf">Build my shelf →</a
            >
        </div>
        <p class="hero-note">Request a copy. Meet its reader. Return it for the next chapter.</p>
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
            <p>Put your physical books on your profile and share them with another reader.</p>
            <a class="button" href="/my-shelf">Add my first copy</a>
        </div>
    {/if}
</section>
<section class="community-banner">
    <div>
        <p class="eyebrow">A COLLECTION THAT BELONGS TO ALL OF US</p>
        <h2>Found a book worth sharing?</h2>
        <p>Add your physical copy, choose whether to lend it, and connect with its next reader.</p>
    </div>
    <a class="button" href="/my-shelf">Add my copy +</a>
</section>
