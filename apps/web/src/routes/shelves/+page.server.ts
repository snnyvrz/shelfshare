import { copyList, borrow } from "$lib/server/community";
import { pageError, requireUser } from "$lib/server/api";
import type { Actions, PageServerLoad } from "./$types";
import { api } from "$lib/server/api";
import type { Profile, OwnProfile, CityOption } from "$lib/types";
import { latinDigits } from "$lib/i18n/calendar";
export const load: PageServerLoad = async (event) => {
    event.depends("shelfshare:community");
    try {
        const q = event.url.searchParams.get("q") || "";
        const mode = event.url.searchParams.get("mode") || "all";
        if (mode !== "all") requireUser(event);
        const [cities, own] = await Promise.all([
            api<{ data: CityOption[] }>(event, "/cities", {}, true),
            event.locals.user ? api<OwnProfile>(event, "/me", {}, true) : Promise.resolve(null),
        ]);
        const cityId = event.url.searchParams.get("cityId") ?? own?.discoveryCityId ?? "";
        const radiusKm = latinDigits(event.url.searchParams.get("radiusKm") || "25").replace(/٫/g, ".");
        const filters = new URLSearchParams({ mode, cityId, radiusKm });
        const page = Number(event.url.searchParams.get("page")) || 1;
        const readersPage = Number(event.url.searchParams.get("readersPage")) || 1;
        const readers = await api<{ data: Profile[]; total: number }>(
            event,
            `${mode === "all" ? "/profiles?" : `/nearby/profiles?${filters}&`}q=${encodeURIComponent(q)}&page=${readersPage}`,
            {},
            true
        );
        return {
            ...(await copyList(
                event,
                mode === "all" ? `/copies?page=${page}` : `/nearby/copies?${filters}&page=${page}`
            )),
            page,
            cities: cities.data,
            mode,
            cityId,
            radiusKm,
            paginationQuery: new URLSearchParams({ mode, cityId, radiusKm, q }).toString(),
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
