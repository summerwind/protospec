package spec

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"

	"github.com/summerwind/protospec/config"
	"github.com/summerwind/protospec/task"
)

type CLIRunner struct {
	c *config.Config
}

func NewCLIRunner(c *config.Config) *CLIRunner {
	return &CLIRunner{c: c}
}

func (r *CLIRunner) Run() error {
	var (
		passed  int
		skipped int
		failed  int
	)

	specs, err := Load(r.c.Specs)
	if err != nil {
		return err
	}

	start := time.Now()
	for _, spec := range specs {
		fmt.Printf("%s\n", spec.Name)

		for i, test := range spec.Tests {
			id := fmt.Sprintf("%s/%d", spec.ID, i+1)
			printRunning(id, test.Name)

			err := runTest(test, r.c)
			if err != nil {
				var (
					f *task.Failed
					s *task.Skipped
				)

				switch {
				case errors.As(err, &f):
					printFailed(id, test, *f)
					failed += 1
				case errors.As(err, &s):
					printSkipped(id, test, *s)
					skipped += 1
				default:
					return err
				}
			} else {
				printPassed(id, test.Name)
				passed += 1
			}
		}

		fmt.Println("")
	}
	end := time.Now()

	fmt.Printf("Finished in %.4f seconds\n", end.Sub(start).Seconds())
	fmt.Printf("%d tests, %d passed, %d failed, %d skipped\n", (passed + failed + skipped), passed, failed, skipped)

	return nil
}

func runTest(t Test, c *config.Config) error {
	var (
		conn task.Conn
		err  error
	)

	for name, opts := range t.Connection {
		conn, err = task.NewConnection(name, []byte(opts))
		if err != nil {
			return err
		}
	}

	if conn == nil {
		// TODO: use default connection.
		return errors.New("no connection specified")
	}

	if c.Verbose {
		conn.HandleDebug(func(str string) {
			printDebug(str)
		})
	}

	err = conn.Connect(c)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx := task.SetConnection(context.Background(), conn)

	for _, step := range t.Steps {
		for name, opts := range step {
			task, err := task.NewTask(name, c, []byte(opts))
			if err != nil {
				return err
			}

			err = task.Run(ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func printRunning(id, str string) {
	fmt.Printf("%s\r", aurora.Gray(12, fmt.Sprintf("%s: %s", id, str)))
}

func printPassed(id, str string) {
	fmt.Println(aurora.Green(fmt.Sprintf("%s: %s", id, str)))
}

func printFailed(id string, t Test, f task.Failed) {
	fmt.Println(aurora.Red(fmt.Sprintf("%s: %s", id, t.Name)))
	fmt.Println(aurora.Red(fmt.Sprintf("  %s", t.Requirement)))
	fmt.Println(aurora.Red(fmt.Sprintf("  Expect: %s", strings.Join(f.Expect, "\n         "))))
	fmt.Println(aurora.Red(fmt.Sprintf("  Actual: %s", f.Actual)))
}

func printSkipped(id string, t Test, s task.Skipped) {
	fmt.Println(aurora.Cyan(fmt.Sprintf("%s: %s", id, t.Name)))
	fmt.Println(aurora.Cyan(fmt.Sprintf("  Skipped: %s", s.Reason)))
}

func printError(id string, t Test, err error) {
	fmt.Println(aurora.Red(fmt.Sprintf("%s: %s", id, t.Name)))
	fmt.Println(aurora.Red(fmt.Sprintf("  Error: %s", err)))
}

func printDebug(str string) {
	fmt.Println(aurora.Gray(12, str))
}
