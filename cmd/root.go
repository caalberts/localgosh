package cmd

import (
	"github.com/caalberts/localroast/filesystem"
	"github.com/caalberts/localroast/http"
	"github.com/caalberts/localroast/json"
	"github.com/caalberts/localroast/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
)

var (
	version string
	port    string
)

func Execute(v string) {
	version = v
	commandBuilder, err := newCommandBuilder()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	commandBuilder.build().Execute()
}

type commandBuilder struct {
	fileHandler fileHandler
	jsonParser  parser
	serverFunc  http.ServerFunc
}

type fileHandler interface {
	Output() chan io.Reader
	Open(fileName string) error
	Watch() error
}

type parser interface {
	Output() chan []types.Schema
	Watch(chan io.Reader)
}

func newCommandBuilder() (*commandBuilder, error) {
	fileHandler, err := filesystem.NewFileHandler()
	if err != nil {
		return nil, err
	}
	jsonParser := json.NewParser()

	return &commandBuilder{
		fileHandler: fileHandler,
		jsonParser:  jsonParser,
		serverFunc:  http.NewServer,
	}, nil
}

func (b *commandBuilder) build() *cobra.Command {
	jsonCmd := newJSONCmd(b.fileHandler, b.jsonParser, b.serverFunc)
	versionCmd := newVersionCmd()
	root := newRootCmd(jsonCmd)
	rootCmd := root.getCommand()
	addSubcommands(rootCmd, jsonCmd, versionCmd)
	return rootCmd
}

func addSubcommands(root *cobra.Command, children ...commander) {
	for _, child := range children {
		root.AddCommand(child.getCommand())
	}
}

type commander interface {
	getCommand() *cobra.Command
}

type basicCommand struct {
	*cobra.Command
}

func (c *basicCommand) getCommand() *cobra.Command {
	return c.Command
}

func newRootCmd(defaultCmder commander) commander {
	defaultCmd := defaultCmder.getCommand()
	cmd := &cobra.Command{
		Use:   "localroast",
		Short: "Localroast quickly stubs a HTTP server",
		Long: `A tool to help developers stub external HTTP services quickly.
See https://github.com/caalberts/localroast/examples/stubs.json
for examples.`,
		Args:    defaultCmd.Args,
		Example: "localroast examples/stubs.json",
		RunE:    defaultCmd.RunE,
	}
	cmd.PersistentFlags().StringVarP(&port, "port", "p", "8080", "port number")

	return &basicCommand{cmd}
}
