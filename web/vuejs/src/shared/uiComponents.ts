import type { Component } from 'vue';
import UiScaffold from '@/components/UiScaffold.vue';
import UiVBox from '@/components/UiVBox.vue';
import UiHBox from '@/components/UiHBox.vue';
import UiChip from '@/components/UiChip.vue';
import UiDialog from '@/components/UiDialog.vue';
import UiDivider from '@/components/UiDivider.vue';
import UiStepper from '@/components/UiStepper.vue';
import UiUploadField from '@/components/UiUploadField.vue';
import UiImage from '@/components/UiImage.vue';
import UiTextField from '@/components/UiTextField.vue';
import UiTextArea from '@/components/UiTextArea.vue';
import UiToggle from '@/components/UiToggle.vue';
import UiText from '@/components/UiText.vue';
import UiButton from '@/components/UiButton.vue';
import UiGrid from '@/components/UiGrid.vue';
import UiTable from '@/components/UiTable.vue';
import UiCard from '@/components/UiCard.vue';
import UiDropdown from '@/components/UiDropdown.vue';
import UiDatepicker from '@/components/UiDatepicker.vue';

// Add new UI components to the following map
const uiComponentsMap: Map<string, Component> = new Map<string, Component>();
uiComponentsMap.set('Scaffold', UiScaffold);
uiComponentsMap.set('VBox', UiVBox);
uiComponentsMap.set('HBox', UiHBox);
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
uiComponentsMap.set('Datepicker', UiDatepicker);

export default uiComponentsMap;
