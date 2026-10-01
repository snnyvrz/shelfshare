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
        <label
            >General location<input
                name="location"
                value={data.profile.location}
                maxlength="120"
                placeholder="City or neighborhood"
            /></label
        >
        <p class="small muted">These fields are public. Use messages to share exchange details privately.</p>
        <button class="button">Save profile</button><a href="/shelves/{data.profile.id}">View public shelf →</a>
    </form>
</section>
