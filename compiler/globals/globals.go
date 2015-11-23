package globals

import "time"

const Version = "0.0.1"

var (
	TopicDelimiter  string = "."
	Gen             string
	Out             string
	FileDir         string
	Now             = time.Now()
	IntermediateIDL = []string{}
)
