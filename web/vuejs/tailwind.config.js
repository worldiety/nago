import plugin from "tailwindcss/plugin";

export default {
	content: [
		"./src/**/*.{html,vue}",
	],
	plugins: [
		plugin(function({addVariant}) {
			addVariant('darkmode', '.darkmode &');
			addVariant('contrast', '.contrast &');
			addVariant('protanopia', '.protanopia &');
			addVariant('deuteranopia', '.deuteranopia &');
			addVariant('tritanopia', '.tritanopia &');
		})
	],
	theme: {
		extend: {
			screens: {
				'xs': '400px',
			},
			colors: ({ colors }) => {
				const customColors = {
					// generic colors
					// TODO: Remove legacy non-generic colors
					'disabled': {
						'text': '#848484',
						'background': '#E2E2E2',
					},
					'placeholder-text': '#848484',
					'error': '#FF543E',
					'success': '#54FF3E',
					'darkmode-gray': '#374151',
					'ora-dropdown-background': '#2B2B2B',
				};
				const colorSteps = [10, 12, 14, 17, 22, 30, 60, 70, 83, 87, 90, 92, 94, 96, 98];
				colorSteps.forEach((step) => {
					customColors[`primary-${step}`] = `hsl(var(--primary-${step}) / <alpha-value>)`;
				});
				customColors['primary'] = `hsl(var(--primary) / <alpha-value>)`;
				customColors['secondary'] = `hsl(var(--secondary) / <alpha-value>)`;
				customColors['tertiary'] = `hsl(var(--tertiary) / <alpha-value>)`;
				customColors['background'] = `hsl(var(--background) / <alpha-value>)`;
				return customColors;
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
}
