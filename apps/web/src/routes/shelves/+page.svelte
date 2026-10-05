<script lang="ts">
    import CopyCard from "$lib/components/CopyCard.svelte";
    import type { PageProps } from "./$types";
    let { data, form }: PageProps = $props();
</script>

<svelte:head><title>Readers' shelves · ShelfShare</title></svelte:head>
<section class="section">
    <p class="eyebrow">REAL BOOKS. REAL CONNECTIONS.</p>
    <h1>Readers' shelves</h1>
    <p>
        Discover physical books shared by other readers. Request a copy, arrange your exchange, and return it for the
        next reader.
    </p>
    <div class="actions">
        <a class="button" href="/my-shelf">Add my books</a><a href="/books">Browse catalog titles →</a>
    </div>
    <h2 class="section">Find a reader</h2>
    <form method="GET" class="actions">
        {#if data.user}
            <label
                >Discovery
                <select aria-label="Discovery" name="mode" value={data.mode}>
                    <option value="all">All readers and books</option>
                    <option value="city">Same city</option>
                    <option value="radius">Within distance</option>
                </select>
            </label>
            <label
                >Search from city
                <select aria-label="Search from city" name="cityId" value={data.cityId}>
                    <option value="">Select a city</option>
                    {#each data.cities as city (city.id)}<option value={city.id}>{city.label}</option>{/each}
                </select>
            </label>
            <label
                >Radius (km)
                <input type="number" name="radiusKm" value={data.radiusKm} min="1" max="500" step="any" list="radii" />
                <datalist id="radii"
                    ><option value="10"></option><option value="25"></option><option value="50"></option><option
                        value="100"
                    ></option></datalist
                >
            </label>
        {:else}<p><a href="/login?returnTo=%2Fshelves">Sign in</a> to discover nearby readers and books.</p>{/if}
        <label
            >Display name or general location<input
                name="q"
                value={data.q}
                maxlength="100"
                placeholder="Find someone nearby"
            /></label
        ><button class="button secondary">Search shelves</button>
    </form>
    {#if data.mode !== "all"}<p class="small muted">
            Distances are approximate, measured between city centres. Books use their owner's selected city. Results are
            ordered by nearest city; the text search filters readers only.
        </p>{/if}
    <p class="small muted">City data: <a href="https://www.geonames.org/">GeoNames</a>, CC BY 4.0.</p>
    <div class="copy-grid">
        {#each data.readers as reader (reader.id)}<article class="panel">
                <h3><a href="/shelves/{reader.id}">{reader.displayName}</a></h3>
                <p class="small muted">{reader.location}</p>
                <p class="clamp">{reader.bio || "A reader with stories to share."}</p>
                <a href="/shelves/{reader.id}">View shelf and message →</a>
            </article>{/each}
    </div>
    {#if !data.readers.length}<p>No readers match this search.</p>{/if}
    <div class="actions">
        {#if data.readersPage > 1}<a href="?{data.paginationQuery}&page={data.page}&readersPage={data.readersPage - 1}"
                >← Previous readers</a
            >{/if}{#if data.readersPage * 50 < data.readerTotal}<a
                href="?{data.paginationQuery}&page={data.page}&readersPage={data.readersPage + 1}">Next readers →</a
            >{/if}
    </div>
    <h2 class="section">Physical books to discover</h2>
    {#if form?.message}<p class="alert" role="alert">{form.message}</p>{/if}
    {#if form?.success}<p class="panel" role="status">{form.success} <a href="/requests">View requests →</a></p>{/if}
    <div class="copy-grid">
        {#each data.copies as copy (copy.id)}<CopyCard
                {copy}
                owner={data.profiles[copy.ownerId]}
                userId={data.user?.id}
            />{/each}
    </div>
    {#if !data.copies.length}<div class="empty-state">
            <h2>{data.mode === "all" ? "Start a shelf worth sharing" : "No books match this area"}</h2>
            <p>
                {data.mode === "all"
                    ? "Add your physical books and invite another reader into their next chapter."
                    : "Try another city or a larger radius. Only opted-in owners' public books appear here."}
            </p>
        </div>{/if}
    <div class="actions">
        {#if data.page > 1}<a href="?{data.paginationQuery}&readersPage={data.readersPage}&page={data.page - 1}"
                >← Previous</a
            >{/if}{#if data.page * 50 < data.total}<a
                href="?{data.paginationQuery}&readersPage={data.readersPage}&page={data.page + 1}">Next →</a
            >{/if}
    </div>
</section>
