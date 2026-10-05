<script lang="ts">
    import { useI18n } from "$lib/i18n/context";
    import DateField from "./DateField.svelte";
    const { t } = useI18n();
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
    let draft = $state<FormValues>({});
    let current = $derived({ ...(form?.values ?? values), ...draft });
    let errors = $derived(form?.errors ?? {});
    let noun = $derived(collection === "books" ? "book" : "author");
    let heading = $derived(
        editing ? (noun === "book" ? "Edit book" : "Edit author") : noun === "book" ? "Add a book" : "Add an author"
    );
</script>

<svelte:head><title>{t(heading)} · ShelfShare</title></svelte:head>
<div class="narrow">
    <a class="back-link" href="/{collection}">{t(collection === "books" ? "← Back to books" : "← Back to authors")}</a>
    <p class="eyebrow">{t("GROW THE COLLECTION")}</p>
    <h1>{t(heading)}</h1>
    <p class="muted">{t("A little detail helps the next reader find their next great read.")}</p>
    {#if form?.message}<div class="alert" role="alert">{t(form.message)}</div>{/if}
    {#if collection === "books" && !authors.length}
        <div class="empty-state">
            <h2>{t("Start with an author")}</h2>
            <p>{t("Create an author before adding their book.")}</p>
            <a class="button" href="/authors/new">{t("Add an author")}</a>
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
                <label for="title">{t("Book title")} <span class="required">*</span></label>
                <input
                    dir="auto"
                    id="title"
                    name="title"
                    required
                    value={current.title ?? ""}
                    oninput={(event) => (draft.title = event.currentTarget.value)}
                    aria-invalid={!!errors.title}
                    aria-describedby={errors.title ? "title-error" : undefined}
                />
                {#if errors.title}<p id="title-error" class="field-error">{t(errors.title)}</p>{/if}
                <label for="author_id">{t("Author")} <span class="required">*</span></label>
                <select
                    id="author_id"
                    name="author_id"
                    required
                    value={current.author_id ?? ""}
                    onchange={(event) => (draft.author_id = event.currentTarget.value)}
                    aria-invalid={!!errors.author_id}
                >
                    <option value="" disabled>{t("Choose an author")}</option>
                    {#each authors as author (author.id)}<option value={author.id}>{author.name}</option>{/each}
                </select>
                {#if errors.author_id}<p class="field-error">{t(errors.author_id)}</p>{/if}
                <a class="small" href="/authors/new">{t("Author missing? Add them first →")}</a>
                <label for="published_at"
                    >{t("Publication date")} <span class="muted small">{t("optional")}</span></label
                >
                <DateField
                    id="published_at"
                    name="published_at"
                    value={current.published_at ?? ""}
                    invalid={!!errors.published_at}
                />
                {#if errors.published_at}<p class="field-error">{t(errors.published_at)}</p>{/if}
                <label for="description">{t("Description")} <span class="muted small">{t("optional")}</span></label>
                <textarea
                    dir="auto"
                    id="description"
                    name="description"
                    rows="7"
                    maxlength="2000"
                    value={current.description ?? ""}
                    oninput={(event) => (draft.description = event.currentTarget.value)}
                    aria-invalid={!!errors.description}></textarea>
                {#if errors.description}<p class="field-error">{t(errors.description)}</p>{/if}
            {:else}
                <label for="name">{t("Author name")} <span class="required">*</span></label>
                <input
                    dir="auto"
                    id="name"
                    name="name"
                    required
                    value={current.name ?? ""}
                    oninput={(event) => (draft.name = event.currentTarget.value)}
                    aria-invalid={!!errors.name}
                />
                {#if errors.name}<p class="field-error">{t(errors.name)}</p>{/if}
                <label for="bio">{t("Biography")} <span class="muted small">{t("optional")}</span></label>
                <textarea
                    dir="auto"
                    id="bio"
                    name="bio"
                    rows="8"
                    maxlength="2000"
                    value={current.bio ?? ""}
                    oninput={(event) => (draft.bio = event.currentTarget.value)}
                    aria-invalid={!!errors.bio}></textarea>
                {#if errors.bio}<p class="field-error">{t(errors.bio)}</p>{/if}
            {/if}
            <div class="actions">
                <button disabled={pending}
                    >{t(
                        pending ? "Saving…" : editing ? "Save changes" : noun === "book" ? "Add book" : "Add author"
                    )}</button
                ><a class="button secondary" href="/{collection}">{t("Cancel")}</a>
                >
            </div>
        </form>
    {/if}
</div>
