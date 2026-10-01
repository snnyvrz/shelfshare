import { api, pageError, requireUser } from "$lib/server/api";
import { profiles } from "$lib/server/community";
import type { Conversation, Message } from "$lib/types";
import type { PageServerLoad } from "./$types";
export const load: PageServerLoad = async (event) => {
    event.depends("shelfshare:community");
    requireUser(event);
    try {
        const page = Number(event.url.searchParams.get("page")) || 1;
        const result = await api<{ data: Conversation[]; total: number }>(event, `/conversations?page=${page}`);
        const selected = event.url.searchParams.get("conversation");
        const request = event.url.searchParams.get("request");
        // A selected thread can live beyond the current inbox page.
        let conversation = result.data.find((c) => c.id === selected || (request && c.requestId === request));
        if (!conversation && (selected || request)) {
            const detail = await api<Conversation>(
                event,
                selected
                    ? `/conversations/${encodeURIComponent(selected)}`
                    : `/request-conversations/${encodeURIComponent(request!)}`
            );
            conversation = detail;
        }
        const messages = conversation
            ? (await api<{ data: Message[] }>(event, `/conversations/${conversation.id}/messages`)).data.reverse()
            : [];
        const blocks = await api<{ data: { userId: string }[] }>(event, "/blocks");
        const all = conversation ? [...result.data, conversation] : result.data;
        return {
            conversations: result.data,
            total: result.total,
            page,
            conversation: conversation || null,
            messages,
            blocked: blocks.data.map((b) => b.userId),
            profiles: await profiles(
                event,
                all.flatMap((c) => [c.userA, c.userB])
            ),
        };
    } catch (cause) {
        pageError(cause);
    }
};
