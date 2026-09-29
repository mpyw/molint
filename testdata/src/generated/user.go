package generated

func Use() *T {
	return Gen() // want `from a call to Gen, which may return nil`
}
