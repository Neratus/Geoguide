package domain

import (
	"fmt"
	"io"
)

type Long int64

func (l *Long) UnmarshalGQL(v interface{}) error {
	switch v := v.(type) {
	case int:
		*l = Long(v)
	case int64:
		*l = Long(v)
	case float64:
		*l = Long(v)
	case string:
		var n int64
		if _, err := fmt.Sscan(v, &n); err != nil {
			return fmt.Errorf("Long must be an integer")
		}
		*l = Long(n)
	default:
		return fmt.Errorf("Long must be an integer")
	}
	return nil
}

func (l Long) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, int64(l))
}
