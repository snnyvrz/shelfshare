<script lang="ts">
    import { useI18n } from "$lib/i18n/context";
    const { t } = useI18n();
    import { enhance } from "$app/forms";
    import type { PageProps } from "./$types";
    let { data, form }: PageProps = $props();
    let pending = $state(false);
    let registering = $derived(data.mode === "register");
</script>

<svelte:head><title>{t(registering ? "Join the shelf" : "Log in")} · ShelfShare</title></svelte:head>
<div class="auth-layout">
    <div class="auth-intro">
        <p class="eyebrow">{t("A PLACE FOR YOUR LOVE OF BOOKS")}</p>
        <h1>{t(registering ? "Every reader brings a new story." : "Welcome back to the shelf.")}</h1>
        <p class="muted">
            {t("Put your physical books on your profile, borrow from other readers, and connect in real time.")}
        </p>
        <span class="auth-decoration" aria-hidden="true">✳</span>
    </div>
    <form
        class="panel auth-form"
        method="POST"
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
        <h2>{t(registering ? "Create your account" : "Make yourself at home")}</h2>
        <p class="muted">
            {t(registering ? "Your next chapter starts here." : "Log in to your shelf, loans, and conversations.")}
        </p>
        {#if form?.message}<div class="alert" role="alert">{t(form.message)}</div>{/if}
        {#if registering}<label for="displayName">{t("Public display name")}</label><input
                dir="auto"
                id="displayName"
                name="displayName"
                required
                maxlength="80"
                autocomplete="nickname"
                placeholder={t("How other readers know you")}
            />{/if}
        <label for="email">{t("Email address")}</label><input
            dir="ltr"
            id="email"
            name="email"
            type="email"
            autocomplete="email"
            required
            value={form?.email ?? ""}
            placeholder="you@example.com"
        />
        <label for="password">{t("Password")}</label><input
            dir="auto"
            id="password"
            name="password"
            type="password"
            autocomplete={registering ? "new-password" : "current-password"}
            required
        />
        <button disabled={pending}>{t(pending ? "One moment…" : registering ? "Join the shelf" : "Log in")}</button>
        <p class="small muted">
            {t(registering ? "Already have an account?" : "New to ShelfShare?")}
            <a href="/{registering ? 'login' : 'register'}?returnTo={encodeURIComponent(data.returnTo)}"
                >{t(registering ? "Log in" : "Create an account")}</a
            >
        </p>
    </form>
</div>
