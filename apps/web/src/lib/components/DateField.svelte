<script lang="ts">
    import { useI18n } from "$lib/i18n/context";
    import { fromJalali, toJalali } from "$lib/i18n/calendar";
    let {
        name,
        id = name,
        value = "",
        invalid = false,
    }: { name: string; id?: string; value?: string; invalid?: boolean } = $props();
    const i18n = useI18n();
    let draft = $state<string | null>(null);
    $effect(() => {
        void value;
        draft = null;
    });
    let iso = $derived(draft === null ? value : (fromJalali(draft) ?? ""));
    let display = $derived((draft ?? toJalali(value)).replace(/\d/g, (digit) => "۰۱۲۳۴۵۶۷۸۹"[Number(digit)]));
</script>

{#if i18n.locale === "fa"}
    <input
        {id}
        name="{name}_jalali"
        type="text"
        dir="ltr"
        placeholder="۱۴۰۵/۰۷/۱۳"
        value={display}
        aria-invalid={invalid}
        aria-describedby="{id}-calendar"
        oninput={(event) => {
            draft = event.currentTarget.value;
            event.currentTarget.setCustomValidity(draft && !fromJalali(draft) ? i18n.t("Enter a valid date.") : "");
        }}
    />
    <input type="hidden" {name} value={iso} />
    <span id="{id}-calendar" class="small muted"
        >{i18n.t("Jalali date")} — {i18n.t("Year")}/{i18n.t("Month")}/{i18n.t("Day")}</span
    >
{:else}
    <input
        {id}
        {name}
        type="date"
        value={iso}
        aria-invalid={invalid}
        oninput={(event) => {
            draft = toJalali(event.currentTarget.value);
        }}
    />
{/if}
