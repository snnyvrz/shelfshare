<script lang="ts">
    import "$lib/styles.css";
    import { page } from "$app/state";
    import type { LayoutProps } from "./$types";
    let { data, children }: LayoutProps = $props();
</script>

<a class="skip-link" href="#main">Skip to content</a>
<header class="site-header">
    <div class="header-inner">
        <a class="brand" href="/" aria-label="ShelfShare home"
            ><span class="brand-mark" aria-hidden="true">▥</span>ShelfShare<span class="brand-dot">.</span></a
        >
        <nav aria-label="Main navigation">
            <a class:active={page.url.pathname.startsWith("/books")} href="/books">The collection</a>
            <a class:active={page.url.pathname.startsWith("/authors")} href="/authors">Authors</a>
        </nav>
        <div class="account-nav">
            {#if data.user}
                <span class="account-email" title={data.user.email}>{data.user.email}</span>
                <form action="/logout" method="POST"><button class="text-button">Log out</button></form>
            {:else}
                <a href="/login">Log in</a><a class="button small-button" href="/register">Join the shelf</a>
            {/if}
        </div>
    </div>
</header>
<main id="main" class="container">{@render children()}</main>
<footer class="site-footer">
    <div>
        <a class="brand" href="/">ShelfShare.</a>
        <p>A shared shelf. A world of stories.</p>
    </div>
    <p class="small">Made for readers, by readers.</p>
</footer>
