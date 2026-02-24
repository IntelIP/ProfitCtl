package cmd

import "errors"

// ExitError carries an explicit process exit code for command failures.
type ExitError struct {
	Code   int
	Err    error
	Silent bool
}

func (e *ExitError) Error() string {
	if e.Err == nil {
		return "command failed"
	}
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error {
	return e.Err
}

func wrapExit(code int, err error) error {
	if err == nil {
		return nil
	}
	return &ExitError{Code: code, Err: err}
}

func wrapExitSilent(code int, err error) error {
	if err == nil {
		return nil
	}
	return &ExitError{Code: code, Err: err, Silent: true}
}

// ExitCode resolves the best process exit code for an error.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var e *ExitError
	if errors.As(err, &e) {
		if e.Code <= 0 {
			return 1
		}
		return e.Code
	}
	return 1
}

// ShouldPrintError reports whether an error should be printed to stderr.
func ShouldPrintError(err error) bool {
	if err == nil {
		return false
	}
	var e *ExitError
	if errors.As(err, &e) {
		return !e.Silent
	}
	return true
}
