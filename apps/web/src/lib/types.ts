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
