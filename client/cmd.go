package client

import (
	"encoding/json"
	"log"
	"os"

	"github.com/lkyzhu/qproxy/client/conf"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "client",
		Run: run,
	}

	return cmd
}

func run(cmd *cobra.Command, args []string) {
	path, err := cmd.InheritedFlags().GetString("conf")
	if err != nil {
		log.Fatalf("get args conf fail, err:%v\n", err.Error())
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read config file[%v] fail, err:%v\n", path, err.Error())
	}

	config := conf.Config{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("unmarsha config fail, err:%v\n", err.Error())
	}

	client := NewClient(&config)
	client.Run()
}
