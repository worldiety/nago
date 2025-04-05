/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

import plugin from "tailwindcss/plugin";

export default {
	content: [
		"./src/**/*.{html,vue}",
	],
	plugins: [
		plugin(function ({addVariant}) {
			addVariant('darkmode', '.darkmode &');
			addVariant('contrast', '.contrast &');
			addVariant('protanopia', '.protanopia &');
			addVariant('deuteranopia', '.deuteranopia &');
			addVariant('tritanopia', '.tritanopia &');
		}),
		require('@tailwindcss/typography'),

	],

	theme: {
		extend: {
			screens: {
				'xs': '400px',
			},
	/*		typography: (theme) => ({
				DEFAULT: {
					css: {
						'h1, h2, h3': {
							color: theme('colors.M8'),
						'code':{
							color: theme('colors.M8'),
						}
						},
					},
				},
			}),*/
			colors: ({colors}) => {
				const customColors = {
					// generic colors
					// TODO: Remove legacy non-generic colors
//					'disabled': {
//						'text': '#848484',
//						'background': '#E2E2E2',
//					},
//					'placeholder-text': '#848484',
//					'error': '#FF543E',
//					'success': '#54FF3E',
//					'darkmode-gray': '#374151',
//					'ora-dropdown-background': '#2B2B2B',
				};
				// const colorSteps = [10, 12, 14, 17, 22, 30, 60, 70, 83, 87, 90, 92, 94, 96, 98];
				// colorSteps.forEach((step) => {
				// 	customColors[`primary-${step}`] = `hsl(var(--primary-${step}) / <alpha-value>)`;
				// });
				//
				//
				// customColors['primary'] = 'rgb(from var(--p0) r g b / <alpha-value>)';
				// customColors['secondary'] = 'rgb(from var(--s0) r g b / <alpha-value>)';
				// customColors['tertiary'] = 'rgb(from var(--t0) r g b / <alpha-value>)';
				//customColors['p0'] = 'hsl(var(--p0) / <alpha-value>)';

				// defines semantic ORA colors
				customColors['SE0'] = 'rgb(from var(--SE0) r g b / <alpha-value>)';
				customColors['SW0'] = 'rgb(from var(--SW0) r g b / <alpha-value>)';
				customColors['SG0'] = 'rgb(from var(--SG0) r g b / <alpha-value>)';
				customColors['SV0'] = 'rgb(from var(--SV0) r g b / <alpha-value>)';
				customColors['SI0'] = 'rgb(from var(--SI0) r g b / <alpha-value>)';
				customColors['ST0'] = 'rgb(from var(--ST0) r g b / <alpha-value>)';

				// defined main ORA colors
				customColors['M0'] = 'rgb(from var(--M0) r g b / <alpha-value>)';
				customColors['M1'] = 'rgb(from var(--M1) r g b / <alpha-value>)';
				customColors['M2'] = 'rgb(from var(--M2) r g b / <alpha-value>)';
				customColors['M3'] = 'rgb(from var(--M3) r g b / <alpha-value>)';
				customColors['M4'] = 'rgb(from var(--M4) r g b / <alpha-value>)';
				customColors['M5'] = 'rgb(from var(--M5) r g b / <alpha-value>)';
				customColors['M6'] = 'rgb(from var(--M6) r g b / <alpha-value>)';
				customColors['M7'] = 'rgb(from var(--M7) r g b / <alpha-value>)';
				customColors['M8'] = 'rgb(from var(--M8) r g b / <alpha-value>)';
				customColors['M9'] = 'rgb(from var(--M9) r g b / <alpha-value>)';

				// defined accent ORA colors
				customColors['A0'] = 'rgb(from var(--A0) r g b / <alpha-value>)';
				customColors['A1'] = 'rgb(from var(--A1) r g b / <alpha-value>)';
				customColors['A2'] = 'rgb(from var(--A2) r g b / <alpha-value>)';

				// defined interactive ORA colors
				customColors['I0'] = 'rgb(from var(--I0) r g b / <alpha-value>)';
				customColors['I1'] = 'rgb(from var(--I1) r g b / <alpha-value>)';

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
	safelist: [
		// "grid",
		// "gap-1", "gap-2", "gap-3", "gap-4", "text-sm", "text-base", "text-lg", "text-xl", "text-2xl",
		// {pattern: /grid-cols-(1|2|3|4|5|6|7|8|9|10|11|12)/, variants: ['sm', 'md', 'lg', 'xl', '2xl'],},
		// {pattern: /col-span-(1|2|3|4|5|6|7|8|9|10|11|12)/, variants: ['sm', 'md', 'lg', 'xl', '2xl'],},
		// {pattern: /grid-rows-(1|2|3|4|5|6)/},
		// {pattern: /col-start-(1|2|3|4|5|6|7|8|9|10|11|12)/},
		// {pattern: /col-end-(1|2|3|4|5|6|7|8|9|10|11|12)/},
		// {pattern: /row-start-(1|2|3|4|5|6|7|8|9|10|11|12)/},
		// {pattern: /row-end-(1|2|3|4|5|6|7|8|9|10|11|12)/},
	]

}
