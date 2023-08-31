package app

import (
	"errors"
	"fmt"
	"strings"

	"github.com/stackup-app/stackup/lib/utils"
)

// ENUM(http, s3, file)
type IncludeType int

const (
	// IncludeTypeHttp is a IncludeType of type Http.
	IncludeTypeHttp IncludeType = iota
	// IncludeTypeS3 is a IncludeType of type S3.
	IncludeTypeS3
	// IncludeTypeFile is a IncludeType of type File.
	IncludeTypeFile
	IncludeTypeUnknown
)

var ErrInvalidIncludeType = errors.New("not a valid IncludeType")

const _IncludeTypeName = "https3fileunknown"

var _IncludeTypeMap = map[IncludeType]string{
	IncludeTypeHttp:    _IncludeTypeName[0:4],
	IncludeTypeS3:      _IncludeTypeName[4:6],
	IncludeTypeFile:    _IncludeTypeName[6:10],
	IncludeTypeUnknown: _IncludeTypeName[10:17],
}

func DetermineIncludeType(strs ...string) IncludeType {
	for _, str := range strs {
		if strings.HasPrefix(str, "http") {
			return IncludeTypeHttp
		}

		if strings.HasPrefix(str, "s3:") {
			return IncludeTypeS3
		}

		if utils.IsFile(str) || len(str) > 0 {
			return IncludeTypeFile
		}
	}

	return IncludeTypeUnknown
}

// String implements the Stringer interface.
func (x IncludeType) String() string {
	if str, ok := _IncludeTypeMap[x]; ok {
		return str
	}
	return fmt.Sprintf("IncludeType(%d)", x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x IncludeType) IsValid() bool {
	_, ok := _IncludeTypeMap[x]
	return ok
}

var _IncludeTypeValue = map[string]IncludeType{
	_IncludeTypeName[0:4]:   IncludeTypeHttp,
	_IncludeTypeName[4:6]:   IncludeTypeS3,
	_IncludeTypeName[6:10]:  IncludeTypeFile,
	_IncludeTypeName[10:17]: IncludeTypeUnknown,
}

// ParseIncludeType attempts to convert a string to a IncludeType.
func ParseIncludeType(name string) (IncludeType, error) {
	if x, ok := _IncludeTypeValue[name]; ok {
		return x, nil
	}
	return IncludeType(0), fmt.Errorf("%s is %w", name, ErrInvalidIncludeType)
}

// MarshalText implements the text marshaller method.
func (x IncludeType) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *IncludeType) UnmarshalText(text []byte) error {
	name := string(text)
	tmp, err := ParseIncludeType(name)
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}
