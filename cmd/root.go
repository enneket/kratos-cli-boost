package cmd

import (
	"kratos_cli_boost/internal/add"
	"kratos_cli_boost/internal/biz"
	"kratos_cli_boost/internal/client"
	"kratos_cli_boost/internal/data"
	"kratos_cli_boost/internal/server"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kratos-cli-boost",
	Short: "Kratos code generation tool",
	Long:  `A tool to generate proto、biz、data and sever layer code for Kratos`,
}

var protoCmd = &cobra.Command{
	Use:   "proto",
	Short: "Generate the proto files.",
	Long:  "Generate the proto files.",
}

func init() {
	rootCmd.AddCommand(protoCmd)

	protoCmd.AddCommand(add.CmdAdd)
	protoCmd.AddCommand(client.CmdClient)
	protoCmd.AddCommand(server.CmdServer)
	protoCmd.AddCommand(biz.CmdBiz)
	protoCmd.AddCommand(data.CmdData)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
