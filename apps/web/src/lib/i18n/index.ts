import { messages, type MessageKey } from "./messages";
import type { CityOption } from "$lib/types";

export type Locale = "fa" | "en";
export const localeCookie = "shelfshare_locale";
export function resolveLocale(value: unknown): Locale {
    return value === "en" ? "en" : "fa";
}
export function direction(locale: Locale) {
    return locale === "fa" ? "rtl" : "ltr";
}
export function translate(locale: Locale, key: string, params: Record<string, string | number> = {}) {
    const text =
        locale === "fa" && Object.hasOwn(messages, key)
            ? messages[key as MessageKey]
            : key === "return_pending"
              ? "return pending"
              : key;
    return text.replace(/\{(\w+)\}/g, (match, name: string) => String(params[name] ?? match));
}
export function number(locale: Locale, value: number) {
    return new Intl.NumberFormat(locale === "fa" ? "fa-IR" : "en-US").format(value);
}
export function date(locale: Locale, value: string | Date, options: Intl.DateTimeFormatOptions = {}) {
    const input = typeof value === "string" && /^\d{4}-\d{2}-\d{2}$/.test(value) ? `${value}T12:00:00Z` : value;
    const parsed = new Date(input);
    if (Number.isNaN(parsed.getTime())) return "";
    return new Intl.DateTimeFormat(locale === "fa" ? "fa-IR-u-ca-persian" : "en-US-u-ca-gregory", {
        timeZone: "UTC",
        year: "numeric",
        month: "short",
        day: "numeric",
        ...options,
    }).format(parsed);
}
export function cityLabel(locale: Locale, city: CityOption) {
    return city.names?.[locale] ?? city.label;
}
export function locationLabel(locale: Locale, location: string, cities: CityOption[]) {
    const city = cities.find((city) => city.label === location || Object.values(city.names ?? {}).includes(location));
    return city ? cityLabel(locale, city) : location;
}
