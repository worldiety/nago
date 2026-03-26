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
import type { Component as NagoComponent } from '@/shared/proto/nprotoc_gen';
import { Canvas } from '@/shared/proto/nprotoc_gen';
import {
	Accordion,
	BarChart,
	Box,
	Checkbox,
	CodeEditor,
	CountDown,
	DatePicker,
	Divider,
	DnDArea,
	Form,
	Grid,
	HoverGroup,
	Img,
	LineChart,
	Menu,
	Modal,
	PasswordField,
	PieChart,
	QrCode,
	QrCodeReader,
	Radiobutton,
	RichText,
	RichTextEditor,
	Scaffold,
	ScrollView,
	Select,
	Spacer,
	Stack,
	Switcher,
	Table,
	TextField,
	TextLayout,
	TextView,
	Toggle,
	Video,
	WebView,
	WindowTitle,
} from '@/shared/proto/nprotoc_gen';

const LazyUiAccordion = defineAsyncComponent(() => import('@/components/UiAccordion.vue'));
const LazyUiBarChart = defineAsyncComponent(() => import('@/components/charts/UiBarChart.vue'));
const LazyUiBox = defineAsyncComponent(() => import('@/components/box/UiBox.vue'));
const LazyUiCanvas = defineAsyncComponent(() => import('@/components/canvas/UiCanvas.vue'));
const LazyUiCheckbox = defineAsyncComponent(() => import('@/components/UiCheckbox.vue'));
const LazyUiCodeEditor = defineAsyncComponent(() => import('@/components/codeeditor/UiCodeEditor.vue'));
const LazyUiCountDown = defineAsyncComponent(() => import('@/components/countdown/UiCountDown.vue'));
const LazyUiDatepicker = defineAsyncComponent(() => import('@/components/datepicker/UiDatepicker.vue'));
const LazyUiDivider = defineAsyncComponent(() => import('@/components/UiDivider.vue'));
const LazyUiDnDArea = defineAsyncComponent(() => import('@/components/dnd/UiDnDArea.vue'));
const LazyUiForm = defineAsyncComponent(() => import('@/components/form/UiForm.vue'));
const LazyUiGrid = defineAsyncComponent(() => import('@/components/UiGrid.vue'));
const LazyUiHoverGroup = defineAsyncComponent(() => import('@/components/hovergroup/UiHoverGroup.vue'));
const LazyUiImage = defineAsyncComponent(() => import('@/components/UiImage.vue'));
const LazyUiLineChart = defineAsyncComponent(() => import('@/components/charts/UiLineChart.vue'));
const LazyUiMenu = defineAsyncComponent(() => import('@/components/menu/UiMenu.vue'));
const LazyUiModal = defineAsyncComponent(() => import('@/components/UiModal.vue'));
const LazyUiPasswordField = defineAsyncComponent(() => import('@/components/UiPasswordField.vue'));
const LazyUiPieChart = defineAsyncComponent(() => import('@/components/charts/UiPieChart.vue'));
const LazyUiQrCode = defineAsyncComponent(() => import('@/components/UiQrCode.vue'));
const LazyUiQrCodeReader = defineAsyncComponent(() => import('@/components/UiQrCodeReader.vue'));
const LazyUiRadioButton = defineAsyncComponent(() => import('@/components/UiRadioButton.vue'));
const LazyUiRichText = defineAsyncComponent(() => import('@/components/richtexteditor/UiRichText.vue'));
const LazyUiRichTextEditor = defineAsyncComponent(() => import('@/components/richtexteditor/UiRichTextEditor.vue'));
const LazyUiScaffold = defineAsyncComponent(() => import('@/components/scaffold/UiScaffold.vue'));
const LazyUiScrollView = defineAsyncComponent(() => import('@/components/scrollview/UiScrollView.vue'));
const LazyUiSelect = defineAsyncComponent(() => import('@/components/UiSelect.vue'));
const LazyUiSpacer = defineAsyncComponent(() => import('@/components/spacer/UiSpacer.vue'));
const LazyUiStack = defineAsyncComponent(() => import('@/components/UiStack.vue'));
const LazyUiSwitcher = defineAsyncComponent(() => import('@/components/switcher/UiSwitcher.vue'));
const LazyUiTable = defineAsyncComponent(() => import('@/components/table/UiTable.vue'));
const LazyUiText = defineAsyncComponent(() => import('@/components/UiText.vue'));
const LazyUiTextField = defineAsyncComponent(() => import('@/components/UiTextField.vue'));
const LazyUiTextLayout = defineAsyncComponent(() => import('@/components/textlayout/UiTextLayout.vue'));
const LazyUiToggle = defineAsyncComponent(() => import('@/components/UiToggle.vue'));
const LazyUiUnknownType = defineAsyncComponent(() => import('@/components/UiUnknownType.vue'));
const LazyUiVideo = defineAsyncComponent(() => import('@/components/video/UiVideo.vue'));
const LazyUiWebView = defineAsyncComponent(() => import('@/components/UiWebView.vue'));
const LazyUiWindowTitle = defineAsyncComponent(() => import('@/components/UiWindowTitle.vue'));
/**
 * vueComponentFor returns an associated vue component for the given nago protocol component.
 * If new components are introduced, this method must be updated by hand, to type-switch and associate
 * the template component properly.
 */
export function vueComponentFor(ngc: NagoComponent): Component {
	if (ngc instanceof Accordion) {
		return LazyUiAccordion;
	}

	if (ngc instanceof BarChart) {
		return LazyUiBarChart;
	}

	if (ngc instanceof Box) {
		return LazyUiBox;
	}

	if (ngc instanceof Canvas) {
		return LazyUiCanvas;
	}

	if (ngc instanceof Checkbox) {
		return LazyUiCheckbox;
	}

	if (ngc instanceof CodeEditor) {
		return LazyUiCodeEditor;
	}

	if (ngc instanceof CountDown) {
		return LazyUiCountDown;
	}

	if (ngc instanceof DatePicker) {
		return LazyUiDatepicker;
	}

	if (ngc instanceof Divider) {
		return LazyUiDivider;
	}

	if (ngc instanceof DnDArea) {
		return LazyUiDnDArea;
	}

	if (ngc instanceof Form) {
		return LazyUiForm;
	}

	if (ngc instanceof Grid) {
		return LazyUiGrid;
	}

	if (ngc instanceof HoverGroup) {
		return LazyUiHoverGroup;
	}

	if (ngc instanceof Img) {
		return LazyUiImage;
	}

	if (ngc instanceof LineChart) {
		return LazyUiLineChart;
	}

	if (ngc instanceof Menu) {
		return LazyUiMenu;
	}

	if (ngc instanceof Modal) {
		return LazyUiModal;
	}

	if (ngc instanceof PasswordField) {
		return LazyUiPasswordField;
	}

	if (ngc instanceof PieChart) {
		return LazyUiPieChart;
	}

	if (ngc instanceof QrCode) {
		return LazyUiQrCode;
	}

	if (ngc instanceof QrCodeReader) {
		return LazyUiQrCodeReader;
	}

	if (ngc instanceof Radiobutton) {
		return LazyUiRadioButton;
	}

	if (ngc instanceof RichText) {
		return LazyUiRichText;
	}

	if (ngc instanceof RichTextEditor) {
		return LazyUiRichTextEditor;
	}

	if (ngc instanceof Scaffold) {
		return LazyUiScaffold;
	}

	if (ngc instanceof ScrollView) {
		return LazyUiScrollView;
	}

	if (ngc instanceof Select) {
		return LazyUiSelect;
	}

	if (ngc instanceof Spacer) {
		return LazyUiSpacer;
	}

	if (ngc instanceof Stack) {
		return LazyUiStack;
	}

	if (ngc instanceof Switcher) {
		return LazyUiSwitcher;
	}

	if (ngc instanceof Table) {
		return LazyUiTable;
	}

	if (ngc instanceof TextField) {
		return LazyUiTextField;
	}

	if (ngc instanceof TextLayout) {
		return LazyUiTextLayout;
	}

	if (ngc instanceof TextView) {
		return LazyUiText;
	}

	if (ngc instanceof Toggle) {
		return LazyUiToggle;
	}

	if (ngc instanceof Video) {
		return LazyUiVideo;
	}

	if (ngc instanceof WebView) {
		return LazyUiWebView;
	}

	if (ngc instanceof WindowTitle) {
		return LazyUiWindowTitle;
	}

	// keep this as the default fallback
	return LazyUiUnknownType;
}
