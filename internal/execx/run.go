package execx

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

func RunCmd(ctx context.Context, dir, name string, args ...string) error {
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
