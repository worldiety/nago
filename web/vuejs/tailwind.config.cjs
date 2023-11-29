/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./src/**/*.{html,vue}",
        'node_modules/flowbite-vue/**/*.{js,jsx,ts,tsx}',
        'node_modules/flowbite/**/*.{js,jsx,ts,tsx}'
    ],
    theme: {
        extend: {}
    },

    plugins: [
        require('flowbite/plugin')
    ],
    darkMode: 'class',
    // safelist does not really work for our dynamic stuff
    // safelist: ["gap-[2fr]", "grid", "grid-cols-2",]
};
