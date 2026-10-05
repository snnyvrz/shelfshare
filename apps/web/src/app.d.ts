declare global {
    namespace App {
        interface Locals {
            locale: import("$lib/i18n").Locale;
            user: { id: string; email: string } | null;
            token: string | null;
        }
    }
}
export {};
