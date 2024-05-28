export default {
	content: [
		"./src/**/*.{html,vue}",
	],
	theme: {
		extend: {
			screens: {
				'xs': '400px',
			},
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
			},
			animation: {
				'spin-slow': 'spin 2s linear infinite',
			},
		},
	},
	darkMode: 'class',
}
