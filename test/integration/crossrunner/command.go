package crossrunner

import (
	"fmt"
	"os/exec"
	"strings"
)

// getCommand returns a Cmd struct used to execute a client or server and a
// nicely formatted string for verbose loggings
func getCommand(config config, port int) (cmd *exec.Cmd, formatted string) {

	command := config.Command[0]

	args := append(config.Command[1:],
		fmt.Sprintf("--protocol=%s", config.Protocol),
		fmt.Sprintf("--transport=%s", config.Transport),
		fmt.Sprintf("--port=%v", port),
	)

	cmd = exec.Command(command, args...)
	cmd.Dir = config.Workdir
	cmd.Stdout = config.Logs
	cmd.Stderr = config.Logs

	// Nicely format command here for use at the top of each log file
	formatted = fmt.Sprintf("%s %s", command, strings.Trim(fmt.Sprint(args), "[]"))

	return cmd, formatted
}
