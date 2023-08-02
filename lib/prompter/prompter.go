package prompter

// borrowed from https://raw.githubusercontent.com/cli/cli/trunk/internal/prompter/prompter.go
import (
	"fmt"
	"io"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

type Prompter interface {
	Select(string, string, []string) (int, error)
	MultiSelect(string, []string, []string) ([]int, error)
	Input(string, string) (string, error)
	Password(string) (string, error)
	AuthToken() (string, error)
	Confirm(string, bool) (bool, error)
	ConfirmDeletion(string) error
}

type fileWriter interface {
	io.Writer
	Fd() uintptr
}

type fileReader interface {
	io.Reader
	Fd() uintptr
}

func New(editorCmd string, stdin fileReader, stdout fileWriter, stderr io.Writer) Prompter {
	return &surveyPrompter{
		editorCmd: editorCmd,
		stdin:     stdin,
		stdout:    stdout,
		stderr:    stderr,
	}
}

type surveyPrompter struct {
	editorCmd string
	stdin     fileReader
	stdout    fileWriter
	stderr    io.Writer
}

// LatinMatchingFilter returns whether the value matches the input filter.
// The strings are compared normalized in case.
// The filter's diactritics are kept as-is, but the value's are normalized,
// so that a missing diactritic in the filter still returns a result.
func LatinMatchingFilter(filter, value string, index int) bool {
	filter = strings.ToLower(filter)
	value = strings.ToLower(value)

	// include this option if it matches.
	return strings.Contains(value, filter) || strings.Contains(value, filter)
}

func (p *surveyPrompter) Select(message, defaultValue string, options []string) (result int, err error) {
	q := &survey.Select{
		Message:  message,
		Options:  options,
		PageSize: 20,
		Filter:   LatinMatchingFilter,
	}

	if defaultValue != "" {
		// in some situations, defaultValue ends up not being a valid option; do
		// not set default in that case as it will make survey panic
		for _, o := range options {
			if o == defaultValue {
				q.Default = defaultValue
				break
			}
		}
	}

	err = p.ask(q, &result)

	return
}

func (p *surveyPrompter) MultiSelect(message string, defaultValues, options []string) (result []int, err error) {
	q := &survey.MultiSelect{
		Message:  message,
		Options:  options,
		PageSize: 20,
		Filter:   LatinMatchingFilter,
	}

	if len(defaultValues) > 0 {
		// TODO I don't actually know that this is needed, just being extra cautious
		validatedDefault := []string{}
		for _, x := range defaultValues {
			for _, y := range options {
				if x == y {
					validatedDefault = append(validatedDefault, x)
				}
			}
		}
		q.Default = validatedDefault
	}

	err = p.ask(q, &result)

	return
}

func (p *surveyPrompter) ask(q survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	opts = append(opts, survey.WithStdio(p.stdin, p.stdout, p.stderr))
	err := survey.AskOne(q, response, opts...)
	if err == nil {
		return nil
	}
	return fmt.Errorf("could not prompt: %w", err)
}

func (p *surveyPrompter) Input(prompt, defaultValue string) (result string, err error) {
	err = p.ask(&survey.Input{
		Message: prompt,
		Default: defaultValue,
	}, &result)

	return
}

func (p *surveyPrompter) ConfirmDeletion(requiredValue string) error {
	var result string
	return p.ask(
		&survey.Input{
			Message: fmt.Sprintf("Type %s to confirm deletion:", requiredValue),
		},
		&result,
		survey.WithValidator(
			func(val interface{}) error {
				if str := val.(string); !strings.EqualFold(str, requiredValue) {
					return fmt.Errorf("You entered %s", str)
				}
				return nil
			}))
}

func (p *surveyPrompter) Password(prompt string) (result string, err error) {
	err = p.ask(&survey.Password{
		Message: prompt,
	}, &result)

	return
}

func (p *surveyPrompter) Confirm(prompt string, defaultValue bool) (result bool, err error) {
	err = p.ask(&survey.Confirm{
		Message: prompt,
		Default: defaultValue,
	}, &result)

	return
}

func (p *surveyPrompter) AuthToken() (result string, err error) {
	err = p.ask(&survey.Password{
		Message: "Paste your authentication token:",
	}, &result, survey.WithValidator(survey.Required))

	return
}
