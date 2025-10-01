package logger

type (
	AppLog interface {
		simple
		withError
		printfStyle
		printfStyleWithError
	}

	simple interface {
		Info(msg string)

		Debug(msg string)

		Warn(msg string)
	}

	// withError will add error param to the end of msg string
	// something like {... message="my info message: err=\"error_text\"" ...}
	// if err ==  nil {... message="my info message" ...}
	withError interface {
		InfoWithError(msg string, err error)

		DebugWithError(msg string, err error)

		WarnWithError(msg string, err error)

		Error(msg string, err error)
	}

	// printfStyle
	printfStyle interface {
		Infof(template string, args ...interface{})

		Debugf(template string, args ...interface{})

		Warnf(template string, args ...interface{})
	}

	// printfStyleWithError will add error param to the end of template string
	// something like {... message="my info message arg1, arg2, arg3: err=\"error_text\"" ...}
	//
	printfStyleWithError interface {
		InfofWithError(template string, err error, args ...interface{})

		DebugfWithError(template string, err error, args ...interface{})

		WarnfWithError(template string, err error, args ...interface{})

		Errorf(template string, err error, args ...interface{})
	}
)
