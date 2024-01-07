package support

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/logrusorgru/aurora"
)

const (
	MessageIndentation = "  "
)

func SkippedMessageWithSymbol(msg string) {
	fmt.Println(MessageIndentation + aurora.White(msg).String() + aurora.BrightYellow(" [skipped] ⚬").String())
}

func SkippedMessageWitReason(msg string, reason string) {
	fmt.Println(MessageIndentation + aurora.White(reason).String() + aurora.BrightYellow(" [skipped] ⚬").String())
}

func SuccessMessageWithCheck(msg string) {
	fmt.Println(MessageIndentation + aurora.White(msg).String() + aurora.BrightGreen(" ✓").String())
}

func FailureMessageWithXMark(msg string) {
	fmt.Println(MessageIndentation + aurora.White(msg).String() + aurora.BrightRed(" ✗").String())
}

func WarningMessage(msg string) {
	fmt.Println(MessageIndentation + aurora.BrightYellow(msg).String())
}

func StatusMessageLine(msg string, highlight bool) {
	var text = aurora.White(msg)
	if highlight {
		text = aurora.BrightYellow(msg)
	}
	fmt.Println(MessageIndentation + text.String())
}

func StatusMessage(msg string, highlight bool) {
	var text = aurora.White(msg)
	if highlight {
		text = aurora.BrightYellow(msg)
	}
	fmt.Print(MessageIndentation + text.String())
}

func PrintCheckMark() {
	fmt.Print(aurora.BrightGreen(" ✓").String())
}

func PrintCheckMarkLine() {
	fmt.Print(aurora.BrightGreen(" ✓\n").String())
}

func PrintXMarkLine() {
	fmt.Print(aurora.BrightRed(" ✗\n").String())
}

// The function `FindExistingFile` takes a list of filenames and a default filename, and returns the
// first existing filename in the list or the default filename if none of the filenames exist.
func FindExistingFile(filenames []string, defaultFilename string) string {
	for _, filename := range filenames {
		if _, err := os.Stat(filename); err == nil {
			return filename
		}
	}

	return defaultFilename
}

// The function `GetCommandOutput` takes a command as input, executes it, and returns the output as a
// string.
func GetCommandOutput(command string) string {
	parts := strings.Split(command, " ")

	cmd := exec.Command(parts[0], parts[1:]...)

	outputBytes, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(outputBytes))
}
