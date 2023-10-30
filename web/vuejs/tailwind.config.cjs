/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["./src/**/*.{html,vue}"],
    theme: {
        extend: {}
    },

    plugins: [],
    // safelist does not really work for our dynamic stuff
    // safelist: ["gap-[2fr]", "grid", "grid-cols-2",]
};
