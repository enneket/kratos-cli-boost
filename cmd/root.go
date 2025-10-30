package cmd

import (
	"os"

	"kratos_cli_boost/internal/generator"

	"github.com/spf13/cobra"
)

var targetDir string

var rootCmd = &cobra.Command{
	Use:   "cli",
	Short: "Kratos code generation tool",
	Long:  `A tool to generate biz and data layer code for Kratos based on proto files`,
}

var protoCmd = &cobra.Command{
	Use:   "proto",
	Short: "Generate code from proto files",
}

var bizCmd = &cobra.Command{
	Use:   "biz [proto-file]",
	Short: "Generate biz layer code",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		protoFile := args[0]
		return generator.GenerateBiz(protoFile, targetDir)
	},
}

var dataCmd = &cobra.Command{
	Use:   "data [proto-file]",
	Short: "Generate data layer code",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		protoFile := args[0]
		return generator.GenerateData(protoFile, targetDir)
	},
}

func init() {
	rootCmd.AddCommand(protoCmd)
	protoCmd.AddCommand(bizCmd)
	protoCmd.AddCommand(dataCmd)

	// Add target directory flag
	bizCmd.Flags().StringVarP(&targetDir, "target", "t", "", "Target directory for generated code")
	dataCmd.Flags().StringVarP(&targetDir, "target", "t", "", "Target directory for generated code")

	_ = bizCmd.MarkFlagRequired("target")
	_ = dataCmd.MarkFlagRequired("target")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
