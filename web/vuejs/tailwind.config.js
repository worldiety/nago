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


				customColors['primary'] = 'rgb(from var(--p0) r g b / <alpha-value>)';
				customColors['secondary'] = 'rgb(from var(--s0) r g b / <alpha-value>)';
				customColors['tertiary'] = 'rgb(from var(--t0) r g b / <alpha-value>)';

				customColors['clE'] = 'rgb(from var(--clE) r g b / <alpha-value>)';
				customColors['clW'] = 'rgb(from var(--clW) r g b / <alpha-value>)';
				customColors['clG'] = 'rgb(from var(--clG) r g b / <alpha-value>)';
				customColors['clI'] = 'rgb(from var(--clI) r g b / <alpha-value>)';
				customColors['clD'] = 'rgb(from var(--clD) r g b / <alpha-value>)';
				customColors['clT'] = 'rgb(from var(--clT) r g b / <alpha-value>)';

				//customColors['p0'] = 'hsl(var(--p0) / <alpha-value>)';
				customColors['p0'] = 'rgb(from var(--p0) r g b / <alpha-value>)';
				customColors['p1'] = 'rgb(from var(--p1) r g b / <alpha-value>)';
				customColors['p2'] = 'rgb(from var(--p2) r g b / <alpha-value>)';
				customColors['p3'] = 'rgb(from var(--p3) r g b / <alpha-value>)';
				customColors['p4'] = 'rgb(from var(--p4) r g b / <alpha-value>)';
				customColors['p5'] = 'rgb(from var(--p5) r g b / <alpha-value>)';
				customColors['p6'] = 'rgb(from var(--p6) r g b / <alpha-value>)';
				customColors['p7'] = 'rgb(from var(--p7) r g b / <alpha-value>)';
				customColors['p8'] = 'rgb(from var(--p8) r g b / <alpha-value>)';
				customColors['p9'] = 'rgb(from var(--p9) r g b / <alpha-value>)';

				customColors['s0'] = 'rgb(from var(--s0) r g b / <alpha-value>)';
				customColors['s1'] = 'rgb(from var(--s1) r g b / <alpha-value>)';
				customColors['s2'] = 'rgb(from var(--s2) r g b / <alpha-value>)';

				customColors['t0'] = 'rgb(from var(--s0) r g b / <alpha-value>)';
				customColors['t1'] = 'rgb(from var(--s1) r g b / <alpha-value>)';

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
