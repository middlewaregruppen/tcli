package logout

import (
	"fmt"

	"github.com/goretk/gore"
	"github.com/spf13/cobra"
)

var (
	binary string
)

func NewCmdLogout() *cobra.Command {
	c := &cobra.Command{
		Use:   "logout",
		Short: "Logout user and remove credentials",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			f, err := gore.Open(binary)
			if err != nil {
				return err
			}
			defer f.Close()
			typs, err := f.GetTypes()
			if err != nil {
				return err
			}
			for _, typ := range typs {
				fmt.Printf("%s: \n", typ.Name)
				for _, field := range typ.Fields {
					fmt.Printf("  %s\n", field.FieldName)
				}
			}
			return nil
		},
	}
	c.Flags().StringVarP(&binary, "binary", "b", "", "Path to binary to open.")
	return c
}
