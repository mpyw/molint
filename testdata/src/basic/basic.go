package basic

type User struct {
	Name   string
	Friend *User
}

func Literal() *User { // want Literal:"nonnil r0"
	return &User{}
}

func New() *User { // want New:"nonnil r0"
	return new(User)
}

func Nil() *User {
	return nil // want `^Nil may return a nil \*User, from a nil constant; return mo.Option\[\*User\] where absence is expected$`
}

func Zero() *User {
	var u *User
	return u // want `from a nil constant`
}

func Named() (u *User) {
	return // want `from a nil constant`
}

func NamedSet() (u *User) { // want NamedSet:"nonnil r0"
	u = &User{}
	return
}

func Param(u *User) *User {
	return u // want `from parameter u`
}

func Checked(u *User) *User { // want Checked:"nonnil r0"
	if u == nil {
		return &User{}
	}
	return u
}

func CheckedNeq(u *User) *User { // want CheckedNeq:"nonnil r0"
	if u != nil {
		return u
	}
	return New()
}

func CheckedNil(u *User) *User {
	if u == nil {
		return u // want `from a value checked to be nil`
	}
	return u
}

func Field(u User) *User {
	return u.Friend // want `from field Friend`
}

func FieldPtr(u *User) *User {
	return u.Friend // want `from field Friend`
}

func FieldAddr(u *User) *string { // want FieldAddr:"nonnil r0"
	return &u.Name
}

func Map(m map[string]*User) *User {
	return m["a"] // want `from a map lookup`
}

func MapOK(m map[string]*User) *User {
	u, ok := m["a"]
	if !ok {
		return &User{}
	}
	return u // want `from a map lookup`
}

func Assert(v any) *User {
	return v.(*User) // want `from a type assertion`
}

func Phi(b bool) *User { // want Phi:"nonnil r0"
	u := Literal()
	if b {
		u = New()
	}
	return u
}

func PhiNil(b bool) *User {
	u := Literal()
	if b {
		u = nil
	}
	return u // want `from a nil constant`
}

func Fallback(u *User) *User { // want Fallback:"nonnil r0"
	if u == nil {
		u = &User{}
	}
	return u
}

func Loop(n int) *User { // want Loop:"nonnil r0"
	u := &User{}
	for range n {
		u = &User{Friend: u}
	}
	return u
}

func CallsNil() *User {
	return Nil() // want `from a call to Nil, which may return nil`
}

func Closure() *User {
	f := func() *User {
		return nil // want `^the function literal in Closure may return a nil \*User, from a nil constant`
	}
	return f() // want `from a call to Closure\$1, which may return nil`
}

func ClosureOK() *User { // want ClosureOK:"nonnil r0"
	f := func() *User { return &User{} }
	return f()
}

func Captured(u *User) func() *User {
	return func() *User {
		return u // want `from captured variable u`
	}
}

func FuncValue(f func() *User) *User {
	return f() // want `from a call through an interface or a function value`
}

type Named2 *User

func NamedPointer() Named2 {
	return nil // want `may return a nil Named2`
}

func Two() (*User, *User) { // want Two:"nonnil r0"
	return &User{}, nil // want `^Two may return a nil \*User as result 1`
}

func (u *User) Self() *User {
	return u // want `^\(\*User\).Self may return a nil \*User, from parameter u`
}

func (u *User) SelfChecked() *User { // want SelfChecked:"nonnil r0"
	if u == nil {
		return &User{}
	}
	return u
}

func MethodValue(u *User) *User { // want MethodValue:"nonnil r0"
	f := u.SelfChecked
	return f()
}
