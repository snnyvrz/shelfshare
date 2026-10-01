<script lang="ts">
    import type { Book } from "$lib/types";
    let { book }: { book: Book } = $props();
    const tones = ["sage", "clay", "blue", "gold"];
    let tone = $derived(tones[Array.from(book.id).reduce((sum, char) => sum + char.charCodeAt(0), 0) % tones.length]);
</script>

<article class="book-card">
    <a class="book-cover {tone}" href="/books/{book.id}" aria-label="Read about {book.title}">
        <span class="cover-top">SHELFSHARE COLLECTION</span>
        <strong>{book.title}</strong>
        <span class="cover-rule"></span>
        <span>{book.author.name}</span>
    </a>
    <div class="book-caption">
        <h3><a href="/books/{book.id}">{book.title}</a></h3>
        <a class="muted" href="/authors/{book.author.id}">{book.author.name}</a>
        <p class="small muted">{book.published_at ? book.published_at.slice(0, 4) : "Publication date unknown"}</p>
    </div>
</article>
