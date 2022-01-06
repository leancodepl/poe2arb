package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	poeCmd = &cobra.Command{
		Use:  "poe",
		RunE: runPoe,
	}
)

func init() {
	poeCmd.Flags().StringP("token", "t", "", "POEditor API token")
	poeCmd.MarkFlagRequired("token")

	poeCmd.Flags().StringP("project-id", "p", "", "POEditor project ID")
	poeCmd.MarkFlagRequired("project-id")
}

func runPoe(cmd *cobra.Command, args []string) error {
	projectID, _ := cmd.Flags().GetString("project-id")
	token, _ := cmd.Flags().GetString("token")

	fmt.Printf("Project ID: %s\nToken: %s\n", projectID, token)

	return nil
}
