// Package returnbool covers return-bool: which signatures end in a bool after
// another result, and which declarations are checked.
package returnbool

type T struct{ N int }

// ===== Functions =====

func Lookup(m map[string]int, k string) (int, bool) { // want `^Lookup reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
	v, ok := m[k]
	return v, ok
}

func Pointer() (*T, bool) { // want `^Pointer reports absence with a trailing bool; return mo\.Option\[\*T\] instead \[return-bool\]$`
	return &T{}, true
}

func Slice() ([]string, bool) { // want `^Slice reports absence with a trailing bool; return mo\.Option\[\[\]string\] instead \[return-bool\]$`
	return nil, false
}

func Interface() (any, bool) { // want `^Interface reports absence with a trailing bool; return mo\.Option\[any\] instead \[return-bool\]$`
	return nil, false
}

// A bool before the trailing bool is a value like any other.
func BoolBool() (bool, bool) { // want `^BoolBool reports absence with a trailing bool; return mo\.Option\[bool\] instead \[return-bool\]$`
	return false, false
}

func Two() (string, int, bool) { // want `^Two reports absence with a trailing bool; return mo\.Option of a struct, or of lo\.Tuple2 instead \[return-bool\]$`
	return "", 0, false
}

func Three() (string, int, *T, bool) { // want `^Three reports absence with a trailing bool; return mo\.Option of a struct, or of lo\.Tuple3 instead \[return-bool\]$`
	return "", 0, nil, false
}

func Nine() (int, int, int, int, int, int, int, int, int, bool) { // want `^Nine reports absence with a trailing bool; return mo\.Option of a struct, or of lo\.Tuple9 instead \[return-bool\]$`
	return 0, 0, 0, 0, 0, 0, 0, 0, 0, false
}

// With more than nine values, lo has no tuple to suggest.
func Ten() (int, int, int, int, int, int, int, int, int, int, bool) { // want `^Ten reports absence with a trailing bool; return mo\.Option of a struct instead \[return-bool\]$`
	return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, false
}

// Aliases are resolved in the message.
type PT = *T

func AliasPT() (PT, bool) { // want `^AliasPT reports absence with a trailing bool; return mo\.Option\[\*T\] instead \[return-bool\]$`
	return &T{}, true
}

// The names of the results do not matter.
func Named() (v int, ok bool) { // want `^Named reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
	return
}

// The same with a trailing error before the bool.
func ErrorThenBool() (int, error, bool) { // want `^ErrorThenBool reports absence with a trailing bool; return mo\.Option of a struct, or of lo\.Tuple2 instead \[return-bool\]$`
	return 0, nil, false
}

// Not reported: nothing comes before the bool.
func Only() bool { return true }

// Not reported: the bool is not last.
func BoolError() (bool, error) { return false, nil }

// Not reported: the bool is not last.
func BoolFirst() (bool, int) { return false, 0 }

// Not reported: no result.
func NoResult() {}

// Not reported: only the predeclared bool counts.
type MyBool bool

func Mine() (int, MyBool) { return 0, false }

// An alias of bool is identical to bool.
type AliasBool = bool

func Aliased() (int, AliasBool) { // want `^Aliased reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
	return 0, false
}

// A function without a body is checked too.
func noBody() (int, bool) // want `^noBody reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`

// ===== Generics =====

func Get[K comparable, V any](m map[K]V, k K) (V, bool) { // want `^Get\[K, V\] reports absence with a trailing bool; return mo\.Option\[V\] instead \[return-bool\]$`
	v, ok := m[k]
	return v, ok
}

// ===== Methods =====

type S struct{}

func (S) Value() (int, bool) { // want `^S\.Value reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
	return 0, false
}

func (*S) Pointer() (int, bool) { // want `^\(\*S\)\.Pointer reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
	return 0, false
}

type Box[E any] struct{}

func (Box[E]) Get() (E, bool) { // want `^Box\[E\]\.Get reports absence with a trailing bool; return mo\.Option\[E\] instead \[return-bool\]$`
	var z E
	return z, false
}

// Not reported: a method with a single bool.
func (S) Ok() bool { return true }

// ===== Named interface types =====

type Finder interface {
	Find(string) (*T, bool) // want `^Finder\.Find reports absence with a trailing bool; return mo\.Option\[\*T\] instead \[return-bool\]$`
	Has(string) bool
}

type GenericFinder[E any] interface {
	Find(string) (E, bool) // want `^GenericFinder\[E\]\.Find reports absence with a trailing bool; return mo\.Option\[E\] instead \[return-bool\]$`
}

// Not reported: an embedded interface is reported where it is declared, not
// where it is embedded.
type Embeds interface {
	Finder
	Close()
}

// Not reported: an interface literal is not a named interface type.
var Anonymous interface {
	Find(string) (*T, bool)
}

// ===== Named function types =====

type FindFunc func(string) (*T, bool) // want `^FindFunc reports absence with a trailing bool; return mo\.Option\[\*T\] instead \[return-bool\]$`

type GenericFunc[E any] func() (E, bool) // want `^GenericFunc\[E\] reports absence with a trailing bool; return mo\.Option\[E\] instead \[return-bool\]$`

type (
	GroupedFunc func() (int, bool) // want `^GroupedFunc reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
)

// Not reported: a named function type that ends in bool alone.
type Pred func(int) bool

// Not reported: an alias is not a named function type.
type AliasFunc = func() (int, bool)

// Not reported: a function type in a parameter is not a declaration.
func Takes(f func() (int, bool)) { _ = f }

// Not reported: a function type in a field is not a declaration.
type Holder struct {
	F func() (int, bool)
}

// ===== Function literals =====

// Not reported: a function literal is exempt.
func Literal() {
	f := func() (int, bool) { return 0, false }
	_ = f
}

// Not reported: a function literal is exempt, also at package level.
var LiteralVar = func() (int, bool) { return 0, false }

// ===== Local declarations =====

func Local() {
	type localFunc func() (int, bool) // want `^localFunc reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
	var f localFunc
	_ = f
}

func LocalInterface() {
	type localIface interface {
		Find() (int, bool) // want `^localIface\.Find reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
	}
	var i localIface
	_ = i
}
