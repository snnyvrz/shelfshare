<script lang="ts">
    import { useI18n } from "$lib/i18n/context";
    const { t, city } = useI18n();
    import { enhance } from "$app/forms";
    import type { PageProps } from "./$types";
    let { data, form }: PageProps = $props();
</script>

<svelte:head><title>{t("My profile · ShelfShare")}</title></svelte:head>
<section class="section narrow">
    <h1>{t("My public profile")}</h1>
    <p>{t("Introduce yourself to other readers. Your account email stays private.")}</p>
    {#if form?.message}<p class="alert" role="alert">{t(form.message)}</p>{/if}{#if form?.success}<p role="status">
            {t("Profile saved.")}
        </p>{/if}
    <form method="POST" use:enhance class="panel stack">
        <label
            >{t("Display name")}<input
                dir="auto"
                name="displayName"
                value={data.profile.displayName}
                required
                maxlength="80"
            /></label
        >
        <label>{t("About me")}<textarea dir="auto" name="bio" maxlength="1000">{data.profile.bio}</textarea></label>
        {#if !data.profile.discoveryCityId}<label
                >{t("General location")}<input
                    dir="auto"
                    name="location"
                    value={data.profile.location}
                    maxlength="120"
                    placeholder={t("City or neighborhood")}
                /></label
            >{/if}
        <label
            >{t("City")}
            <select aria-label={t("City")} name="discoveryCityId" value={data.profile.discoveryCityId}>
                <option value="">{t("No selected city")}</option>
                {#each data.cities as option (option.id)}<option value={option.id}>{city(option)}</option>{/each}
            </select>
        </label>
        <label class="check-label"
            ><input type="checkbox" name="discoveryEnabled" checked={data.profile.discoveryEnabled} />
            {t("Show me and my public books in nearby discovery")}</label
        >
        <p class="small muted">
            {t(
                "Your selected city becomes your public location. We use its representative centre, never GPS or your address. Nearby discovery is optional; your public shelf remains accessible when it is off."
            )}
        </p>
        <p class="small muted">
            {t("City data:")} <a href="https://www.geonames.org/">GeoNames</a>, <bdi>CC BY 4.0.</bdi>
        </p>
        <p class="small muted">{t("These fields are public. Use messages to share exchange details privately.")}</p>
        <button class="button">{t("Save profile")}</button><a href="/shelves/{data.profile.id}"
            >{t("View public shelf →")}</a
        >
    </form>
</section>
