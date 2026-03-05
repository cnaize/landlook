package get

import (
	"context"
	"os/exec"
	"slices"
	"strings"
)

func BinaryDeps(ctx context.Context, binPath string) ([]string, error) {
	var skipList = []string{"vdso"}

	output, err := exec.CommandContext(ctx, "ldd", binPath).Output()
	if err != nil {
		// assume static binary, so skip
		return nil, nil
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
