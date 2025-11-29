export default {
    "*.{js,mjs,ts,tsx}": ["bun x eslint --fix", "bun x prettier --write"],

    "*.go": (files) => [`gofmt -w ${files.join(" ")}`],

    "*.sh": ["shfmt -ci -i 4 -w"],
};
