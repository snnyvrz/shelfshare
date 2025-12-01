export default {
    "*.{js,mjs,ts,tsx}": ["bun x eslint --fix", "bun x prettier --write"],

    "*.go": (files) => [`gofmt -w ${files.join(" ")}`],

    "*.sh": ["shfmt -ci -i 4 -w"],

    "*.{json,md,css,scss,html}": ["bun x prettier --write"],

    "*.{yml,yaml}": (files) =>
        files
            .filter((file) => !file.match(/^charts\/[^/]+\/templates\//))
            .map((file) => `bun x prettier --write ${file}`),

    "charts/**": ["scripts/lint-helm.sh"],
};
