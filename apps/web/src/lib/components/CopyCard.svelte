<script lang="ts">
    import { useI18n } from "$lib/i18n/context";
    const { t } = useI18n();
    import { enhance } from "$app/forms";
    import type { PhysicalCopy, Profile } from "$lib/types";
    let {
        copy,
        owner,
        userId,
        action = "?/borrow",
    }: { copy: PhysicalCopy; owner?: Profile; userId?: string; action?: string } = $props();
    let pending = $state(false);
</script>

<article class="panel copy-card">
    <p class="eyebrow">{t(copy.availability)}</p>
    <h3><a dir="auto" href="/books/{copy.bookId}">{copy.book.Title}</a></h3>
    <p class="muted">{t("by {name}", { name: copy.book.Author.Name })}</p>
    <p><a href="/shelves/{copy.ownerId}">{t("{name}'s shelf", { name: owner?.displayName || t("Reader") })}</a></p>
    <p><strong>{t("Condition:")}</strong> <bdi>{copy.condition}</bdi></p>
    {#if copy.notes}<p class="message-body" dir="auto">{copy.notes}</p>{/if}
    {#if copy.ownerId === userId}
        <a class="button secondary" href="/my-shelf">{t("Manage my copy")}</a>
    {:else if copy.availability === "available"}
        {#if userId}
            <form
                method="POST"
                {action}
                use:enhance={() => {
                    pending = true;
                    return async ({ update }) => {
                        await update();
                        pending = false;
                    };
                }}
            >
                <input type="hidden" name="copyId" value={copy.id} />
                <label
                    >{t("Message to the owner")}<textarea
                        dir="auto"
                        name="message"
                        maxlength="4000"
                        placeholder={t("Introduce yourself and suggest a time to exchange the book.")}
                    ></textarea></label
                >
                <button class="button" disabled={pending}>{t(pending ? "Sending…" : "Request to borrow")}</button>
            </form>
        {:else}<a class="button" href="/login?returnTo=%2Fshelves%2F{copy.ownerId}">{t("Log in to borrow")}</a>{/if}
    {/if}
</article>
