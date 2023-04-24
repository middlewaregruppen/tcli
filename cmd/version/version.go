package version

import (
	"fmt"

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

func NewCmdVersion() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:     "version",
		Short:   "Prints the tcli version",
		Example: `tcli version`,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("Version: %s%v%s\t\n", colorGreen, VERSION, colorReset)
			fmt.Printf("Built: %v\t\n", DATE)
			fmt.Printf("Commit: %v\t\n", COMMIT)
			fmt.Printf("Branch: %v\t\n", BRANCH)
		},
	}
	return versionCmd
}
