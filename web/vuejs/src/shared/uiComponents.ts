import type {Component} from 'vue';
import UiCheckbox from '@/components/UiCheckbox.vue';
import UiDivider from '@/components/UiDivider.vue';
import UiGrid from '@/components/UiGrid.vue';
import UiImage from '@/components/UiImage.vue';
import UiModal from '@/components/UiModal.vue';
import UiPasswordField from '@/components/UiPasswordField.vue';
import UiRadioButton from '@/components/UiRadioButton.vue';
import UiText from '@/components/UiText.vue';
import UiTextField from '@/components/UiTextField.vue';
import UiToggle from '@/components/UiToggle.vue';
import UiWindowTitle from '@/components/UiWindowTitle.vue';
import UiBox from '@/components/box/UiBox.vue';
import UiDatepicker from '@/components/datepicker/UiDatepicker.vue';
import UiHStack from '@/components/hstack/UiHStack.vue';
import UiScaffold from '@/components/scaffold/UiScaffold.vue';
import UiScrollView from '@/components/scrollview/UiScrollView.vue';
import UiSpacer from '@/components/spacer/UiSpacer.vue';
import UiTable from '@/components/table/UiTable.vue';
import UiTextLayout from '@/components/textlayout/UiTextLayout.vue';
import UiVStack from '@/components/vstack/UiVStack.vue';


import {
	Box,
	Checkbox,
	Component as NagoComponent,
	DatePicker,
	Divider,
	Grid,
	HStack,
	Img,
	Modal,
	PasswordField,
	Radiobutton,
	Scaffold,
	ScrollView,
	Spacer,
	Table,
	TextField,
	TextLayout,
	TextView,
	Toggle,
	VStack,
	WebView,
	WindowTitle
} from '@/shared/proto/nprotoc_gen';
import UiUnknownType from "@/components/UiUnknownType.vue";
import UiWebView from "@/components/UiWebView.vue";

/**
 * vueComponentFor returns an associated vue component for the given nago protocol component.
 * If new components are introduced, this method must be updated by hand, to type-switch and associate
 * the template component properly.
 */
export function vueComponentFor(ngc: NagoComponent): Component {
	if (ngc instanceof TextView) {
		return UiText;
	}

	if (ngc instanceof VStack) {
		return UiVStack;
	}

	if (ngc instanceof HStack) {
		return UiHStack;
	}

	if (ngc instanceof Img) {
		return UiImage;
	}

	if (ngc instanceof TextField) {
		return UiTextField;
	}

	if (ngc instanceof Toggle) {
		return UiToggle;
	}

	if (ngc instanceof Grid) {
		return UiGrid;
	}

	if (ngc instanceof Table) {
		return UiTable;
	}

	if (ngc instanceof DatePicker) {
		return UiDatepicker;
	}

	if (ngc instanceof PasswordField) {
		return UiPasswordField;
	}

	if (ngc instanceof Checkbox) {
		return UiCheckbox;
	}

	if (ngc instanceof Radiobutton) {
		return UiRadioButton;
	}

	if (ngc instanceof Box) {
		return UiBox;
	}

	if (ngc instanceof Spacer) {
		return UiSpacer;
	}

	if (ngc instanceof Modal) {
		return UiModal;
	}

	if (ngc instanceof WindowTitle) {
		return UiWindowTitle;
	}

	if (ngc instanceof ScrollView) {
		return UiScrollView;
	}

	if (ngc instanceof TextLayout) {
		return UiTextLayout;
	}

	if (ngc instanceof Scaffold) {
		return UiScaffold;
	}

	if (ngc instanceof Divider) {
		return UiDivider
	}

	if (ngc instanceof WebView) {
		return UiWebView;
	}

	// keep this as the default fallback
	return UiUnknownType;
}
