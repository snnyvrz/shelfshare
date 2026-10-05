import { getContext, setContext } from "svelte";
import { cityLabel, locationLabel, date, number, translate, type Locale } from "./index";
import type { CityOption } from "$lib/types";

const key = Symbol("locale");
const citiesKey = Symbol("cities");
export function provideLocale(getLocale: () => Locale, getCities: () => CityOption[]) {
    setContext(key, getLocale);
    setContext(citiesKey, getCities);
}
export function useI18n() {
    const getLocale = getContext<() => Locale>(key) ?? (() => "fa" as const);
    const getCities = getContext<() => CityOption[]>(citiesKey) ?? (() => []);
    return {
        get locale() {
            return getLocale();
        },
        t: (key: string, params?: Record<string, string | number>) => translate(getLocale(), key, params),
        n: (value: number) => number(getLocale(), value),
        date: (value: string | Date, options?: Intl.DateTimeFormatOptions) => date(getLocale(), value, options),
        city: (value: Parameters<typeof cityLabel>[1]) => cityLabel(getLocale(), value),
        location: (value: string) => locationLabel(getLocale(), value, getCities()),
    };
}
