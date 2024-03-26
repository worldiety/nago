export default {
	content: [
		"./src/**/*.{html,vue}",
		'node_modules/flowbite-vue/**/*.{js,jsx,ts,tsx}',
		'node_modules/flowbite/**/*.{js,jsx,ts,tsx}'
	],
	theme: {
		extend: {},
		colors: {
			'wdy-green': '#1B8C30',
			'ora-orange': '#F7A823',
			'disabled': {
				'text': '#848484',
				'background': '#E2E2E2',
			},
			'error': '#FF543E',
		},
	},

	mode: 'jit',
	plugins: [
		require('flowbite/plugin')
	],
	darkMode: 'class',
	// safelist does not really work for our dynamic stuff
	// safelist: ["gap-[2fr]", "grid", "grid-cols-2",]
	safelist: ["grid",
		"gap-1","gap-2","gap-3","gap-4","text-sm","text-base","text-lg","text-xl","text-2xl",
		{ pattern: /grid-cols-(1|2|3|4|5|6|7|8|9|10|11|12)/ , variants: ['sm', 'md', 'lg', 'xl', '2xl'],},
		{ pattern: /col-span-(1|2|3|4|5|6|7|8|9|10|11|12)/ , variants: ['sm', 'md', 'lg', 'xl', '2xl'],},
		{ pattern: /grid-rows-(1|2|3|4|5|6)/ },
		{ pattern: /col-start-(1|2|3|4|5|6|7|8|9|10|11|12)/ },
		{ pattern: /col-end-(1|2|3|4|5|6|7|8|9|10|11|12)/ },
		{ pattern: /row-start-(1|2|3|4|5|6|7|8|9|10|11|12)/ },
		{ pattern: /row-end-(1|2|3|4|5|6|7|8|9|10|11|12)/ },
	]
}
