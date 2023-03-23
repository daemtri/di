package di

import (
	"strings"
)

type multiError []error

func (m multiError) Error() string {
	errString := make([]string, 0, len(m))
	for i := range m {
		errString = append(errString, m[i].Error())
	}
	return strings.Join(errString, ",")
}

func appendError(err1, err2 error) error {
	if err1 == nil {
		return err2
	}
	if me, ok := err1.(multiError); ok {
		return append(me, err2)
	}
	return multiError([]error{err1, err2})
}
