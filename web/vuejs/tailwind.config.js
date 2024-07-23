import plugin from "tailwindcss/plugin";
import {NamedColor} from "./src/components/shared/colors.ts";

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
//					'error': '#FF543E',
//					'success': '#54FF3E',
					'darkmode-gray': '#374151',
					'ora-dropdown-background': '#2B2B2B',
				};
				const colorSteps = [10, 12, 14, 17, 22, 30, 60, 70, 83, 87, 90, 92, 94, 96, 98];
				colorSteps.forEach((step) => {
					customColors[`primary-${step}`] = `hsl(var(--primary-${step}) / <alpha-value>)`;
				});

				// TODO we cannot apply tailwind alpha value, because we have
				// 1. hex notation
				// 2. already applied alpha values anyway => this is a system collision between ora and tailwind
				customColors['primary'] = `var(--p0) `;
				customColors['secondary'] = `var(--s0) `;
				customColors['tertiary'] = `var(--t0)`;

				customColors['clE'] = `var(--clE)`;
				customColors['clW'] = `var(--clW)`;
				customColors['clG'] = `var(--clG)`;
				customColors['clI'] = `var(--clI)`;
				customColors['clD'] = `var(--clD)`;
				customColors['clT'] = `var(--clT)`;

				customColors['p0'] = `var(--p0) `;
				customColors['p1'] = `var(--p1) `;
				customColors['p2'] = `var(--p2) `;
				customColors['p3'] = `var(--p3) `;
				customColors['p4'] = `var(--p4) `;
				customColors['p5'] = `var(--p5) `;
				customColors['p6'] = `var(--p6) `;
				customColors['p7'] = `var(--p7) `;
				customColors['p8'] = `var(--p8) `;
				customColors['p9'] = `var(--p9) `;

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
	// keep this, because it is used by our (deprecated) grid and text components
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
