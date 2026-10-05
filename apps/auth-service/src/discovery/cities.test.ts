import assert from "node:assert/strict";
import { it } from "node:test";
import { BadRequestException } from "@nestjs/common";
import { cities, cityById, cityOptions, discoveryCities, distanceKm } from "./cities";

it("finds bilingual names, transliterations and Persian/Arabic variants without exposing coordinates", () => {
    for (const q of ["Karaj", "کرج", "كرج"]) assert.equal(cityOptions(q)[0].id, "128747");
    assert.equal(cityOptions("Esfahan")[0].id, "418863");
    assert.equal(cityOptions("Orumiyeh")[0].id, "121801");
    assert.equal(cityOptions().length, 12);
    assert.deepEqual(Object.keys(cityOptions()[0]), ["id", "label"]);
    assert.equal(new Set(cities.map((city) => city.id)).size, 12);
});

it("calculates city-centre radius boundaries and keeps same-city matching independent of radius", () => {
    const tehran = cityById("112931"),
        karaj = cityById("128747");
    const distance = distanceKm(tehran, karaj);
    assert.ok(distance > 40 && distance < 45);
    assert.equal(distanceKm(tehran, tehran), 0);
    assert.deepEqual(discoveryCities({ mode: "city", cityId: tehran.id }), [{ id: tehran.id, distance: 0 }]);
    const ids = (radiusKm: number) =>
        discoveryCities({ mode: "radius", cityId: tehran.id, radiusKm: String(radiusKm) }).map((c) => c.id);
    assert.equal(ids(distance - 0.001).includes(karaj.id), false);
    assert.equal(ids(distance + 0.001).includes(karaj.id), true);
    assert.equal(ids(500)[0], tehran.id);
    for (const params of [
        { mode: "radius", cityId: tehran.id, radiusKm: "NaN" },
        { mode: "radius", cityId: tehran.id, radiusKm: "501" },
        { mode: "radius", cityId: tehran.id, radiusKm: "0" },
        { mode: "radius", cityId: tehran.id, radiusKm: ["25"] },
        { mode: "radius", cityId: "invalid", radiusKm: "25" },
        { mode: "invalid", cityId: tehran.id },
    ])
        assert.throws(() => discoveryCities(params), BadRequestException);
});
