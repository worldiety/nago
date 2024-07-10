import UiButton from '@/components/UiButton.vue';
import UiCard from '@/components/UiCard.vue';
import UiChip from '@/components/UiChip.vue';
import UiDialog from '@/components/UiDialog.vue';
import UiDivider from '@/components/UiDivider.vue';
import UiDropdown from '@/components/dropdown/UiDropdown.vue';
import UiGrid from '@/components/UiGrid.vue';
import UiImage from '@/components/UiImage.vue';
import UiScaffold from '@/components/scaffold/UiScaffold.vue';
import UiStepper from '@/components/UiStepper.vue';
import UiTable from '@/components/UiTable.vue';
import UiText from '@/components/UiText.vue';
import UiTextArea from '@/components/UiTextArea.vue';
import UiTextField from '@/components/UiTextField.vue';
import UiToggle from '@/components/UiToggle.vue';
import UiUploadField from '@/components/uploadfield/UiUploadField.vue';
import UiDatepicker from '@/components/datepicker/UiDatepicker.vue';
import UiSlider from '@/components/UiSlider.vue';
import UiNumberField from '@/components/UiNumberField.vue';
import type {Component} from 'vue';
import UiWebView from "@/components/UiWebView.vue";
import UiPage from "@/components/UiPage.vue";
import UiPasswordField from '@/components/UiPasswordField.vue';
import UiBreadcrumbs from '@/components/breadcrumbs/UiBreadcrumbs.vue';
import UiCheckbox from "@/components/UiCheckbox.vue";
import UiRadioButton from "@/components/UiRadioButton.vue";
import UiFlexContainer from '@/components/UiFlexContainer.vue';
import UiProgressBar from '@/components/UiProgressBar.vue';
import UiStr from "@/components/UiStr.vue";
import UiHStack from "@/components/hstack/UiHStack.vue";
import UiVStack from "@/components/vstack/UiVStack.vue";
import UiBox from "@/components/box/UiBox.vue";

// Add new UI components to the following map
const uiComponentsMap: Map<string, Component> = new Map<string, Component>();
uiComponentsMap.set('Scaffold', UiScaffold);
uiComponentsMap.set('Chip', UiChip);
uiComponentsMap.set('Dialog', UiDialog);
uiComponentsMap.set('Divider', UiDivider);
uiComponentsMap.set('Stepper', UiStepper);
uiComponentsMap.set('FileField', UiUploadField);
uiComponentsMap.set('Image', UiImage);
uiComponentsMap.set('TextField', UiTextField);
uiComponentsMap.set('TextArea', UiTextArea);
uiComponentsMap.set('Toggle', UiToggle);
uiComponentsMap.set('Text', UiText);
uiComponentsMap.set('Button', UiButton);
uiComponentsMap.set('Grid', UiGrid);
uiComponentsMap.set('Table', UiTable);
uiComponentsMap.set('Card', UiCard);
uiComponentsMap.set('Dropdown', UiDropdown);
uiComponentsMap.set('DatePicker', UiDatepicker);
uiComponentsMap.set('Slider', UiSlider);
uiComponentsMap.set('NumberField', UiNumberField);
uiComponentsMap.set('WebView', UiWebView);
uiComponentsMap.set('Page', UiPage);
uiComponentsMap.set('PasswordField', UiPasswordField);
uiComponentsMap.set('Breadcrumbs', UiBreadcrumbs);
uiComponentsMap.set('Checkbox', UiCheckbox);
uiComponentsMap.set('Radiobutton', UiRadioButton);
uiComponentsMap.set('FlexContainer', UiFlexContainer);
uiComponentsMap.set('ProgressBar', UiProgressBar);
uiComponentsMap.set('S', UiStr);
uiComponentsMap.set('hs', UiHStack)
uiComponentsMap.set('vs', UiVStack)
uiComponentsMap.set('bx', UiBox)

export default uiComponentsMap;
