<script lang="ts">
    import type { PageProps } from "./$types";
    let { data }: PageProps = $props();
</script>

<svelte:head><title>Authors · ShelfShare</title></svelte:head>
<div class="page-heading">
    <div>
        <p class="eyebrow">THE PEOPLE BEHIND THE PAGES</p>
        <h1>Meet the authors</h1>
        <p class="muted">Every story starts with a voice. Discover the voices on our shelf.</p>
    </div>
    <a class="button" href="/authors/new">Add an author +</a>
</div>
<form method="GET" class="panel author-search">
    <label for="q">Find an author</label>
    <div class="actions">
        <input id="q" name="q" type="search" placeholder="Search names and biographies" value={data.q} /><button
            >Search</button
        >{#if data.q}<a href="/authors">Clear</a>{/if}
    </div>
</form>
<p class="muted">{data.authors.length} {data.authors.length === 1 ? "author" : "authors"}</p>
<div class="author-grid">
    {#each data.authors as author (author.id)}<article class="panel author-card">
            <div class="author-initial" aria-hidden="true">{author.name.slice(0, 1)}</div>
            <h2><a href="/authors/{author.id}">{author.name}</a></h2>
            <p class="clamp muted">{author.bio || "A voice in our growing collection."}</p>
            <a href="/authors/{author.id}">{author.books?.length ?? 0} books · Meet the author →</a>
        </article>{/each}
</div>
{#if !data.authors.length}<div class="empty-state">
        <h2>{data.q ? "No matching authors" : "A new chapter awaits"}</h2>
        <p>
            {data.q
                ? "Try a different name or clear your search."
                : "Add the first author to start growing the collection."}
        </p>
        <a class="button" href={data.q ? "/authors" : "/authors/new"}>{data.q ? "Clear search" : "Add an author"}</a>
    </div>{/if}
