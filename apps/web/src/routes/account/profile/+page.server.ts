import { fail } from "@sveltejs/kit";
import { api, ApiError, pageError, requireUser } from "$lib/server/api";
import type { Profile } from "$lib/types";
import type { Actions, PageServerLoad } from "./$types";
export const load: PageServerLoad = async (event) => {
    requireUser(event);
    try {
        return { profile: await api<Profile>(event, "/me", {}, true) };
    } catch (cause) {
        pageError(cause);
    }
};
export const actions: Actions = {
    default: async (event) => {
        requireUser(event);
        const form = await event.request.formData();
        try {
            await api(
                event,
                "/me",
                {
                    method: "PATCH",
                    body: JSON.stringify(
                        Object.fromEntries(
                            ["displayName", "bio", "location"].map((key) => [key, String(form.get(key) || "")])
                        )
                    ),
                },
                true
            );
            return { success: true };
        } catch (cause) {
            if (!(cause instanceof ApiError)) throw cause;
            return fail(cause.status, { message: cause.message });
        }
    },
};
