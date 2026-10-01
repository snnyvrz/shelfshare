<script lang="ts">
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
    <p class="eyebrow">{copy.availability}</p>
    <h3><a href="/books/{copy.bookId}">{copy.book.Title}</a></h3>
    <p class="muted">by {copy.book.Author.Name}</p>
    <p><a href="/shelves/{copy.ownerId}">{owner?.displayName || "Reader"}'s shelf</a></p>
    <p><strong>Condition:</strong> {copy.condition}</p>
    {#if copy.notes}<p class="message-body">{copy.notes}</p>{/if}
    {#if copy.ownerId === userId}
        <a class="button secondary" href="/my-shelf">Manage my copy</a>
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
                    >Message to the owner<textarea
                        name="message"
                        maxlength="4000"
                        placeholder="Introduce yourself and suggest a time to exchange the book."></textarea></label
                >
                <button class="button" disabled={pending}>{pending ? "Sending…" : "Request to borrow"}</button>
            </form>
        {:else}<a class="button" href="/login?returnTo=%2Fshelves%2F{copy.ownerId}">Log in to borrow</a>{/if}
    {/if}
</article>
