<script lang="ts">
    import { enhance } from "$app/forms";
    import type { BorrowRequest } from "$lib/types";
    import type { PageProps } from "./$types";
    let { data, form }: PageProps = $props();
    function actions(r: BorrowRequest) {
        const owner = r.ownerId === data.user?.id;
        if (r.status === "pending")
            return owner
                ? [
                      ["accept", "Approve"],
                      ["decline", "Decline"],
                  ]
                : [["cancel", "Cancel request"]];
        if (r.status === "accepted")
            return owner
                ? [
                      ["handover", "Confirm handover"],
                      ["cancel", "Cancel reservation"],
                  ]
                : [["cancel", "Cancel reservation"]];
        if (r.status === "borrowed" && !owner) return [["return", "I've returned the book"]];
        if (r.status === "return_pending" && owner) return [["confirm-return", "Confirm receipt"]];
        return [];
    }
</script>

<svelte:head><title>Requests and loans · ShelfShare</title></svelte:head>
<section class="section">
    <h1>Requests and loans</h1>
    <p>Approve requests, arrange exchanges in messages, and confirm every book's return.</p>
    {#if form?.message}<p class="alert" role="alert">{form.message}</p>{/if}{#if form?.success}<p role="status">
            {form.success}
        </p>{/if}
    <div class="copy-grid">
        {#each data.requests as request (request.id)}
            {@const peer = request.ownerId === data.user?.id ? request.borrowerId : request.ownerId}
            {@const conversation = data.conversations.find((c) => c.requestId === request.id)}
            <article class="panel stack">
                <p class="eyebrow">
                    {request.ownerId === data.user?.id ? "Lending" : "Borrowing"} · {request.status.replaceAll(
                        "_",
                        " "
                    )}
                </p>
                <h2><a href="/books/{request.copy.bookId}">{request.copy.book.Title}</a></h2>
                <p>With <a href="/shelves/{peer}">{data.profiles[peer]?.displayName || "Reader"}</a></p>
                {#if request.dueAt}<p>
                        Due {new Date(request.dueAt).toLocaleDateString()}{new Date(request.dueAt) < new Date() &&
                        ["borrowed", "return_pending"].includes(request.status)
                            ? " · overdue"
                            : ""}
                    </p>{/if}
                <p class="message-body">{request.message}</p>
                <a href={conversation ? `/messages?conversation=${conversation.id}` : `/messages?request=${request.id}`}
                    >Arrange exchange in messages →</a
                >
                {#each actions(request) as [action, label] (action)}<form
                        method="POST"
                        action="?/transition"
                        use:enhance
                        class="stack"
                    >
                        <input type="hidden" name="id" value={request.id} /><input
                            type="hidden"
                            name="action"
                            value={action}
                        />
                        {#if action === "accept" || action === "handover"}<label
                                >Agreed due date (optional)<input type="date" name="dueAt" /></label
                            >{/if}
                        <button class="button secondary">{label}</button>
                    </form>{/each}
            </article>{/each}
    </div>
    {#if !data.requests.length}<div class="empty-state">
            <h2>No requests yet</h2>
            <a href="/shelves">Find a book to borrow →</a>
        </div>{/if}
    <div class="actions">
        {#if data.page > 1}<a href="?page={data.page - 1}">← Previous</a>{/if}{#if data.page * 50 < data.total}<a
                href="?page={data.page + 1}">Next →</a
            >{/if}
    </div>
</section>
