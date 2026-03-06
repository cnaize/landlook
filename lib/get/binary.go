package get

import (
	"context"
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

func BinaryDeps(ctx context.Context, binPath string) ([]string, error) {
	var skipList = []string{"vdso"}

	output, err := exec.CommandContext(ctx, "ldd", binPath).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ldd binary: %s", output)
	}

	var deps []string
	for dep := range strings.SplitSeq(string(output), "\n") {
		items := strings.Split(strings.TrimSpace(dep), " ")
		if len(items) < 1 {
			continue
		}

		if len(items) < 3 {
			dep = items[0]
		} else {
			dep = items[2]
		}

		if !slices.ContainsFunc(skipList, func(skip string) bool { return strings.Contains(dep, skip) }) {
			deps = append(deps, dep)
		}
	}

	return deps, nil
}
