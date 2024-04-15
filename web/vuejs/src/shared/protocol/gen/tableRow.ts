// DO NOT EDIT. Generated by ora-gen-ts

import type { Pointer } from '@/shared/protocol/pointer';
import type { Property } from '@/shared/protocol/property';
import type { TableCell } from '@/shared/protocol/gen/tableCell';


export interface TableRow {
    id: Pointer;
    type: 'TableRow';
    cells: Property<TableCell[]>;
    
}
