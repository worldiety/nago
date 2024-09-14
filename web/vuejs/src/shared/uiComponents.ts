import UiDivider from '@/components/UiDivider.vue';
import UiGrid from '@/components/UiGrid.vue';
import UiImage from '@/components/UiImage.vue';
import UiScaffold from '@/components/scaffold/UiScaffold.vue';
import UiText from '@/components/UiText.vue';
import UiTextField from '@/components/UiTextField.vue';
import UiToggle from '@/components/UiToggle.vue';
import UiDatepicker from '@/components/datepicker/UiDatepicker.vue';
import type {Component} from 'vue';
import UiCheckbox from "@/components/UiCheckbox.vue";
import UiRadioButton from "@/components/UiRadioButton.vue";
import UiHStack from "@/components/hstack/UiHStack.vue";
import UiVStack from "@/components/vstack/UiVStack.vue";
import UiBox from "@/components/box/UiBox.vue";
import UiSpacer from "@/components/spacer/UiSpacer.vue";
import UiModal from "@/components/UiModal.vue";
import UiWindowTitle from "@/components/UiWindowTitle.vue";
import UiTable from "@/components/table/UiTable.vue";
import UiPasswordField from "@/components/UiPasswordField.vue";
import UiScrollView from "@/components/scrollview/UiScrollView.vue";
import UiTextLayout from "@/components/textlayout/UiTextLayout.vue";

// Add new UI components to the following map
const uiComponentsMap: Map<string, Component> = new Map<string, Component>();
uiComponentsMap.set('A', UiScaffold);
// uiComponentsMap.set('Chip', UiChip);
// uiComponentsMap.set('Dialog', UiDialog);
uiComponentsMap.set('d', UiDivider);
// uiComponentsMap.set('Stepper', UiStepper);
// uiComponentsMap.set('FileField', UiUploadField);
uiComponentsMap.set('I', UiImage);
uiComponentsMap.set('F', UiTextField);
// uiComponentsMap.set('TextArea', UiTextArea);
uiComponentsMap.set('t', UiToggle);
uiComponentsMap.set('T', UiText);
// uiComponentsMap.set('Button', UiButton);
uiComponentsMap.set('G', UiGrid);
 uiComponentsMap.set('B', UiTable);
// uiComponentsMap.set('Card', UiCard);
// uiComponentsMap.set('Dropdown', UiDropdown);
uiComponentsMap.set('P', UiDatepicker);
// uiComponentsMap.set('Slider', UiSlider);
// uiComponentsMap.set('NumberField', UiNumberField);
// uiComponentsMap.set('WebView', UiWebView);
// uiComponentsMap.set('Page', UiPage);
 uiComponentsMap.set('p', UiPasswordField);
// uiComponentsMap.set('Breadcrumbs', UiBreadcrumbs);
uiComponentsMap.set('c', UiCheckbox);
uiComponentsMap.set('R', UiRadioButton);
// uiComponentsMap.set('FlexContainer', UiFlexContainer);
// uiComponentsMap.set('ProgressBar', UiProgressBar);
uiComponentsMap.set('hs', UiHStack)
uiComponentsMap.set('vs', UiVStack)
uiComponentsMap.set('bx', UiBox)
uiComponentsMap.set('s', UiSpacer)
uiComponentsMap.set('M', UiModal)
uiComponentsMap.set('W', UiWindowTitle)
uiComponentsMap.set('V', UiScrollView)
uiComponentsMap.set('ts',UiTextLayout)

export default uiComponentsMap;
