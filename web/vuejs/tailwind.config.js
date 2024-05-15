export default {
	content: [
		"./src/**/*.{html,vue}",
	],
	theme: {
		extend: {
			colors: {
				'wdy-green': '#1B8C30',
				'ora-orange': '#F7A823',
				'disabled': {
					'text': '#848484',
					'background': '#E2E2E2',
				},
				'placeholder-text': '#848484',
				'error': '#FF543E',
				'success': '#54FF3E',
				'darkmode-gray': '#374151',
				'ora-dropdown-background': '#2B2B2B',
			},
			boxShadow: {
				'ora-shadow': '0 3px 6px rgba(0, 0, 0, 0.16)',
			},
			borderRadius: {
				'2lg': '0.625rem',
			},
			padding: {
				'1.75': '0.4375rem'
			}

		},
	},

	mode: 'jit',
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
