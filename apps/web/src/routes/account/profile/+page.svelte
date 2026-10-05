<script lang="ts">
    import { enhance } from "$app/forms";
    import type { PageProps } from "./$types";
    let { data, form }: PageProps = $props();
</script>

<svelte:head><title>My profile · ShelfShare</title></svelte:head>
<section class="section narrow">
    <h1>My public profile</h1>
    <p>Introduce yourself to other readers. Your account email stays private.</p>
    {#if form?.message}<p class="alert" role="alert">{form.message}</p>{/if}{#if form?.success}<p role="status">
            Profile saved.
        </p>{/if}
    <form method="POST" use:enhance class="panel stack">
        <label>Display name<input name="displayName" value={data.profile.displayName} required maxlength="80" /></label>
        <label>About me<textarea name="bio" maxlength="1000">{data.profile.bio}</textarea></label>
        {#if !data.profile.discoveryCityId}<label
                >General location<input
                    name="location"
                    value={data.profile.location}
                    maxlength="120"
                    placeholder="City or neighborhood"
                /></label
            >{/if}
        <label
            >City
            <select aria-label="City" name="discoveryCityId" value={data.profile.discoveryCityId}>
                <option value="">No selected city</option>
                {#each data.cities as city (city.id)}<option value={city.id}>{city.label}</option>{/each}
            </select>
        </label>
        <label class="check-label"
            ><input type="checkbox" name="discoveryEnabled" checked={data.profile.discoveryEnabled} />
            Show me and my public books in nearby discovery</label
        >
        <p class="small muted">
            Your selected city becomes your public location. We use its representative centre, never GPS or your
            address. Nearby discovery is optional; your public shelf remains accessible when it is off.
        </p>
        <p class="small muted">City data: <a href="https://www.geonames.org/">GeoNames</a>, CC BY 4.0.</p>
        <p class="small muted">These fields are public. Use messages to share exchange details privately.</p>
        <button class="button">Save profile</button><a href="/shelves/{data.profile.id}">View public shelf →</a>
    </form>
</section>
