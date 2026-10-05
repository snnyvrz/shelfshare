<script lang="ts">
    import { useI18n } from "$lib/i18n/context";
    const { t, location } = useI18n();
    import { enhance } from "$app/forms";
    import CopyCard from "$lib/components/CopyCard.svelte";
    import type { PageProps } from "./$types";
    let { data, form }: PageProps = $props();
</script>

<svelte:head><title>{t("{name}'s shelf", { name: data.profile.displayName })} · ShelfShare</title></svelte:head>
<section class="section">
    <a class="back-link" href="/shelves">{t("← Readers' shelves")}</a>
    <h1>{t("{name}'s shelf", { name: data.profile.displayName })}</h1>
    {#if data.profile.location}<p class="muted" dir="auto">{location(data.profile.location)}</p>{/if}
    <p class="message-body" dir="auto">{data.profile.bio || t("A reader with stories to share.")}</p>
    {#if data.user?.id === data.profile.id}<a class="button secondary" href="/account/profile">{t("Edit profile")}</a>
    {:else if data.user}<form method="POST" action="?/message" use:enhance>
            <button class="button">{t("Message {name}", { name: data.profile.displayName })}</button>
        </form>
    {:else}<a class="button" href="/login?returnTo={encodeURIComponent(`/shelves/${data.profile.id}`)}"
            >{t("Log in to message")}</a
        >{/if}
    {#if form?.message}<p class="alert" role="alert">{t(form.message)}</p>{/if}
    {#if form?.success}<p class="panel" role="status">
            {t(form.success)} <a href="/requests">{t("View requests →")}</a>
        </p>{/if}
    <div class="copy-grid">
        {#each data.copies as copy (copy.id)}<CopyCard {copy} owner={data.profile} userId={data.user?.id} />{/each}
    </div>
    {#if !data.copies.length}<p class="empty-state">{t("No public copies on this shelf yet.")}</p>{/if}
    <div class="actions">
        {#if data.page > 1}<a href="?page={data.page - 1}">{t("← Previous")}</a>{/if}{#if data.page * 50 < data.total}<a
                href="?page={data.page + 1}">{t("Next →")}</a
            >
            >{/if}
    </div>
</section>
