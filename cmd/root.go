package cmd

import (
	"os"

	"github.com/enneket/kratos_cli_boost/internal/add"
	"github.com/enneket/kratos_cli_boost/internal/biz"
	"github.com/enneket/kratos_cli_boost/internal/client"
	"github.com/enneket/kratos_cli_boost/internal/data"
	"github.com/enneket/kratos_cli_boost/internal/server"

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
