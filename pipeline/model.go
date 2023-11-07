package pipeline

import "errors"

// Return this error in executors if you just want to skip this entry
var SkipError error = errors.New("skip this output")
