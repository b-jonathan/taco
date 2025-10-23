package execx

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// TODO: Make a helper so that you can run whole strings instead of a bunch of strings.
func RunCmd(ctx context.Context, dir string, cmd string) error {
	cmdArgs := strings.Split(cmd, " ")
	name := cmdArgs[0]
	args := cmdArgs[1:]
	c := exec.CommandContext(ctx, name, args...)
	c.Dir = dir
	var out, errb bytes.Buffer
	c.Stdout, c.Stderr = &out, &errb
	if err := c.Run(); err != nil {
		return fmt.Errorf("%s %v failed: %v\nstdout:\n%s\nstderr:\n%s",
			name, args, err, out.String(), errb.String())
	}
	return nil
}
