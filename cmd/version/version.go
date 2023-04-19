package version

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var (
	// VERSION of the app. Is set when project is built and should never be set manually
	VERSION string
	// COMMIT is the Git commit currently used when compiling. Is set when project is built and should never be set manually
	COMMIT string
	// BRANCH is the Git branch currently used when compiling. Is set when project is built and should never be set manually
	BRANCH string
	// GOVERSION used to compile. Is set when project is built and should never be set manually
	GOVERSION string
	// DATE used to compile. Is set when project is built and should never be set manually
	DATE string
)

func NewCmdVersion(w io.Writer) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:     "version",
		Short:   "Prints the tcli version",
		Example: `tcli version`,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("tcli version %v\n", VERSION)
			fmt.Printf("built %v from commit %v branch %s", DATE, COMMIT, BRANCH)
		},
	}
	return versionCmd
}
