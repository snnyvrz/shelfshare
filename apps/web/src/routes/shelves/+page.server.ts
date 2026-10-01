import { copyList, borrow } from "$lib/server/community";
import { pageError } from "$lib/server/api";
import type { Actions, PageServerLoad } from "./$types";
import { api } from "$lib/server/api";
import type { Profile } from "$lib/types";
export const load: PageServerLoad = async (event) => {
    event.depends("shelfshare:community");
    try {
        const q = event.url.searchParams.get("q") || "";
        const readersPage = Number(event.url.searchParams.get("readersPage")) || 1;
        const readers = await api<{ data: Profile[]; total: number }>(
            event,
            `/profiles?q=${encodeURIComponent(q)}&page=${readersPage}`,
            {},
            true
        );
        return {
            ...(await copyList(event, `/copies?page=${Number(event.url.searchParams.get("page")) || 1}`)),
            page: Number(event.url.searchParams.get("page")) || 1,
            readers: readers.data,
            readerTotal: readers.total,
            q,
            readersPage,
        };
    } catch (cause) {
        pageError(cause);
    }
};
export const actions: Actions = { borrow };
