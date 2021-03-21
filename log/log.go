package log

import (
	"fmt"

	"github.com/logrusorgru/aurora"
)

var Verbose bool = false

func SpecName(id, name string) {
	fmt.Printf("%s: %s\n", id, name)
}

func TestName(num int, name string) {
	fmt.Printf("%s\r", aurora.Gray(12, fmt.Sprintf("#%d: %s", num, name)))
}

func Pass(num int, name string) {
	fmt.Printf("%s\n", aurora.Green(fmt.Sprintf("#%d: %s", num, name)))
}

func Fail(num int, name, reason string) {
	fmt.Printf("%s\n", aurora.Red(fmt.Sprintf("#%d: %s", num, name)))
	fmt.Printf("%s\n", aurora.Red(fmt.Sprintf("Failed: %s", reason)))
}

func Skip(num int, name, reason string) {
	fmt.Printf("%s\n", aurora.Cyan(fmt.Sprintf("#%d: %s", num, name)))
	fmt.Printf("%s\n", aurora.Cyan(fmt.Sprintf("Skipped: %s", reason)))
}

func Error(num int, name string) {
	fmt.Printf("%s\n", aurora.Red(fmt.Sprintf("#%d: %s", num, name)))
}

func Debug(msg string) {
	if Verbose {
		fmt.Printf("%s\n", aurora.Gray(12, msg))
	}
}

func Summary(tc, pc, sc, fc int) {
	fmt.Printf("%d tests, %d passed, %d skipped, %d failed\n", tc, pc, sc, fc)
}

func BlankLine() {
	fmt.Print("\n")
}
