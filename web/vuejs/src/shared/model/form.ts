export interface Form {
	type: 'Form'
	submitText: string
	deleteText: string
	links: FormLinks
}

export interface FormLinks {
	load: URL | null
	submit: URL | null
	delete: URL | null
}
