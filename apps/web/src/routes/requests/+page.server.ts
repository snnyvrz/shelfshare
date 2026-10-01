import { fail } from "@sveltejs/kit";
import { api, ApiError, pageError, requireUser } from "$lib/server/api";
import { profiles } from "$lib/server/community";
import type { BorrowRequest, Conversation } from "$lib/types";
import type { Actions, PageServerLoad } from "./$types";
export const load: PageServerLoad = async (event) => {
    event.depends("shelfshare:community");
    requireUser(event);
    try {
        const page = Number(event.url.searchParams.get("page")) || 1;
        const result = await api<{ data: BorrowRequest[]; total: number }>(event, `/requests?page=${page}`);
        const cs = await api<{ data: Conversation[] }>(event, "/conversations");
        return {
            requests: result.data,
            total: result.total,
            page,
            conversations: cs.data,
            profiles: await profiles(
                event,
                result.data.flatMap((r) => [r.ownerId, r.borrowerId])
            ),
        };
    } catch (cause) {
        pageError(cause);
    }
};
export const actions: Actions = {
    transition: async (event) => {
        requireUser(event);
        const form = await event.request.formData();
        const due = String(form.get("dueAt") || "");
        try {
            await api(
                event,
                `/requests/${encodeURIComponent(String(form.get("id")))}/${encodeURIComponent(String(form.get("action")))}`,
                {
                    method: "POST",
                    body: JSON.stringify(due ? { dueAt: new Date(`${due}T23:59:59Z`).toISOString() } : {}),
                }
            );
            return { success: "Request updated." };
        } catch (cause) {
            if (!(cause instanceof ApiError)) throw cause;
            return fail(cause.status, { message: cause.message });
        }
    },
};
