import type { LiveTextField } from '@/shared/model/liveTextField';
import type { VBox } from '@/shared/model/vBox';
import type { LiveTable } from '@/shared/model/liveTable';
import type { LiveDropdown } from '@/shared/model/liveDropdown';
import type { LiveDropdownItem } from '@/shared/model/liveDropdownItem';
import type { LiveButton } from '@/shared/model/liveButton';
import type { LiveTableCell } from '@/shared/model/liveTableCell';
import type { LiveTableRow } from '@/shared/model/liveTableRow';
import type { LiveGridCell } from '@/shared/model/liveGridCell';
import type { LiveGrid } from '@/shared/model/liveGrid';
import type { LiveDialog } from '@/shared/model/liveDialog';
import type { LiveToggle } from '@/shared/model/liveToggle';
import type { LiveStepper } from '@/shared/model/liveStepper';
import type { LiveStepInfo } from '@/shared/model/liveStepInfo';
import type { LiveTextArea } from '@/shared/model/liveTextArea';
import type { LiveChip } from '@/shared/model/liveChip';
import type { LivePage } from '@/shared/model/livePage';
import type { LiveImage } from '@/shared/model/liveImage';

export type LiveComponent =
	LiveTextField
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
	| LiveImage
