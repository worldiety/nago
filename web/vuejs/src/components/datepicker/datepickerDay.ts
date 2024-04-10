export default interface DatepickerDay {
	dayOfWeek: number;
	dayOfMonth: number;
	monthIndex: number;
	year: number;
	selectedStart: boolean;
	selectedEnd: boolean;
	withinRange: boolean;
}
