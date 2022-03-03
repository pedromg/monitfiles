package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type testCases struct {
	ID   int
	Code int
	Res  exitError
}

var cases = []testCases{
	{1, -1, exitError{errUnknown, 1, "unknown error."}},
	{2, 0, exitError{errOK, 0, "Exited, bye."}},
	{3, 1, exitError{errNoParams, 1, "Unspecified params: "}},
	{4, 2, exitError{errParh, 1, "error in path: "}},
	{5, 3, exitError{errFileTypes, 1, "error in file types: "}},
	{6, 4, exitError{errScript, 1, "error in script: "}},
	{7, 5, exitError{errStorage, 1, "error building storage"}},
	{8, 100, exitError{errUnknown, 1, "unknown error."}},
	{9, -1, exitError{errUnknown, 1, "unknown error."}},
	{10, -2, exitError{errContinue, 0, ""}},
}

func TestExitError(t *testing.T) {

	for _, c := range cases {
		if r := cmp.Equal(exitErrorFor(c.Code), c.Res); !r {
			t.Errorf("case %d diffs: \ngot: %v\nshould be: %v", c.ID, exitErrorFor(c.Code), c.Res)
		}
	}
}

func TestExitErrorCode(t *testing.T) {
	for _, c := range cases {
		if exitErrorCode(c.Code) != c.Res.Exit {
			t.Errorf("case %d failed, got [%d] should be [%d] ", c.ID, exitErrors[c.Code].Exit, c.Res.Exit)
		}
	}

}

func TestExitErrorInfo(t *testing.T) {
	for _, c := range cases {
		if exitErrorInfo(c.Code) != c.Res.Description {
			t.Errorf("case %d failed, got [%s] should be [%s]", c.ID, exitErrors[c.Code].Description, c.Res.Description)
		}
	}

}
