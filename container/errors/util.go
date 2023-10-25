package errors

func OrPanic(e error) {
	if e != nil {
		panic(e)
	}
}
