// Command hwhkit is the hwhkit-go CLI: scaffolds projects, runs migrations, brings up dev deps.
package main

import (
	"fmt"
	"os"

	"github.com/hwhkit/hwhkit-go/buildinfo"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "hwhkit",
		Short: "hwhkit-go scaffold and ops CLI",
	}
	root.AddCommand(initCmd())
	root.AddCommand(migrateCmd())
	root.AddCommand(devCmd())
	root.AddCommand(versionCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version + git SHA",
		Run: func(*cobra.Command, []string) {
			info := buildinfo.Get()
			fmt.Printf("hwhkit %s (sha=%s, built=%s, go=%s)\n",
				info.Version, info.GitSHA, info.BuildTime, info.GoVersion)
		},
	}
}
