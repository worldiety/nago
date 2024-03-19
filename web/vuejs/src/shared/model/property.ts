import type { PropertyBool } from '@/shared/model/propertyBool';
import type { PropertyInt } from '@/shared/model/propertyInt';
import type { PropertyString } from '@/shared/model/propertyString';

export type Property = PropertyString | PropertyBool | PropertyInt;
