<script lang="ts">
    import { useI18n } from "$lib/i18n/context";
    const { t, n } = useI18n();
    import { page } from "$app/state";
</script>

<svelte:head><title>{n(page.status)} · ShelfShare</title></svelte:head>
<div class="empty-state error-page">
    <p class="eyebrow">{n(page.status)}</p>
    <h1>{t(page.status === 404 ? "This page has left the shelf." : "A little pause in the story.")}</h1>
    <p>{t(page.error?.message ?? "Something went wrong. Please try again.")}</p>
    <div class="actions">
        <a class="button" href="/">{t("Back home")}</a><a
            class="button secondary"
            href={page.url.pathname + page.url.search}>{t("Try again")}</a
        >
        >
    </div>
</div>
