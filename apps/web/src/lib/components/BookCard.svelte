<script lang="ts">
    import { useI18n } from "$lib/i18n/context";
    const { t, date } = useI18n();
    import type { Book } from "$lib/types";
    let { book }: { book: Book } = $props();
    const tones = ["sage", "clay", "blue", "gold"];
    let tone = $derived(tones[Array.from(book.id).reduce((sum, char) => sum + char.charCodeAt(0), 0) % tones.length]);
</script>

<article class="book-card">
    <a class="book-cover {tone}" href="/books/{book.id}" aria-label={t("Read about {title}", { title: book.title })}>
        <span class="cover-top">{t("SHELFSHARE COLLECTION")}</span>
        <strong dir="auto">{book.title}</strong>
        <span class="cover-rule"></span>
        <span dir="auto">{book.author.name}</span>
    </a>
    <div class="book-caption">
        <h3><a dir="auto" href="/books/{book.id}">{book.title}</a></h3>
        <a class="muted" dir="auto" href="/authors/{book.author.id}">{book.author.name}</a>
        <p class="small muted">
            {book.published_at
                ? date(book.published_at, { month: undefined, day: undefined })
                : t("Publication date unknown")}
        </p>
    </div>
</article>
