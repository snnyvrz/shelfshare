import assert from "node:assert/strict";
import { it } from "node:test";
import { readdirSync, readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { messages } from "../src/lib/i18n/messages";
import { cityLabel, date, direction, locationLabel, number, resolveLocale, translate } from "../src/lib/i18n/index";
import { fromJalali, latinDigits, readDate, toJalali } from "../src/lib/i18n/calendar";

it("defaults to Persian and isolates English/Persian formatting", async () => {
    for (const value of [undefined, "", "fr", "FA", "<script>"]) assert.equal(resolveLocale(value), "fa");
    assert.equal(resolveLocale("en"), "en");
    assert.equal(direction("fa"), "rtl");
    assert.equal(direction("en"), "ltr");
    const [fa, en] = await Promise.all([
        Promise.resolve(translate("fa", "My shelf")),
        Promise.resolve(translate("en", "My shelf")),
    ]);
    assert.equal(fa, "قفسهٔ من");
    assert.equal(en, "My shelf");
    assert.equal(
        translate("fa", "Page {page} of {total}", { page: number("fa", 2), total: number("fa", 12) }),
        "صفحهٔ ۲ از ۱۲"
    );
    assert.equal(number("en", 12), "12");
    assert.match(date("fa", "2026-10-05"), /۱۴۰۵/);
    assert.match(date("en", "2026-10-05"), /2026/);
    assert.equal(date("fa", "not-a-date"), "");
    assert.equal(translate("fa", "constructor"), "constructor");
});

it("covers literal interface translation keys in the Persian catalog", () => {
    const root = fileURLToPath(new URL("../src/", import.meta.url));
    for (const path of readdirSync(root, { recursive: true })) {
        if (typeof path !== "string" || !path.endsWith(".svelte")) continue;
        const source = readFileSync(`${root}/${path}`, "utf8");
        for (const match of source.matchAll(/\bt\("([^"\n]+)"/g))
            assert.ok(Object.hasOwn(messages, match[1]), `${path}: missing Persian translation for ${match[1]}`);
    }
});

it("localizes authoritative city labels without translating free-text locations", () => {
    const city = {
        id: "112931",
        label: "تهران / Tehran, ایران / Iran",
        names: { en: "Tehran, Iran", fa: "تهران، ایران" },
    };
    assert.equal(cityLabel("fa", city), "تهران، ایران");
    assert.equal(cityLabel("en", city), "Tehran, Iran");
    assert.equal(locationLabel("en", city.label, [city]), "Tehran, Iran");
    assert.equal(locationLabel("fa", "Near Tehran University", [city]), "Near Tehran University");
});

it("converts Jalali dates including leap days and rejects impossible dates", () => {
    assert.equal(fromJalali("۱۴۰۵/۰۷/۱۳"), "2026-10-05");
    assert.equal(fromJalali("١٤٠٥/٠٧/١٣"), "2026-10-05");
    assert.equal(fromJalali("1399/12/30"), "2021-03-20");
    assert.equal(fromJalali("1400/12/30"), null);
    assert.equal(fromJalali("1405/07/31"), null);
    assert.equal(fromJalali("1405/13/01"), null);
    assert.equal(fromJalali("invalid"), null);
    assert.equal(toJalali("2026-10-05"), "1405/07/13");
    // Cross Nowruz, leap-year, and year-boundary transitions.
    for (const iso of ["2020-03-20", "2021-03-20", "2021-03-21", "2024-02-29", "2026-12-31", "2027-01-01"])
        assert.equal(fromJalali(toJalali(iso)), iso);
    assert.equal(latinDigits("۱۲.٥"), "12.5");
});

it("normalizes both enhanced and non-JavaScript date submissions", () => {
    const form = new FormData();
    form.set("dueAt", "2026-10-05");
    assert.equal(readDate(form, "dueAt"), "2026-10-05");
    form.set("dueAt_jalali", "۱۴۰۶/۰۱/۰۱");
    assert.equal(readDate(form, "dueAt"), "2027-03-21");
    form.set("dueAt_jalali", "1400/12/30");
    assert.equal(readDate(form, "dueAt"), "1400/12/30");
    form.set("dueAt_jalali", "");
    assert.equal(readDate(form, "dueAt"), "");
    assert.equal(
        readDate(new URLSearchParams("published_after_jalali=1405%2F07%2F13"), "published_after"),
        "2026-10-05"
    );
});
