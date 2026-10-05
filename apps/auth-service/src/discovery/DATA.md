# Iranian launch cities

`cities.ts` contains the twelve handpicked major cities, extracted from
[GeoNames' Iran gazetteer](https://download.geonames.org/export/dump/IR.zip)
on 2026-10-05. Source: [GeoNames](https://www.geonames.org/), licensed under
[Creative Commons Attribution 4.0](https://creativecommons.org/licenses/by/4.0/).
The extract is derived data; bilingual labels and selected search aliases are curated.

Coordinates represent cities, never individual readers. They are approximate representative
points, not municipal boundaries or travel distances. No external service is used at runtime.

To update: download IR.zip, match the existing `geonameid` values, verify country `IR` and
feature codes `PPLC`/`PPLA`, and review coordinates/names. Do not match by name alone:
the gazetteer includes smaller places with the same names. Preserve IDs when updating labels.
For new cities, choose the actual populated-place record, add Persian/English names and aliases,
and run discovery tests. Record the new retrieval date here and in `cities.ts`.
Dataset schema: https://download.geonames.org/export/dump/readme.txt.
