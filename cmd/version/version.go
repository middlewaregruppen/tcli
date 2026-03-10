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
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Printf("Version:   %s%v%s\n", colorGreen, VERSION, colorReset)
			fmt.Printf("Built:     %v\n", DATE)
			fmt.Printf("Commit:    %v\n", COMMIT)
			fmt.Printf("Branch:    %v\n", BRANCH)
			fmt.Printf("Go:        %v\n", GOVERSION)
			return nil
		},
	}
	return versionCmd
}
