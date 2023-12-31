package app

import (
	"errors"
	"fmt"
)

// ENUM(not verified, pending, verified, mismatch, error)
type ChecksumVerificationState int

const (
	// ChecksumVerificationStateNotVerified is a ChecksumVerificationState of type Not Verified.
	ChecksumVerificationStateNotVerified ChecksumVerificationState = iota
	// ChecksumVerificationStatePending is a ChecksumVerificationState of type Pending.
	ChecksumVerificationStatePending
	// ChecksumVerificationStateVerified is a ChecksumVerificationState of type Verified.
	ChecksumVerificationStateVerified
	// ChecksumVerificationStateMismatch is a ChecksumVerificationState of type Mismatch.
	ChecksumVerificationStateMismatch
	// ChecksumVerificationStateError is a ChecksumVerificationState of type Error.
	ChecksumVerificationStateError
)

type ChecksumVerificationStates []ChecksumVerificationState

var AllFinalCHecksumVerificationStates = ChecksumVerificationStates{
	ChecksumVerificationStateVerified,
	ChecksumVerificationStateMismatch,
	ChecksumVerificationStateError,
}

var NonErrorFinalChecksumVerificationStates = ChecksumVerificationStates{
	ChecksumVerificationStateVerified,
	ChecksumVerificationStateMismatch,
}

var ErrInvalidChecksumVerificationState = errors.New("not a valid ChecksumVerificationState")

const _ChecksumVerificationStateName = "not verifiedpendingverifiedmismatcherror"

var _ChecksumVerificationStateMap = map[ChecksumVerificationState]string{
	ChecksumVerificationStateNotVerified: _ChecksumVerificationStateName[0:12],
	ChecksumVerificationStatePending:     _ChecksumVerificationStateName[12:19],
	ChecksumVerificationStateVerified:    _ChecksumVerificationStateName[19:27],
	ChecksumVerificationStateMismatch:    _ChecksumVerificationStateName[27:35],
	ChecksumVerificationStateError:       _ChecksumVerificationStateName[35:40],
}

// String implements the Stringer interface.
func (x ChecksumVerificationState) String() string {
	if str, ok := _ChecksumVerificationStateMap[x]; ok {
		return str
	}
	return fmt.Sprintf("ChecksumVerificationState(%d)", x)
}

func (x ChecksumVerificationState) IsVerified() bool {
	return x == ChecksumVerificationStateVerified
}

func (x ChecksumVerificationState) IsPending() bool {
	return x == ChecksumVerificationStatePending
}

func (x ChecksumVerificationState) IsMismatch() bool {
	return x == ChecksumVerificationStateMismatch
}

func (x ChecksumVerificationState) IsError() bool {
	return x == ChecksumVerificationStateError
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x ChecksumVerificationState) IsValid() bool {
	_, ok := _ChecksumVerificationStateMap[x]
	return ok
}

var _ChecksumVerificationStateValue = map[string]ChecksumVerificationState{
	_ChecksumVerificationStateName[0:12]:  ChecksumVerificationStateNotVerified,
	_ChecksumVerificationStateName[12:19]: ChecksumVerificationStatePending,
	_ChecksumVerificationStateName[19:27]: ChecksumVerificationStateVerified,
	_ChecksumVerificationStateName[27:35]: ChecksumVerificationStateMismatch,
	_ChecksumVerificationStateName[35:40]: ChecksumVerificationStateError,
}

var _ChecksumVerificationStateTransitionMap = map[ChecksumVerificationState]ChecksumVerificationStates{
	ChecksumVerificationStateNotVerified: {ChecksumVerificationStatePending},
	ChecksumVerificationStatePending:     AllFinalCHecksumVerificationStates, // {ChecksumVerificationStateVerified, ChecksumVerificationStateMismatch, ChecksumVerificationStateError},
	ChecksumVerificationStateVerified:    {},
	ChecksumVerificationStateMismatch:    {},
	ChecksumVerificationStateError:       {},
}

// ParseChecksumVerificationState attempts to convert a string to a ChecksumVerificationState.
func ParseChecksumVerificationState(name string) (ChecksumVerificationState, error) {
	if x, ok := _ChecksumVerificationStateValue[name]; ok {
		return x, nil
	}
	return ChecksumVerificationState(0), fmt.Errorf("%s is %w", name, ErrInvalidChecksumVerificationState)
}

func (x *ChecksumVerificationState) IsInFinalState() bool {
	for _, finalState := range AllFinalCHecksumVerificationStates {
		if *x == finalState {
			return true
		}
	}

	return false
}

func (x *ChecksumVerificationState) TransitionToNext(err error, matched bool) *ChecksumVerificationState {
	possibleStates := _ChecksumVerificationStateTransitionMap[*x]

	if len(possibleStates) == 0 {
		return x
	}

	if len(possibleStates) == 1 {
		*x = possibleStates[0]
		return x
	}

	for _, state := range possibleStates {
		if state == ChecksumVerificationStateError {
			*x = ChecksumVerificationStateError
			return x
		}
	}

	for _, state := range possibleStates {
		for _, finalState := range AllFinalCHecksumVerificationStates {
			if state == finalState {
				x.SetVerified(matched)
				return x
			}
		}
	}

	return x
}

// MarshalText implements the text marshaller method.
func (x ChecksumVerificationState) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *ChecksumVerificationState) UnmarshalText(text []byte) error {
	name := string(text)
	tmp, err := ParseChecksumVerificationState(name)
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}

func (x *ChecksumVerificationState) SetVerified(value bool) {
	value = value && !x.IsError()

	if value {
		*x = ChecksumVerificationStateVerified
		return
	}

	*x = ChecksumVerificationStateMismatch
}

func (x *ChecksumVerificationState) Reset() {
	*x = ChecksumVerificationStateNotVerified
}

func (x *ChecksumVerificationState) ResetIf(value bool) {
	if value {
		x.Reset()
	}
}
