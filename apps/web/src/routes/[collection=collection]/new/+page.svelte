<script lang="ts">
    import Editor from "$lib/components/Editor.svelte";
    import type { PageProps } from "./$types";
    let { data, form }: PageProps = $props();
</script>

<Editor {...data} {form} />
