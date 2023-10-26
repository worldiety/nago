package serrors

func OrPanic(e error) {
	if e != nil {
		panic(e)
	}
}
