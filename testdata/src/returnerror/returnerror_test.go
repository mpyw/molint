package returnerror

// A test file is checked like any other file.
func fixture() (string, error) { // want `^fixture reports failure with a trailing error; return mo\.Result\[string\] instead \[return-error\]$`
	return "", nil
}
