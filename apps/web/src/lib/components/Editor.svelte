<script lang="ts">
    import { enhance } from "$app/forms";
    import type { Author, Collection, FormValues } from "$lib/types";
    let {
        collection,
        authors,
        values,
        editing,
        form,
    }: {
        collection: Collection;
        authors: Author[];
        values: FormValues;
        editing: boolean;
        form?: { values?: FormValues; errors?: FormValues; message?: string } | null;
    } = $props();
    let pending = $state(false);
    let current = $derived(form?.values ?? values);
    let errors = $derived(form?.errors ?? {});
    let noun = $derived(collection === "books" ? "book" : "author");
</script>

<svelte:head><title>{editing ? "Edit" : "Add"} {noun} · ShelfShare</title></svelte:head>
<div class="narrow">
    <a class="back-link" href="/{collection}">← Back to {collection}</a>
    <p class="eyebrow">GROW THE COLLECTION</p>
    <h1>{editing ? "Edit" : "Add"} {noun === "author" ? "an" : "a"} {noun}</h1>
    <p class="muted">A little detail helps the next reader find their next great read.</p>
    {#if form?.message}<div class="alert" role="alert">{form.message}</div>{/if}
    {#if collection === "books" && !authors.length}
        <div class="empty-state">
            <h2>Start with an author</h2>
            <p>Create an author before adding their book.</p>
            <a class="button" href="/authors/new">Add an author</a>
        </div>
    {:else}
        <form
            method="POST"
            class="panel editor"
            use:enhance={() => {
                pending = true;
                return async ({ update }) => {
                    try {
                        await update();
                    } finally {
                        pending = false;
                    }
                };
            }}
        >
            {#if collection === "books"}
                <label for="title">Book title <span class="required">*</span></label>
                <input
                    id="title"
                    name="title"
                    required
                    value={current.title ?? ""}
                    aria-invalid={!!errors.title}
                    aria-describedby={errors.title ? "title-error" : undefined}
                />
                {#if errors.title}<p id="title-error" class="field-error">{errors.title}</p>{/if}
                <label for="author_id">Author <span class="required">*</span></label>
                <select
                    id="author_id"
                    name="author_id"
                    required
                    value={current.author_id ?? ""}
                    aria-invalid={!!errors.author_id}
                >
                    <option value="" disabled>Choose an author</option>
                    {#each authors as author (author.id)}<option value={author.id}>{author.name}</option>{/each}
                </select>
                {#if errors.author_id}<p class="field-error">{errors.author_id}</p>{/if}
                <a class="small" href="/authors/new">Author missing? Add them first →</a>
                <label for="published_at">Publication date <span class="muted small">optional</span></label>
                <input
                    id="published_at"
                    name="published_at"
                    type="date"
                    value={current.published_at ?? ""}
                    aria-invalid={!!errors.published_at}
                />
                {#if errors.published_at}<p class="field-error">{errors.published_at}</p>{/if}
                <label for="description">Description <span class="muted small">optional</span></label>
                <textarea
                    id="description"
                    name="description"
                    rows="7"
                    maxlength="2000"
                    value={current.description ?? ""}
                    aria-invalid={!!errors.description}></textarea>
                {#if errors.description}<p class="field-error">{errors.description}</p>{/if}
            {:else}
                <label for="name">Author name <span class="required">*</span></label>
                <input id="name" name="name" required value={current.name ?? ""} aria-invalid={!!errors.name} />
                {#if errors.name}<p class="field-error">{errors.name}</p>{/if}
                <label for="bio">Biography <span class="muted small">optional</span></label>
                <textarea
                    id="bio"
                    name="bio"
                    rows="8"
                    maxlength="2000"
                    value={current.bio ?? ""}
                    aria-invalid={!!errors.bio}></textarea>
                {#if errors.bio}<p class="field-error">{errors.bio}</p>{/if}
            {/if}
            <div class="actions">
                <button disabled={pending}>{pending ? "Saving…" : editing ? "Save changes" : `Add ${noun}`}</button><a
                    class="button secondary"
                    href="/{collection}">Cancel</a
                >
            </div>
        </form>
    {/if}
</div>
