// Code generated by "stringer -type Direction"; DO NOT EDIT

package log

import "fmt"

const _Direction_name = "DirectionSendDirectionRecv"

var _Direction_index = [...]uint8{0, 13, 26}

func (i Direction) String() string {
	if i < 0 || i >= Direction(len(_Direction_index)-1) {
		return fmt.Sprintf("Direction(%d)", i)
	}
	return _Direction_name[_Direction_index[i]:_Direction_index[i+1]]
}
