/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import type { Component } from 'vue';
import { defineAsyncComponent } from 'vue';
import {
	BarChart,
	Box,
	Checkbox,
	CodeEditor,
	CountDown,
	DatePicker,
	Divider,
	Form,
	Grid,
	HStack,
	HoverGroup,
	Img,
	LineChart,
	Menu,
	Modal,
	Component as NagoComponent,
	PasswordField,
	PieChart,
	QrCode,
	QrCodeReader,
	Radiobutton,
	RichText,
	RichTextEditor,
	Scaffold,
	ScrollView,
	Spacer,
	Table,
	TextField,
	TextLayout,
	TextView,
	Toggle,
	VStack,
	Video,
	WebView,
	WindowTitle,
} from '@/shared/proto/nprotoc_gen';

const LazyUiCheckbox = defineAsyncComponent(() => import('@/components/UiCheckbox.vue'));
const LazyUiDivider = defineAsyncComponent(() => import('@/components/UiDivider.vue'));
const LazyUiGrid = defineAsyncComponent(() => import('@/components/UiGrid.vue'));
const LazyUiImage = defineAsyncComponent(() => import('@/components/UiImage.vue'));
const LazyUiModal = defineAsyncComponent(() => import('@/components/UiModal.vue'));
const LazyUiPasswordField = defineAsyncComponent(() => import('@/components/UiPasswordField.vue'));
const LazyUiRadioButton = defineAsyncComponent(() => import('@/components/UiRadioButton.vue'));
const LazyUiText = defineAsyncComponent(() => import('@/components/UiText.vue'));
const LazyUiTextField = defineAsyncComponent(() => import('@/components/UiTextField.vue'));
const LazyUiToggle = defineAsyncComponent(() => import('@/components/UiToggle.vue'));
const LazyUiUnknownType = defineAsyncComponent(() => import('@/components/UiUnknownType.vue'));
const LazyUiWebView = defineAsyncComponent(() => import('@/components/UiWebView.vue'));
const LazyUiWindowTitle = defineAsyncComponent(() => import('@/components/UiWindowTitle.vue'));
const LazyUiBox = defineAsyncComponent(() => import('@/components/box/UiBox.vue'));
const LazyUiCountDown = defineAsyncComponent(() => import('@/components/countdown/UiCountDown.vue'));
const LazyUiDatepicker = defineAsyncComponent(() => import('@/components/datepicker/UiDatepicker.vue'));
const LazyUiForm = defineAsyncComponent(() => import('@/components/form/UiForm.vue'));
const LazyUiHoverGroup = defineAsyncComponent(() => import('@/components/hovergroup/UiHoverGroup.vue'));
const LazyUiHStack = defineAsyncComponent(() => import('@/components/hstack/UiHStack.vue'));
const LazyUiMenu = defineAsyncComponent(() => import('@/components/menu/UiMenu.vue'));
const LazyUiRichText = defineAsyncComponent(() => import('@/components/richtexteditor/UiRichText.vue'));
const LazyUiRichTextEditor = defineAsyncComponent(() => import('@/components/richtexteditor/UiRichTextEditor.vue'));
const LazyUiScaffold = defineAsyncComponent(() => import('@/components/scaffold/UiScaffold.vue'));
const LazyUiScrollView = defineAsyncComponent(() => import('@/components/scrollview/UiScrollView.vue'));
const LazyUiSpacer = defineAsyncComponent(() => import('@/components/spacer/UiSpacer.vue'));
const LazyUiTable = defineAsyncComponent(() => import('@/components/table/UiTable.vue'));
const LazyUiTextLayout = defineAsyncComponent(() => import('@/components/textlayout/UiTextLayout.vue'));
const LazyUiVStack = defineAsyncComponent(() => import('@/components/vstack/UiVStack.vue'));
const LazyUiCodeEditor = defineAsyncComponent(() => import('@/components/codeeditor/UiCodeEditor.vue'));
const LazyUiQrCode = defineAsyncComponent(() => import('@/components/UiQrCode.vue'));
const LazyUiQrCodeReader = defineAsyncComponent(() => import('@/components/UiQrCodeReader.vue'));
const LazyUiBarChart = defineAsyncComponent(() => import('@/components/charts/UiBarChart.vue'));
const LazyUiLineChart = defineAsyncComponent(() => import('@/components/charts/UiLineChart.vue'));
const LazyUiVideo = defineAsyncComponent(() => import('@/components/video/UiVideo.vue'));
const LazyUiPieChart = defineAsyncComponent(() => import('@/components/charts/UiPieChart.vue'));

/**
 * vueComponentFor returns an associated vue component for the given nago protocol component.
 * If new components are introduced, this method must be updated by hand, to type-switch and associate
 * the template component properly.
 */
export function vueComponentFor(ngc: NagoComponent): Component {
	if (ngc instanceof TextView) {
		return LazyUiText;
	}

	if (ngc instanceof VStack) {
		return LazyUiVStack;
	}

	if (ngc instanceof HStack) {
		return LazyUiHStack;
	}

	if (ngc instanceof Img) {
		return LazyUiImage;
	}

	if (ngc instanceof TextField) {
		return LazyUiTextField;
	}

	if (ngc instanceof Toggle) {
		return LazyUiToggle;
	}

	if (ngc instanceof Grid) {
		return LazyUiGrid;
	}

	if (ngc instanceof Table) {
		return LazyUiTable;
	}

	if (ngc instanceof DatePicker) {
		return LazyUiDatepicker;
	}

	if (ngc instanceof PasswordField) {
		return LazyUiPasswordField;
	}

	if (ngc instanceof Checkbox) {
		return LazyUiCheckbox;
	}

	if (ngc instanceof Radiobutton) {
		return LazyUiRadioButton;
	}

	if (ngc instanceof Box) {
		return LazyUiBox;
	}

	if (ngc instanceof Spacer) {
		return LazyUiSpacer;
	}

	if (ngc instanceof Modal) {
		return LazyUiModal;
	}

	if (ngc instanceof WindowTitle) {
		return LazyUiWindowTitle;
	}

	if (ngc instanceof ScrollView) {
		return LazyUiScrollView;
	}

	if (ngc instanceof TextLayout) {
		return LazyUiTextLayout;
	}

	if (ngc instanceof Scaffold) {
		return LazyUiScaffold;
	}

	if (ngc instanceof Divider) {
		return LazyUiDivider;
	}

	if (ngc instanceof WebView) {
		return LazyUiWebView;
	}

	if (ngc instanceof Menu) {
		return LazyUiMenu;
	}

	if (ngc instanceof Form) {
		return LazyUiForm;
	}

	if (ngc instanceof CountDown) {
		return LazyUiCountDown;
	}

	if (ngc instanceof CodeEditor) {
		return LazyUiCodeEditor;
	}

	if (ngc instanceof RichText) {
		return LazyUiRichText;
	}

	if (ngc instanceof RichTextEditor) {
		return LazyUiRichTextEditor;
	}

	if (ngc instanceof HoverGroup) {
		return LazyUiHoverGroup;
	}

	if (ngc instanceof QrCode) {
		return LazyUiQrCode;
	}

	if (ngc instanceof QrCodeReader) {
		return LazyUiQrCodeReader;
	}

	if (ngc instanceof BarChart) {
		return LazyUiBarChart;
	}

	if (ngc instanceof LineChart) {
		return LazyUiLineChart;
	}

	if (ngc instanceof Video) {
		return LazyUiVideo;
	}

	if (ngc instanceof PieChart) {
		return LazyUiPieChart;
	}

	// keep this as the default fallback
	return LazyUiUnknownType;
}
