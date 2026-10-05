import { BadRequestException } from "@nestjs/common";

// GeoNames IR.zip, retrieved 2026-10-05. City records (PPLC/PPLA), not administrative areas.
// CC BY 4.0: https://www.geonames.org/ — see DATA.md for provenance and update instructions.
export const cities = [
    { id: "112931", name: "Tehran", nameFa: "تهران", latitude: 35.69439, longitude: 51.42151, aliases: ["Teheran"] },
    { id: "124665", name: "Mashhad", nameFa: "مشهد", latitude: 36.29807, longitude: 59.60567, aliases: ["Mashad"] },
    { id: "418863", name: "Isfahan", nameFa: "اصفهان", latitude: 32.65246, longitude: 51.67462, aliases: ["Esfahan"] },
    { id: "128747", name: "Karaj", nameFa: "کرج", latitude: 35.83266, longitude: 50.99155, aliases: [] },
    { id: "115019", name: "Shiraz", nameFa: "شیراز", latitude: 29.61031, longitude: 52.53113, aliases: [] },
    { id: "113646", name: "Tabriz", nameFa: "تبریز", latitude: 38.08, longitude: 46.2919, aliases: [] },
    { id: "119208", name: "Qom", nameFa: "قم", latitude: 34.6401, longitude: 50.8764, aliases: ["Ghom", "Qum"] },
    { id: "144448", name: "Ahvaz", nameFa: "اهواز", latitude: 31.31901, longitude: 48.6842, aliases: ["Ahwaz"] },
    { id: "128226", name: "Kermanshah", nameFa: "کرمانشاه", latitude: 34.31417, longitude: 47.065, aliases: [] },
    {
        id: "121801",
        name: "Urmia",
        nameFa: "ارومیه",
        latitude: 37.55274,
        longitude: 45.07605,
        aliases: ["Orumiyeh", "Urumiyeh"],
    },
    { id: "118743", name: "Rasht", nameFa: "رشت", latitude: 37.27611, longitude: 49.58862, aliases: [] },
    { id: "1159301", name: "Zahedan", nameFa: "زاهدان", latitude: 29.4963, longitude: 60.8629, aliases: [] },
];

export function cityLabel(city: (typeof cities)[number]) {
    return `${city.nameFa} / ${city.name}, ایران / Iran`;
}

export function cityById(id: unknown) {
    const city = cities.find((city) => city.id === id);
    if (!city) throw new BadRequestException("Select a supported city");
    return city;
}

function normalize(value: string) {
    return value
        .normalize("NFKD")
        .replace(/[\u0300-\u036f\u064b-\u065f\u200c]/g, "")
        .replace(/ي/g, "ی")
        .replace(/ك/g, "ک")
        .toLowerCase()
        .trim();
}

export function cityOptions(query = "") {
    const q = normalize(query.slice(0, 100));
    return cities
        .filter((city) => [city.name, city.nameFa, ...city.aliases].some((name) => normalize(name).includes(q)))
        .map((city) => ({
            id: city.id,
            label: cityLabel(city),
            names: { en: `${city.name}, Iran`, fa: `${city.nameFa}، ایران` },
        }));
}

export function distanceKm(a: (typeof cities)[number], b: (typeof cities)[number]) {
    const rad = (n: number) => (n * Math.PI) / 180;
    const h =
        Math.sin(rad(b.latitude - a.latitude) / 2) ** 2 +
        Math.cos(rad(a.latitude)) * Math.cos(rad(b.latitude)) * Math.sin(rad(b.longitude - a.longitude) / 2) ** 2;
    return 6371.0088 * 2 * Math.asin(Math.sqrt(Math.min(1, h)));
}

export function discoveryCities(params: Record<string, unknown>) {
    if (params.mode !== "city" && params.mode !== "radius") throw new BadRequestException("Invalid discovery mode");
    const origin = cityById(params.cityId);
    const radius = Number(params.radiusKm);
    if (
        params.mode === "radius" &&
        (typeof params.radiusKm !== "string" || !Number.isFinite(radius) || radius < 1 || radius > 500)
    )
        throw new BadRequestException("Radius must be between 1 and 500 km");
    return cities
        .map((city) => ({ id: city.id, distance: distanceKm(origin, city) }))
        .filter((city) => (params.mode === "city" ? city.id === origin.id : city.distance <= radius))
        .sort((a, b) => a.distance - b.distance || a.id.localeCompare(b.id));
}
