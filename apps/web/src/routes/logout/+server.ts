import { redirect } from "@sveltejs/kit";
import { sessionCookie } from "$lib/server/session";
import type { RequestHandler } from "./$types";

export const POST: RequestHandler = ({ cookies }) => {
    cookies.delete(sessionCookie, { path: "/" });
    redirect(303, "/");
};
