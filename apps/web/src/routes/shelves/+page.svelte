<script lang="ts">
    import { useI18n } from "$lib/i18n/context";
    const i18n = useI18n();
    const { t, city, location } = i18n;
    import CopyCard from "$lib/components/CopyCard.svelte";
    import type { PageProps } from "./$types";
    let { data, form }: PageProps = $props();
</script>

<svelte:head><title>{t("Readers' shelves · ShelfShare")}</title></svelte:head>
<section class="section">
    <p class="eyebrow">{t("REAL BOOKS. REAL CONNECTIONS.")}</p>
    <h1>{t("Readers' shelves")}</h1>
    <p>
        {t(
            "Discover physical books shared by other readers. Request a copy, arrange your exchange, and return it for the next reader."
        )}
    </p>
    <div class="actions">
        <a class="button" href="/my-shelf">{t("Add my books")}</a><a href="/books">{t("Browse catalog titles →")}</a>
    </div>
    <h2 class="section">{t("Find a reader")}</h2>
    <form method="GET" class="actions">
        {#if data.user}
            <label
                >{t("Discovery")}
                <select aria-label={t("Discovery")} name="mode" value={data.mode}>
                    <option value="all">{t("All readers and books")}</option>
                    <option value="city">{t("Same city")}</option>
                    <option value="radius">{t("Within distance")}</option>
                </select>
            </label>
            <label
                >{t("Search from city")}
                <select aria-label={t("Search from city")} name="cityId" value={data.cityId}>
                    <option value="">{t("Select a city")}</option>
                    {#each data.cities as option (option.id)}<option value={option.id}>{city(option)}</option>{/each}
                </select>
            </label>
            <label
                >{t("Radius (km)")}
                <input
                    type={i18n.locale === "fa" ? "text" : "number"}
                    inputmode="decimal"
                    name="radiusKm"
                    value={i18n.locale === "fa"
                        ? data.radiusKm.replace(/\d/g, (digit) => "۰۱۲۳۴۵۶۷۸۹"[Number(digit)])
                        : data.radiusKm}
                    min="1"
                    max="500"
                    step="any"
                    list="radii"
                />
                <datalist id="radii"
                    ><option value="10"></option><option value="25"></option><option value="50"></option><option
                        value="100"
                    ></option></datalist
                >
            </label>
        {:else}<p>
                <a href="/login?returnTo=%2Fshelves">{t("Sign in")}</a>
                {t("to discover nearby readers and books.")}
            </p>{/if}
        <label
            >{t("Display name or general location")}<input
                dir="auto"
                name="q"
                value={data.q}
                maxlength="100"
                placeholder={t("Find someone nearby")}
            /></label
        ><button class="button secondary">{t("Search shelves")}</button>
    </form>
    {#if data.mode !== "all"}<p class="small muted">
            {t(
                "Distances are approximate, measured between city centres. Books use their owner's selected city. Results are ordered by nearest city; the text search filters readers only."
            )}
        </p>{/if}
    <p class="small muted">{t("City data:")} <a href="https://www.geonames.org/">GeoNames</a>, <bdi>CC BY 4.0.</bdi></p>
    <div class="copy-grid">
        {#each data.readers as reader (reader.id)}<article class="panel">
                <h3><a dir="auto" href="/shelves/{reader.id}">{reader.displayName}</a></h3>
                <p class="small muted" dir="auto">{location(reader.location)}</p>
                <p class="clamp" dir="auto">{reader.bio || t("A reader with stories to share.")}</p>
                <a href="/shelves/{reader.id}">{t("View shelf and message →")}</a>
            </article>{/each}
    </div>
    {#if !data.readers.length}<p>{t("No readers match this search.")}</p>{/if}
    <div class="actions">
        {#if data.readersPage > 1}<a href="?{data.paginationQuery}&page={data.page}&readersPage={data.readersPage - 1}"
                >{t("← Previous readers")}</a
            >{/if}{#if data.readersPage * 50 < data.readerTotal}<a
                href="?{data.paginationQuery}&page={data.page}&readersPage={data.readersPage + 1}"
                >{t("Next readers →")}</a
            >{/if}
    </div>
    <h2 class="section">{t("Physical books to discover")}</h2>
    {#if form?.message}<p class="alert" role="alert">{t(form.message)}</p>{/if}
    {#if form?.success}<p class="panel" role="status">
            {t(form.success)} <a href="/requests">{t("View requests →")}</a>
        </p>{/if}
    <div class="copy-grid">
        {#each data.copies as copy (copy.id)}<CopyCard
                {copy}
                owner={data.profiles[copy.ownerId]}
                userId={data.user?.id}
            />{/each}
    </div>
    {#if !data.copies.length}<div class="empty-state">
            <h2>{t(data.mode === "all" ? "Start a shelf worth sharing" : "No books match this area")}</h2>
            <p>
                {t(
                    data.mode === "all"
                        ? "Add your physical books and invite another reader into their next chapter."
                        : "Try another city or a larger radius. Only opted-in owners' public books appear here."
                )}
            </p>
        </div>{/if}
    <div class="actions">
        {#if data.page > 1}<a href="?{data.paginationQuery}&readersPage={data.readersPage}&page={data.page - 1}"
                >{t("← Previous")}</a
            >{/if}{#if data.page * 50 < data.total}<a
                href="?{data.paginationQuery}&readersPage={data.readersPage}&page={data.page + 1}">{t("Next →")}</a
            >
            >{/if}
    </div>
</section>
