package support

import (
	"fmt"

	"github.com/logrusorgru/aurora"
)

const (
	MessageIndentation = "  "
)

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
