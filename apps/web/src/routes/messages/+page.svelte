<script lang="ts">
    import { useI18n } from "$lib/i18n/context";
    const { t, n, date } = useI18n();
    const connectionLabels = {
        connected: "Connected",
        disconnected: "Disconnected",
        connecting: "Connecting…",
        reconnecting: "Reconnecting…",
        offline: "Offline",
        expired: "Session expired — log in again",
    };
    import { onDestroy } from "svelte";
    import { invalidateAll } from "$app/navigation";
    import { community, connection, realtimeEvent, sendEvent } from "$lib/realtime";
    import type { Conversation, Message } from "$lib/types";
    import type { PageProps } from "./$types";
    let { data }: PageProps = $props();
    let body = $state("");
    let error = $state("");
    let sending = $state(false);
    let retryMessage = $state<{ id: string; body: string; conversationId: string } | null>(null);
    let online = $state(false);
    let typing = $state(false);
    let readAt = $state("");
    let older = $state<Message[]>([]);
    let loadingOlder = $state(false);
    let typingTimer: ReturnType<typeof setTimeout>;
    let ackTimer: ReturnType<typeof setTimeout>;
    let lastTyping = 0;
    let selectedId = "";
    let lastRead = "";
    const peer = (c: Conversation) => (c.userA === data.user?.id ? c.userB : c.userA);
    let current = $derived(data.conversation);
    let messages = $derived(
        [...older, ...data.messages].filter((m, i, ms) => ms.findIndex((x) => x.id === m.id) === i)
    );
    let canSend = $derived(
        !!current &&
            current.status !== "declined" &&
            (current.requestId
                ? !["returned", "declined", "cancelled"].includes(current.request?.status || "")
                : !data.blocked.includes(peer(current))) &&
            (current.status === "accepted" || (current.initiator === data.user?.id && messages.length === 0))
    );

    $effect(() => {
        const c = current;
        if (!c) return;
        if (selectedId !== c.id) {
            selectedId = c.id;
            older = [];
            online = false;
            typing = false;
            readAt = c.peerReadAt || "";
            body = "";
            error = "";
            retryMessage = null;
            sending = false;
        }
        if ($connection === "connected") sendEvent({ event: "presence", conversationId: c.id });
        else {
            online = false;
            typing = false;
        }
        const key = c.id + ":" + (data.messages.at(-1)?.id || "empty");
        if (lastRead !== key) {
            lastRead = key;
            void community(`/conversations/${c.id}/read`, "POST", {}).catch(() => {
                lastRead = "";
            });
        }
    });
    $effect(() => {
        const e = $realtimeEvent;
        if (!e || !current) return;
        if (e.event === "blocked") {
            online = false;
            typing = false;
        }
        if (!e.data) return;
        if (e.data.conversationId !== current.id) return;
        if (e.event === "presence") online = !!e.data.online;
        if (e.event === "typing") {
            typing = true;
            clearTimeout(typingTimer);
            typingTimer = setTimeout(() => (typing = false), 3000);
        }
        if (e.event === "read" && e.data.userId !== data.user?.id) readAt = String(e.data.at);
        if (e.event === "ack" && retryMessage?.id === e.data.id) {
            clearTimeout(ackTimer);
            retryMessage = null;
            sending = false;
            body = "";
            error = "";
        }
    });
    $effect(() => {
        const e = $realtimeEvent;
        if (e?.event === "error" && e.data && retryMessage?.id === e.data.id) {
            clearTimeout(ackTimer);
            sending = false;
            error = String(e.data.message);
        }
    });
    onDestroy(() => {
        clearTimeout(typingTimer);
        clearTimeout(ackTimer);
    });

    async function action(path: string, method = "POST") {
        try {
            error = "";
            await community(path, method, method === "POST" ? {} : undefined);
            await invalidateAll();
        } catch (cause) {
            error = cause instanceof Error ? cause.message : "Unable to complete action";
        }
    }
    function send(event: SubmitEvent) {
        event.preventDefault();
        if (!current || !body.trim() || sending) return;
        if (!retryMessage || retryMessage.body !== body.trim() || retryMessage.conversationId !== current.id)
            retryMessage = { id: crypto.randomUUID(), body: body.trim(), conversationId: current.id };
        if (!sendEvent({ event: "message", ...retryMessage })) {
            error = "Reconnecting. Your message is ready to send when connected.";
            return;
        }
        sending = true;
        error = "";
        ackTimer = setTimeout(() => {
            sending = false;
            error = "Delivery not confirmed. Retry to safely send the same message.";
        }, 10000);
    }
    function signalTyping() {
        if (current && Date.now() - lastTyping > 1500) {
            lastTyping = Date.now();
            sendEvent({ event: "typing", conversationId: current.id });
        }
    }
    async function loadOlder() {
        if (!current || !messages.length) return;
        loadingOlder = true;
        try {
            const result = await community<{ data: Message[] }>(
                `/conversations/${current.id}/messages?before=${encodeURIComponent(messages[0].createdAt)}&beforeId=${messages[0].id}`
            );
            older = [...result.data.reverse(), ...older];
            if (!result.data.length) error = "You've reached the start of this conversation.";
        } catch (cause) {
            error = cause instanceof Error ? cause.message : "Unable to load history";
        } finally {
            loadingOlder = false;
        }
    }
</script>

<svelte:head><title>{t("Messages · ShelfShare")}</title></svelte:head>
<section class="section">
    <div class="section-heading">
        <h1>{t("Messages")}</h1>
        <span role="status" class="small muted">{t(connectionLabels[$connection])}</span>
    </div>
    {#if error}<p class="alert" role="alert">{t(error)}</p>{/if}
    <div class="inbox-layout">
        <aside class="panel inbox-list" aria-label={t("Conversations")}>
            {#each data.conversations as c (c.id)}<a
                    class:chosen={current?.id === c.id}
                    class="conversation-link"
                    href="/messages?conversation={c.id}&page={data.page}"
                >
                    <strong dir="auto">{data.profiles[peer(c)]?.displayName || t("Reader")}</strong>
                    <span class="small"
                        ><bdi>{c.request ? c.request.copy.book.Title : t("Direct message")}</bdi> · {t(c.status)}</span
                    >
                    {#if c.lastMessage}<span class="message-preview" dir="auto">{c.lastMessage.body.slice(0, 100)}</span
                        >{/if}
                    {#if c.unread && current?.id !== c.id}<span class="unread-badge"
                            >{t("{count} unread", { count: n(c.unread) })}</span
                        >{/if}
                </a>{/each}
            {#if !data.conversations.length}<p>
                    {t("No conversations yet.")} <a href="/shelves">{t("Find a reader →")}</a>
                </p>{/if}
            <div class="actions">
                {#if data.page > 1}<a href="?page={data.page - 1}">{t("← Previous")}</a
                    >{/if}{#if data.page * 50 < data.total}<a href="?page={data.page + 1}">{t("Next →")}</a>{/if}
            </div>
        </aside>
        <div class="panel conversation-panel">
            {#if current}
                <div class="section-heading">
                    <h2>
                        <a dir="auto" href="/shelves/{peer(current)}"
                            >{data.profiles[peer(current)]?.displayName || t("Reader")}</a
                        >
                    </h2>
                    {#if current.status === "accepted" && !data.blocked.includes(peer(current))}<span
                            class="small muted">{t(online ? "Online" : "Offline")}</span
                        >{/if}
                </div>
                {#if current.request}<p>
                        <a href="/requests"
                            ><bdi>{current.request.copy.book.Title}</bdi> · {t(current.request.status)} — {t(
                                "View loan actions →"
                            )}</a
                        >
                    </p>{/if}
                {#if current.status === "pending"}
                    {#if current.initiator !== data.user?.id}<p>
                            {t("Message request: accept to reply and share online status.")}
                        </p>
                        <div class="actions">
                            <button class="button" onclick={() => action(`/conversations/${current!.id}/accept`)}
                                >{t("Accept")}</button
                            ><button
                                class="button secondary"
                                onclick={() => action(`/conversations/${current!.id}/decline`)}>{t("Decline")}</button
                            >
                        </div>
                    {:else}<p class="small muted">
                            {t("Send one introduction. Further messages become available when the recipient accepts.")}
                        </p>{/if}
                {/if}
                <button
                    class="text-button danger"
                    onclick={() =>
                        action(`/blocks/${peer(current!)}`, data.blocked.includes(peer(current!)) ? "DELETE" : "POST")}
                    >{t(data.blocked.includes(peer(current)) ? "Unblock reader" : "Block reader")}</button
                >
                {#if current.request}<p class="small muted">
                        {t(
                            "Blocking stops direct messages. Active loan conversations remain available for return arrangements."
                        )}
                    </p>{/if}
                {#if messages.length >= 50}<button class="text-button" disabled={loadingOlder} onclick={loadOlder}
                        >{t(loadingOlder ? "Loading…" : "Load older messages")}</button
                    >{/if}
                <div class="message-list" role="log" aria-label={t("Conversation messages")} aria-live="polite">
                    {#each messages as message (message.id)}<article
                            class:mine={message.senderId === data.user?.id}
                            class="message-bubble"
                        >
                            <p class="small muted">
                                {message.senderId === data.user?.id
                                    ? t("You")
                                    : data.profiles[message.senderId]?.displayName || t("Reader")} · {date(
                                    message.createdAt,
                                    { hour: "2-digit", minute: "2-digit" }
                                )}
                            </p>
                            <p class="message-body" dir="auto">{message.body}</p>
                            {#if message.senderId === data.user?.id && readAt && new Date(message.createdAt) <= new Date(readAt)}<span
                                    class="small muted">{t("Read")}</span
                                >{/if}
                        </article>{/each}
                    {#if !messages.length}<p class="muted">{t("Start with a friendly hello.")}</p>{/if}
                </div>
                <p class="small muted typing-indicator" role="status">{typing ? t("Typing…") : ""}</p>
                {#if canSend}<form onsubmit={send} class="stack">
                        <label
                            >{t("Your message")}<textarea
                                dir="auto"
                                bind:value={body}
                                oninput={signalTyping}
                                maxlength="4000"
                                required
                                disabled={sending}></textarea></label
                        ><button class="button" disabled={sending || $connection !== "connected"}
                            >{t(sending ? "Sending…" : "Send message")}</button
                        >
                    </form>
                {:else}<p class="muted">
                        {t(
                            current.status === "pending"
                                ? "Waiting for the message request to be accepted."
                                : "This conversation is read-only."
                        )}
                    </p>{/if}
            {:else}<div class="empty-state">
                    <h2>{t("A conversation starts a connection")}</h2>
                    <p>{t("Choose a conversation or message someone from their public shelf.")}</p>
                    <a href="/shelves">{t("Explore shelves →")}</a>
                </div>{/if}
        </div>
    </div>
</section>
