import { fail } from "@sveltejs/kit";
import { api, ApiError, pageError, requireUser } from "$lib/server/api";
import { copyList } from "$lib/server/community";
import type { Book, BookList } from "$lib/types";
import type { Actions, PageServerLoad } from "./$types";
export const load: PageServerLoad = async (event) => {
    event.depends("shelfshare:community");
    requireUser(event);
    try {
        const q = event.url.searchParams.get("q") || "";
        const books = await api<BookList>(event, `/books?page_size=50&q=${encodeURIComponent(q)}`);
        const selectedBook = event.url.searchParams.get("book") || "";
        if (selectedBook && !books.data.some((book) => book.id === selectedBook))
            books.data.unshift((await api<{ data: Book }>(event, `/books/${encodeURIComponent(selectedBook)}`)).data);
        return {
            ...(await copyList(event, `/me/copies?page=${Number(event.url.searchParams.get("page")) || 1}`)),
            books: books.data,
            q,
            selectedBook,
            page: Number(event.url.searchParams.get("page")) || 1,
        };
    } catch (cause) {
        pageError(cause);
    }
};
async function change(event: Parameters<NonNullable<Actions[string]>>[0], action: "add" | "edit" | "archive") {
    requireUser(event);
    const form = await event.request.formData();
    const body = {
        bookId: String(form.get("bookId") || ""),
        condition: String(form.get("condition") || ""),
        notes: String(form.get("notes") || ""),
        visible: form.has("visible"),
        lendable: form.has("lendable"),
    };
    try {
        await api(event, action === "add" ? "/copies" : `/copies/${encodeURIComponent(String(form.get("id")))}`, {
            method: action === "add" ? "POST" : action === "edit" ? "PATCH" : "DELETE",
            ...(action !== "archive" ? { body: JSON.stringify(body) } : {}),
        });
        return { success: "Shelf updated." };
    } catch (cause) {
        if (!(cause instanceof ApiError)) throw cause;
        return fail(cause.status, { message: cause.message });
    }
}
export const actions: Actions = {
    add: (event) => change(event, "add"),
    edit: (event) => change(event, "edit"),
    archive: (event) => change(event, "archive"),
};
