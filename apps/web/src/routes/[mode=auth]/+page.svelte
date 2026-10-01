<script lang="ts">
    import { enhance } from "$app/forms";
    import type { PageProps } from "./$types";
    let { data, form }: PageProps = $props();
    let pending = $state(false);
    let registering = $derived(data.mode === "register");
</script>

<svelte:head><title>{registering ? "Join the shelf" : "Log in"} · ShelfShare</title></svelte:head>
<div class="auth-layout">
    <div class="auth-intro">
        <p class="eyebrow">A PLACE FOR YOUR LOVE OF BOOKS</p>
        <h1>{registering ? "Every reader brings a new story." : "Welcome back to the shelf."}</h1>
        <p class="muted">
            Join a growing collection of books and the people who love them. Discover something new. Share something
            wonderful.
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
        <h2>{registering ? "Create your account" : "Make yourself at home"}</h2>
        <p class="muted">{registering ? "Your next chapter starts here." : "Log in to help grow the collection."}</p>
        {#if form?.message}<div class="alert" role="alert">{form.message}</div>{/if}
        <label for="email">Email address</label><input
            id="email"
            name="email"
            type="email"
            autocomplete="email"
            required
            value={form?.email ?? ""}
            placeholder="you@example.com"
        />
        <label for="password">Password</label><input
            id="password"
            name="password"
            type="password"
            autocomplete={registering ? "new-password" : "current-password"}
            required
        />
        <button disabled={pending}>{pending ? "One moment…" : registering ? "Join the shelf" : "Log in"}</button>
        <p class="small muted">
            {registering ? "Already have an account?" : "New to ShelfShare?"}
            <a href="/{registering ? 'login' : 'register'}?returnTo={encodeURIComponent(data.returnTo)}"
                >{registering ? "Log in" : "Create an account"}</a
            >
        </p>
    </form>
</div>
