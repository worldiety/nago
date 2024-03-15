export interface TimelineElement {
	type: 'Timeline';
	items: TimelineItem[];
}

export interface TimelineItem {
	type: 'TimelineItem'
	icon: Image
	color: string | null
	title: string
	alternateDotText: string | null
	target: string
}
