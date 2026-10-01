import { api, pageError } from "$lib/server/api";
import type { Author } from "$lib/types";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async (event) => {
    try {
        const authors = await api<{ data: Author }[]>(event, "/authors");
        const q = event.url.searchParams.get("q") ?? "";
        return {
            authors: authors
                .map((a) => a.data)
                .filter((a) => `${a.name} ${a.bio}`.toLowerCase().includes(q.toLowerCase())),
            q,
        };
    } catch (cause) {
        pageError(cause);
    }
};
