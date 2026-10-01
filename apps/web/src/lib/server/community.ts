import { fail } from "@sveltejs/kit";
import { ApiError, api, requireUser } from "./api";
import type { RequestEvent } from "@sveltejs/kit";
import type { PhysicalCopy, Profile } from "$lib/types";

export async function profiles(event: RequestEvent, ids: string[]) {
    const result: Record<string, Profile> = {};
    await Promise.all(
        [...new Set(ids)].map(async (id) => {
            try {
                result[id] = await api<Profile>(event, `/profiles/${encodeURIComponent(id)}`, {}, true);
            } catch (cause) {
                if (!(cause instanceof ApiError) || cause.status !== 404) throw cause;
            }
        })
    );
    return result;
}

export async function borrow(event: RequestEvent) {
    requireUser(event);
    const form = await event.request.formData();
    try {
        await api(event, `/copies/${encodeURIComponent(String(form.get("copyId")))}/requests`, {
            method: "POST",
            body: JSON.stringify({ message: String(form.get("message") || "") }),
        });
        return { success: "Request sent. Open Requests to follow its progress and arrange the exchange." };
    } catch (cause) {
        if (!(cause instanceof ApiError)) throw cause;
        return fail(cause.status, { message: cause.message });
    }
}

export async function copyList(event: RequestEvent, path: string) {
    const result = await api<{ data: PhysicalCopy[]; total: number }>(event, path);
    return {
        copies: result.data,
        total: result.total,
        profiles: await profiles(
            event,
            result.data.map((c) => c.ownerId)
        ),
    };
}
