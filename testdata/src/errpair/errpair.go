package errpair

import (
	"errors"
	"strconv"
)

type User struct{ ID int }

var ErrNotFound = errors.New("not found")

func Find(id int) (*User, error) { // want Find:"nonnil r0"
	if id == 0 {
		return nil, ErrNotFound
	}
	return &User{ID: id}, nil
}

func NilNil() (*User, error) {
	return nil, nil // want `^NilNil may return a nil \*User with a nil error, from a nil constant; return an error, or mo.Option\[\*User\], where absence is expected$`
}

func NilNilVar() (*User, error) {
	var err error
	return nil, err // want `with a nil error, from a nil constant`
}

func Wrapped(id int) (*User, error) { // want Wrapped:"nonnil r0"
	u, err := Find(id)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func WrappedEq(id int) (*User, error) { // want WrappedEq:"nonnil r0"
	u, err := Find(id)
	if err == nil {
		return u, nil
	}
	return nil, err
}

func Unchecked(id int) (*User, error) {
	u, _ := Find(id)
	return u, nil // want `from a call to Find, whose error is not checked for nil`
}

func Passed(id int) (*User, error) { // want Passed:"nonnil r0"
	return Find(id)
}

func PassedVars(id int) (*User, error) { // want PassedVars:"nonnil r0"
	u, err := Find(id)
	return u, err
}

func PassedBad() (*User, error) {
	u, err := NilNil()
	return u, err // want `from a call to NilNil, which may return nil`
}

func CheckedBad() (*User, error) {
	u, err := NilNil()
	if err != nil {
		return nil, err
	}
	return u, nil // want `from a call to NilNil, which may return nil`
}

func Parse(s string) (*User, error) { // want Parse:"nonnil r0"
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return &User{ID: n}, nil
}

func Phi(b bool) (*User, error) { // want Phi:"nonnil r0"
	var u *User
	var err error
	if b {
		u = &User{}
	} else {
		err = ErrNotFound
	}
	return u, err
}

func PhiBad(b bool) (*User, error) {
	var u *User
	var err error
	if b {
		err = ErrNotFound
	}
	return u, err // want `with a nil error, from a nil constant`
}

func NestedPhi(a, b bool) (*User, error) {
	var err error
	if a {
		err = ErrNotFound
	}
	if b {
		return &User{}, nil
	}
	return nil, err // want `with a nil error, from a nil constant`
}

type Repo interface {
	Find(id int) (*User, error)
	Get(id int) *User
}

func ViaInterface(r Repo, id int) (*User, error) { // want ViaInterface:"nonnil r0"
	u, err := r.Find(id)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func ViaInterfaceUnchecked(r Repo, id int) (*User, error) {
	u, _ := r.Find(id)
	return u, nil // want `from a call through an interface or a function value, whose error is not checked for nil`
}

func ViaInterfaceBare(r Repo, id int) *User {
	return r.Get(id) // want `from a call through an interface or a function value$|from a call through an interface or a function value;`
}

func Recursive(n int) (*User, error) { // want Recursive:"nonnil r0"
	if n == 0 {
		return &User{}, nil
	}
	return Recursive(n - 1)
}

func Even(n int) *User { // want Even:"nonnil r0"
	if n == 0 {
		return &User{}
	}
	return Odd(n - 1)
}

func Odd(n int) *User { // want Odd:"nonnil r0"
	if n == 0 {
		return &User{}
	}
	return Even(n - 1)
}

// Spin never returns, so whatever it would return is vacuously proven.
func Spin(n int) *User { // want Spin:"nonnil r0"
	return Spin(n)
}
