<script lang="ts">
    import "$lib/styles.css";
    import { page, navigating } from "$app/state";
    import type { LayoutProps } from "./$types";
    import { invalidate } from "$app/navigation";
    import { startRealtime, unread, realtimeEvent } from "$lib/realtime";
    import { onMount } from "svelte";
    import { provideLocale, useI18n } from "$lib/i18n/context";
    import { direction, type Locale } from "$lib/i18n";
    let { data, children }: LayoutProps = $props();
    let selectedLocale = $state<Locale | null>(null);
    let locale = $derived(selectedLocale ?? data.locale);
    provideLocale(
        () => locale,
        () => data.cities
    );
    const { t, n } = useI18n();
    let switching = $state(false);
    let languageError = $state(false);
    async function switchLanguage(event: SubmitEvent) {
        event.preventDefault();
        if (switching) return;
        switching = true;
        languageError = false;
        try {
            const response = await fetch("/api/locale", {
                method: "POST",
                body: new FormData(event.currentTarget as HTMLFormElement),
                headers: { Accept: "application/json" },
            });
            if (response.ok) selectedLocale = locale === "fa" ? "en" : "fa";
            else languageError = true;
        } catch {
            languageError = true;
        } finally {
            switching = false;
        }
    }
    $effect(() => {
        document.documentElement.lang = locale;
        document.documentElement.dir = direction(locale);
    });
    let userId = $derived(data.user?.id);
    let notification = $state<string | null>(null);
    let seenMessage = "";
    $effect(() => {
        if (!userId) {
            notification = null;
            return;
        }
        const event = $realtimeEvent;
        if (!event?.data || event.event !== "message" || event.data.senderId === userId) return;
        if (seenMessage === event.data.id) return;
        seenMessage = String(event.data.id);
        const conversation = String(event.data.conversationId);
        if (page.url.pathname === "/messages" && page.url.searchParams.get("conversation") === conversation) return;
        notification = conversation;
    });
    onMount(() => {
        document.documentElement.dataset.hydrated = "true";
    });
    $effect(() => {
        if (userId)
            return startRealtime(() => {
                // A background refresh must not race an enhanced form redirect.
                if (!navigating.to && ["/messages", "/requests"].includes(page.url.pathname))
                    void invalidate("shelfshare:community");
            });
    });
</script>

<a class="skip-link" href="#main">{t("Skip to content")}</a>
<header class="site-header">
    <div class="header-inner">
        <a class="brand" href="/" aria-label={t("ShelfShare home")}
            ><span class="brand-mark" aria-hidden="true">▥</span>ShelfShare<span class="brand-dot">.</span></a
        >
        <nav aria-label={t("Main navigation")}>
            <a class:active={page.url.pathname.startsWith("/shelves")} href="/shelves">{t("Readers' shelves")}</a>
            <a class:active={page.url.pathname.startsWith("/books")} href="/books">{t("Book catalog")}</a>
            {#if data.user}
                <a href="/my-shelf">{t("My shelf")}</a><a href="/requests">{t("Requests")}</a>
                <a href="/messages"
                    >{t("Messages")}{#if $unread}<span class="unread-badge">{n($unread)}</span>{/if}</a
                >
            {/if}
        </nav>
        <div class="account-nav">
            <form method="POST" action="/api/locale" onsubmit={switchLanguage}>
                <input type="hidden" name="locale" value={locale === "fa" ? "en" : "fa"} />
                <input type="hidden" name="returnTo" value={page.url.pathname + page.url.search} />
                <button
                    class="text-button language-switch"
                    disabled={switching}
                    aria-label={t("Language")}
                    lang={locale === "fa" ? "en" : "fa"}
                    dir="auto">{locale === "fa" ? "English" : "فارسی"}</button
                >
            </form>
            {#if data.user}
                <a href="/account/profile">{t("My profile")}</a>
                <form action="/logout" method="POST"><button class="text-button">{t("Log out")}</button></form>
            {:else}
                <a href="/login">{t("Log in")}</a><a class="button small-button" href="/register"
                    >{t("Join the shelf")}</a
                >
            {/if}
        </div>
    </div>
</header>
{#if languageError}<p class="alert container" role="alert">{t("Unable to change language. Please try again.")}</p>{/if}
<main id="main" class="container">{@render children()}</main>
{#if notification}<aside class="message-notification" role="status">
        <a href="/messages?conversation={notification}" onclick={() => (notification = null)}
            >{t("New message — open conversation →")}</a
        ><button
            class="text-button"
            aria-label={t("Dismiss message notification")}
            onclick={() => (notification = null)}>×</button
        >
    </aside>{/if}
<footer class="site-footer">
    <div>
        <a class="brand" href="/">ShelfShare.</a>
        <p>{t("Physical books. Shared stories. Connected readers.")}</p>
    </div>
    <p class="small">{t("Made for readers, by readers.")}</p>
</footer>
