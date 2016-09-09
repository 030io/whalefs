package kingpin_parser

import (
	"strconv"
	"gopkg.in/alecthomas/kingpin.v2"
	"fmt"
)

type size uint64

func (s *size) Set(value string) error {
	num, err := strconv.ParseUint(value[:len(value) - 1], 10, 64)
	if err != nil {
		return fmt.Errorf("can't parse \"%s\"", value)
	}

	switch value[len(value) - 1] {
	case 'B', 'b':
	case 'K', 'k':
		num = num << 10
	case 'M', 'm':
		num = num << 20
	case 'G', 'g':
		num = num << 30
	default:
		return fmt.Errorf("can't parse \"%s\"", value)
	}
	*s = size(num)
	return nil
}

func (s *size) String() string {
	return strconv.FormatUint(uint64(*s), 10)
}

func Size(s kingpin.Settings) (target *uint64) {
	target = new(uint64)
	s.SetValue((*size)(target))
	return
}
