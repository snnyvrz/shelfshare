export type Book = {
    id: string;
    title: string;
    description: string;
    author: { id: string; name: string; bio: string };
    published_at?: string;
    created_at: string;
    updated_at: string;
};
export type Author = {
    id: string;
    name: string;
    bio: string;
    books?: Omit<Book, "author">[];
    created_at: string;
    updated_at: string;
};
export type BookList = {
    data: Book[];
    pagination: { page: number; page_size: number; total: number; total_pages: number };
};
export type Collection = "books" | "authors";
export type FormValues = Record<string, string>;

export type Profile = { id: string; displayName: string; bio: string; location: string };
export type OwnProfile = Profile & { discoveryCityId: string; discoveryEnabled: boolean };
export type CityOption = { id: string; label: string };
export type PhysicalCopy = {
    id: string;
    bookId: string;
    ownerId: string;
    condition: string;
    notes: string;
    visible: boolean;
    lendable: boolean;
    archived: boolean;
    availability: string;
    book: { ID: string; Title: string; Author: { Name: string } };
};
export type BorrowRequest = {
    id: string;
    copyId: string;
    ownerId: string;
    borrowerId: string;
    status: string;
    message: string;
    dueAt: string | null;
    copy: PhysicalCopy;
};
export type Conversation = {
    id: string;
    userA: string;
    userB: string;
    initiator: string;
    status: string;
    requestId: string | null;
    request?: BorrowRequest;
    unread: number;
    peerReadAt?: string;
    lastMessage?: Message;
};
export type Message = { id: string; conversationId: string; senderId: string; body: string; createdAt: string };
