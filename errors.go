package main

// errorList is a small struct of exit errors
type exitError struct {
	ID          int
	Exit        int
	Description string
}

var (
	errUnknown   = -1
	errOK        = 0 // no error
	errNoParams  = 1
	errParh      = 2
	errFileTypes = 3
	errScript    = 4
	errStorage   = 5

	exitErrors = map[int]exitError{
		errUnknown:   {errUnknown, 1, "unknown error."},
		errOK:        {errOK, 0, "Exited, bye."},
		errNoParams:  {errNoParams, 1, "Unspecified params: "},
		errParh:      {errParh, 1, "error in path: "},
		errFileTypes: {errFileTypes, 1, "error in file types: "},
		errScript:    {errScript, 1, "error in script: "},
		errStorage:   {errStorage, 1, "error building storage"},
	}
)

func exitErrorFor(id int) exitError {
	e, ok := exitErrors[id]
	if !ok {
		return exitErrors[errUnknown]
	}
	return e
}

func exitErrorCode(id int) int {
	e, ok := exitErrors[id]
	if !ok {
		return exitErrors[errUnknown].Exit
	}
	return e.Exit
}

func exitErrorInfo(id int) string {
	e, ok := exitErrors[id]
	if !ok {
		return exitErrors[errUnknown].Description
	}
	return e.Description
}
