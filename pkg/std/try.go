package std

// Try is best used with defer and an according function pointer, which is evaluated when the defer runs the try.
func Try(f func() error, err *error) {
	newErr := f()
	if *err == nil {
		*err = newErr
	}
}
