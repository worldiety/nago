import type { LiveButton } from '@/shared/model/liveButton';
import type { LiveChip } from '@/shared/model/liveChip';
import type { LiveDialog } from '@/shared/model/liveDialog';
import type { LiveDropdown } from '@/shared/model/liveDropdown';
import type { LiveDropdownItem } from '@/shared/model/liveDropdownItem';
import type { LiveGrid } from '@/shared/model/liveGrid';
import type { LiveGridCell } from '@/shared/model/liveGridCell';
import type { LiveImage } from '@/shared/model/liveImage';
import type { LivePage } from '@/shared/model/livePage';
import type { LiveStepInfo } from '@/shared/model/liveStepInfo';
import type { LiveStepper } from '@/shared/model/liveStepper';
import type { LiveTable } from '@/shared/model/liveTable';
import type { LiveTableCell } from '@/shared/model/liveTableCell';
import type { LiveTableRow } from '@/shared/model/liveTableRow';
import type { LiveTextArea } from '@/shared/model/liveTextArea';
import type { LiveTextField } from '@/shared/model/liveTextField';
import type { LiveToggle } from '@/shared/model/liveToggle';
import type { VBox } from '@/shared/model/vBox';

export type LiveComponent =
	| LiveTextField
	| VBox
	| LiveTable
	| LiveDropdown
	| LiveDropdownItem
	| LiveButton
	| LiveTableCell
	| LiveTableRow
	| LiveGridCell
	| LiveGrid
	| LiveDialog
	| LiveToggle
	| LiveStepper
	| LiveStepInfo
	| LiveTextArea
	| LiveChip
	| LivePage
	| LiveImage;
