<script lang="ts">
    import "$lib/styles.css";
    import { page, navigating } from "$app/state";
    import type { LayoutProps } from "./$types";
    import { invalidate } from "$app/navigation";
    import { startRealtime, unread, realtimeEvent } from "$lib/realtime";
    import { onMount } from "svelte";
    let { data, children }: LayoutProps = $props();
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

<a class="skip-link" href="#main">Skip to content</a>
<header class="site-header">
    <div class="header-inner">
        <a class="brand" href="/" aria-label="ShelfShare home"
            ><span class="brand-mark" aria-hidden="true">▥</span>ShelfShare<span class="brand-dot">.</span></a
        >
        <nav aria-label="Main navigation">
            <a class:active={page.url.pathname.startsWith("/shelves")} href="/shelves">Readers' shelves</a>
            <a class:active={page.url.pathname.startsWith("/books")} href="/books">Book catalog</a>
            {#if data.user}
                <a href="/my-shelf">My shelf</a><a href="/requests">Requests</a>
                <a href="/messages"
                    >Messages{#if $unread}<span class="unread-badge">{$unread}</span>{/if}</a
                >
            {/if}
        </nav>
        <div class="account-nav">
            {#if data.user}
                <a href="/account/profile">My profile</a>
                <form action="/logout" method="POST"><button class="text-button">Log out</button></form>
            {:else}
                <a href="/login">Log in</a><a class="button small-button" href="/register">Join the shelf</a>
            {/if}
        </div>
    </div>
</header>
<main id="main" class="container">{@render children()}</main>
{#if notification}<aside class="message-notification" role="status">
        <a href="/messages?conversation={notification}" onclick={() => (notification = null)}
            >New message — open conversation →</a
        ><button class="text-button" aria-label="Dismiss message notification" onclick={() => (notification = null)}
            >×</button
        >
    </aside>{/if}
<footer class="site-footer">
    <div>
        <a class="brand" href="/">ShelfShare.</a>
        <p>Physical books. Shared stories. Connected readers.</p>
    </div>
    <p class="small">Made for readers, by readers.</p>
</footer>
