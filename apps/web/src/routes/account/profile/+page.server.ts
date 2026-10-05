import { fail } from "@sveltejs/kit";
import { api, ApiError, pageError, requireUser } from "$lib/server/api";
import type { OwnProfile, CityOption } from "$lib/types";
import type { Actions, PageServerLoad } from "./$types";
export const load: PageServerLoad = async (event) => {
    requireUser(event);
    try {
        const [profile, cities] = await Promise.all([
            api<OwnProfile>(event, "/me", {}, true),
            api<{ data: CityOption[] }>(event, "/cities", {}, true),
        ]);
        return { profile, cities: cities.data };
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
                    body: JSON.stringify({
                        ...Object.fromEntries(["displayName", "bio"].map((key) => [key, String(form.get(key) || "")])),
                        ...(form.get("location") !== null ? { location: String(form.get("location")) } : {}),
                        discoveryCityId: String(form.get("discoveryCityId") || ""),
                        discoveryEnabled: form.get("discoveryEnabled") === "on",
                    }),
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
