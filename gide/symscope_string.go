// Code generated by "stringer -type=SymScope"; DO NOT EDIT.

package gide

import (
	"errors"
	"strconv"
)

var _ = errors.New("dummy error")

const _SymScope_name = "SymScopeFileSymScopePackageSymScopeN"

var _SymScope_index = [...]uint8{0, 12, 27, 36}

func (i SymScope) String() string {
	if i < 0 || i >= SymScope(len(_SymScope_index)-1) {
		return "SymScope(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _SymScope_name[_SymScope_index[i]:_SymScope_index[i+1]]
}

func (i *SymScope) FromString(s string) error {
	for j := 0; j < len(_SymScope_index)-1; j++ {
		if s == _SymScope_name[_SymScope_index[j]:_SymScope_index[j+1]] {
			*i = SymScope(j)
			return nil
		}
	}
	return errors.New("String: " + s + " is not a valid option for type: SymScope")
}
