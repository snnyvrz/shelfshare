import { createServer } from "node:http";

// In-memory fixture API: localization browser tests never access application databases.
const userId = "000000000000000000000001";
const peerId = "000000000000000000000002";
const author = { id: "author", name: "نویسنده / Author", bio: "", books: [] };
let book = {
    id: "book",
    title: "کتاب / Book",
    description: "",
    author,
    published_at: "2026-10-05",
    created_at: "2026-10-05",
    updated_at: "2026-10-05",
};
const profile = {
    id: userId,
    displayName: "خواننده / Reader",
    bio: "",
    location: "تهران / Tehran, ایران / Iran",
    discoveryCityId: "112931",
    discoveryEnabled: false,
};
const cities = [{ id: "112931", label: profile.location, names: { en: "Tehran, Iran", fa: "تهران، ایران" } }];
const copy = {
    id: "copy",
    bookId: book.id,
    ownerId: userId,
    condition: "خوب / Good",
    notes: "",
    visible: true,
    lendable: true,
    archived: false,
    availability: "available",
    book: { ID: book.id, Title: book.title, Author: { Name: author.name } },
};
const loan = {
    id: "loan",
    ownerId: userId,
    borrowerId: peerId,
    copyId: copy.id,
    status: "accepted",
    message: "سلام / Hello",
    dueAt: "2027-03-21T23:59:59Z",
    copy,
};
const conversation = {
    id: "conversation",
    userA: userId,
    userB: peerId,
    initiator: userId,
    status: "accepted",
    requestId: null,
    unread: 0,
};
let lastDueAt = "";

createServer(async (request, response) => {
    const url = new URL(request.url || "/", "http://localhost");
    const path = url.pathname;
    response.setHeader("Content-Type", "application/json");
    const send = (value: unknown) => response.end(JSON.stringify(value));
    if (path === "/health") return send({ ok: true });
    if (path === "/test/state") return send({ book, lastDueAt });
    if (path === "/api/auth/cities") return send({ data: cities });
    if (path === "/api/auth/me") return send(profile);
    if (path.startsWith("/api/auth/profiles/")) return send({ ...profile, id: path.split("/").at(-1) });
    if (path === "/api/auth/profiles") return send({ data: [profile], total: 1 });
    if (path === "/api/auth/login") {
        response.statusCode = 401;
        return send({ message: "Invalid credentials" });
    }
    if (path === "/api/auth/register") {
        response.statusCode = 409;
        return send({ message: "User already exists" });
    }
    if (path === "/api/auth/nearby/profiles") return send({ data: [profile], total: 1 });
    if (path === "/api/auth/nearby/owners") return send({ data: [{ ownerId: userId, rank: 0 }] });
    if (path === "/api/authors") return send([{ data: author }]);
    if (path === "/api/books" && request.method === "POST") {
        let body = "";
        for await (const chunk of request) body += chunk;
        book = { ...book, ...JSON.parse(body) };
        return send({ data: book });
    }
    if (path === "/api/books")
        return send({ data: [book], pagination: { page: 1, page_size: 12, total: 1, total_pages: 1 } });
    if (path === "/api/books/book") return send({ data: book });
    if (["/api/copies", "/api/me/copies", "/api/nearby/copies"].includes(path)) return send({ data: [copy], total: 1 });
    if (path === "/api/requests") return send({ data: [loan], total: 1 });
    if (path === "/api/requests/loan/handover") {
        let body = "";
        for await (const chunk of request) body += chunk;
        lastDueAt = JSON.parse(body).dueAt || "";
        return send({});
    }
    if (path === "/api/conversations") return send({ data: [conversation], total: 1 });
    if (path === "/api/conversations/conversation/messages") return send({ data: [] });
    if (path === "/api/conversations/conversation/read") return send({});
    if (path === "/api/blocks") return send({ data: [] });
    response.statusCode = 404;
    send({ message: "Resource not found" });
}).listen(31303, "127.0.0.1");
