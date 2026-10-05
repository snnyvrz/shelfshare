import { api } from "$lib/server/api";
import type { CityOption } from "$lib/types";
import type { LayoutServerLoad } from "./$types";

export const load: LayoutServerLoad = async (event) => {
    const cities = await api<{ data: CityOption[] }>(event, "/cities", {}, true).catch(() => ({ data: [] }));
    return { user: event.locals.user, locale: event.locals.locale, cities: cities.data };
};
