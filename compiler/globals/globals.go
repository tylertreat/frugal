package globals

import "time"

const Version = "0.0.1"

var (
	TopicDelimiter  string = "."
	Gen             string
	Out             string
	FileDir         string
	DryRun          bool
	Now             = time.Now()
	IntermediateIDL = []string{}
)

func Reset() {
	TopicDelimiter = "."
	Gen = ""
	Out = ""
	FileDir = ""
	DryRun = false
	Now = time.Now()
	IntermediateIDL = []string{}
}
