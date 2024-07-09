export enum NamedColor {
	// Primary call-to-action intention.
	Primary = "p",

	// Secondary call-to-action intention.
	Secondary = "s",

	// Tertiary call-to-action intention.
	Tertiary = "t",

	// Error describes a negative or a destructive intention. In Western Europe usually red. Use it, when the
	// user cannot continue normally and has to fix the problem first.
	Error = "n",

	// Warning describes a critical condition. In Western Europe usually yellow. Use it to warn on situations,
	// which may result in a future error condition.
	Warning = "c",

	// Positive describes a good condition or a confirming intention. In Western Europe usually green.
	// Use it to symbolize something which has been successfully applied.
	Positive = "o",

	// Informative shall be used to highlight something, which just changed. E.g. a newly added component or
	// a recommendation from a system. Do not use it to highlight text. In Western Europe usually blue.
	Informative = "i",

	// Regular shall be used for any default of any UI element which has no special semantic intention.
	// An empty color is always regular.
	Regular = "r"
}


export function namedColorStyles(attr:string,v? :string):string{
	if (!v){
		return ""
	}

	if (v.startsWith("#")){
		return `${attr}: ${v};`
	}

	return ""
}

export function namedColorClasses(v? :string):string{
	if (!v || v===""){
		return "regular"
	}

	if (v.startsWith("#")){
		return ""
	}

	switch (v){
		case NamedColor.Primary:
			return "primary"
		case NamedColor.Secondary:
			return "secondary"
		case NamedColor.Tertiary:
			return "tertiary"
		case NamedColor.Regular:
			return "regular"
		case NamedColor.Positive:
			return "positive"
		case NamedColor.Error:
			return "error"
		case NamedColor.Warning:
			return "warning"
		case NamedColor.Informative:
			return "informative"
		default:
			console.log("unknown color class name",v)
			return ""

	}
}
