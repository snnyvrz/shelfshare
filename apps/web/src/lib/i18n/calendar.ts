const persian = new Intl.DateTimeFormat("en-US-u-ca-persian", {
    timeZone: "UTC",
    year: "numeric",
    month: "numeric",
    day: "numeric",
});

export function latinDigits(value: string) {
    return value.replace(/[۰-۹٠-٩]/g, (digit) => String(digit.charCodeAt(0) - (digit >= "۰" ? 1776 : 1632)));
}
function parts(value: Date) {
    const values = Object.fromEntries(persian.formatToParts(value).map((part) => [part.type, part.value]));
    return [Number(values.year), Number(values.month), Number(values.day)];
}
export function toJalali(iso: string) {
    if (!/^\d{4}-\d{2}-\d{2}$/.test(iso)) return iso;
    const parsed = new Date(`${iso}T12:00:00Z`);
    if (Number.isNaN(parsed.getTime())) return iso;
    return parts(parsed)
        .map((value, i) => String(value).padStart(i === 0 ? 4 : 2, "0"))
        .join("/");
}
export function fromJalali(input: string): string | null {
    const match = /^(\d{1,4})[/-](\d{1,2})[/-](\d{1,2})$/.exec(latinDigits(input.trim()));
    if (!match) return null;
    const [year, month, day] = match.slice(1).map(Number);
    if (year < 1 || year > 3177 || month < 1 || month > 12 || day < 1 || day > 31) return null;
    const target = year * 10000 + month * 100 + day;
    let low = Math.floor(Date.UTC(year + 620, 0, 1) / 86400000);
    let high = Math.floor(Date.UTC(year + 623, 0, 1) / 86400000);
    while (low <= high) {
        const mid = Math.floor((low + high) / 2);
        const date = new Date(mid * 86400000);
        const [y, m, d] = parts(date);
        const value = y * 10000 + m * 100 + d;
        if (value === target) return date.toISOString().slice(0, 10);
        if (value < target) low = mid + 1;
        else high = mid - 1;
    }
    return null;
}
// Used on the server as well as in enhanced forms, so date fields work without JS.
export function readDate(form: Pick<FormData, "get" | "has">, name: string) {
    if (form.has(`${name}_jalali`)) {
        const value = String(form.get(`${name}_jalali`) ?? "").trim();
        return value ? (fromJalali(value) ?? value) : "";
    }
    return String(form.get(name) ?? "").trim();
}
