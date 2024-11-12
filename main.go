package main

import (
	"os"

	"github.com/lkyzhu/qproxy/client"
	"github.com/lkyzhu/qproxy/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	logrus.SetOutput(os.Stderr)
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {

	cmd := cobra.Command{
		Use: "qproxy",
	}

	cmd.PersistentFlags().String("conf", "./config.json", "config file path")
	cmd.AddCommand(client.NewCommand())
	cmd.AddCommand(server.NewCommand())

	cmd.Execute()
}
