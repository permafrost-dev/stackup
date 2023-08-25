package checksums

import (
	"errors"
	"fmt"
)

// ENUM(sha1, sha256, sha512, unsupported, error)
type ChecksumAlgorithm int

const (
	// ChecksumAlgorithmSha1 is a ChecksumAlgorithm of type Sha1.
	ChecksumAlgorithmSha1 ChecksumAlgorithm = iota
	// ChecksumAlgorithmSha256 is a ChecksumAlgorithm of type Sha256.
	ChecksumAlgorithmSha256
	// ChecksumAlgorithmSha512 is a ChecksumAlgorithm of type Sha512.
	ChecksumAlgorithmSha512
	// ChecksumAlgorithmUnsupported is a ChecksumAlgorithm of type Unsupported.
	ChecksumAlgorithmUnsupported
	// ChecksumAlgorithmError is a ChecksumAlgorithm of type Error.
	ChecksumAlgorithmError
)

var ErrInvalidChecksumAlgorithm = errors.New("not a valid ChecksumAlgorithm")

const _ChecksumAlgorithmName = "sha1sha256sha512unsupportederror"

var _ChecksumAlgorithmMap = map[ChecksumAlgorithm]string{
	ChecksumAlgorithmSha1:        _ChecksumAlgorithmName[0:4],
	ChecksumAlgorithmSha256:      _ChecksumAlgorithmName[4:10],
	ChecksumAlgorithmSha512:      _ChecksumAlgorithmName[10:16],
	ChecksumAlgorithmUnsupported: _ChecksumAlgorithmName[16:27],
	ChecksumAlgorithmError:       _ChecksumAlgorithmName[27:32],
}

// String implements the Stringer interface.
func (x ChecksumAlgorithm) String() string {
	if str, ok := _ChecksumAlgorithmMap[x]; ok {
		return str
	}
	return fmt.Sprintf("ChecksumAlgorithm(%d)", x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x ChecksumAlgorithm) IsValid() bool {
	_, ok := _ChecksumAlgorithmMap[x]
	return ok
}

func (x ChecksumAlgorithm) IsSupportedAlgorithm() bool {
	return x == ChecksumAlgorithmSha256 || x == ChecksumAlgorithmSha512
}

var _ChecksumAlgorithmValue = map[string]ChecksumAlgorithm{
	_ChecksumAlgorithmName[0:4]:   ChecksumAlgorithmSha1,
	_ChecksumAlgorithmName[4:10]:  ChecksumAlgorithmSha256,
	_ChecksumAlgorithmName[10:16]: ChecksumAlgorithmSha512,
	_ChecksumAlgorithmName[16:27]: ChecksumAlgorithmUnsupported,
	_ChecksumAlgorithmName[27:32]: ChecksumAlgorithmError,
}

// ParseChecksumAlgorithm attempts to convert a string to a ChecksumAlgorithm.
func ParseChecksumAlgorithm(name string) ChecksumAlgorithm {
	if x, ok := _ChecksumAlgorithmValue[name]; ok {
		return x
	}

	return ChecksumAlgorithmUnsupported
	// return ChecksumAlgorithm(0), fmt.Errorf("%s is %w", name, ErrInvalidChecksumAlgorithm)
}

// MarshalText implements the text marshaller method.
func (x ChecksumAlgorithm) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *ChecksumAlgorithm) UnmarshalText(text []byte) error {
	name := string(text)
	*x = ParseChecksumAlgorithm(name)

	return nil
}
