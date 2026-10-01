import { fail, redirect } from "@sveltejs/kit";
import { api, ApiError, pageError, requireUser } from "$lib/server/api";
import { copyList, borrow } from "$lib/server/community";
import type { Profile, Conversation } from "$lib/types";
import type { Actions, PageServerLoad } from "./$types";
export const load: PageServerLoad = async (event) => {
    event.depends("shelfshare:community");
    try {
        const profile = await api<Profile>(event, `/profiles/${encodeURIComponent(event.params.id)}`, {}, true);
        return {
            profile,
            ...(await copyList(
                event,
                `/copies?ownerId=${encodeURIComponent(event.params.id)}&page=${Number(event.url.searchParams.get("page")) || 1}`
            )),
            page: Number(event.url.searchParams.get("page")) || 1,
        };
    } catch (cause) {
        pageError(cause);
    }
};
export const actions: Actions = {
    borrow,
    message: async (event) => {
        requireUser(event);
        let conversation: Conversation;
        try {
            await api(event, `/profiles/${encodeURIComponent(event.params.id)}`, {}, true);
            conversation = await api<Conversation>(event, "/conversations", {
                method: "POST",
                body: JSON.stringify({ userId: event.params.id }),
            });
        } catch (cause) {
            if (!(cause instanceof ApiError)) throw cause;
            return fail(cause.status, { message: cause.message });
        }
        redirect(303, `/messages?conversation=${conversation.id}`);
    },
};
