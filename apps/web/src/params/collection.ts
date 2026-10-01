export function match(value: string): value is "books" | "authors" {
    return value === "books" || value === "authors";
}
