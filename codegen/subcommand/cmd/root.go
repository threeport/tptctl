/*
Copyright © 2023 Threeport admin@threeport.io
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
)

type Subcommand struct {
	Name             string
	CamelName        string
	LowerCamelName   string
	Parent           string
	CamelParent      string
	LowerCamelParent string
	Filename         string
	LoadConfig       bool
	OptionalConfig   bool
}

var (
	loadConfig     bool
	optionalConfig bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "subcommand <parent command> <subcommand>",
	Example: `Create a new "widget" subcommand that includes code to load a
required config file path with widget attributes:
go run codegen/subcommand/main.go create widget --load-config`,
	Short: "Genearate the source code for a new subcommand",
	Long: `Genearate the source code for a new subcommand.

Use this tool to add a new subcommand to an existing parent command.  The
common parent commands are:
* get
* create
* update
* delete
For example, if you want to extend tptctl to create a new Threeport object
called "widget," the parent command will be "create" and the subcommand "widget".

The --load-config flag will add source code to load a config file for the attributes
of the object will be included.

The --optional-config flag will make that config file an optional flag.`,
	SilenceUsage: true,
	Args:         cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		parent := strings.ToLower(args[0])
		name := strings.ToLower(args[1])

		// check for existing source code file
		filename := fmt.Sprintf(
			"cmd/%s_%s.go",
			strcase.ToSnake(parent),
			strcase.ToSnake(name),
		)
		if _, err := os.Stat(filename); err == nil {
			return errors.New(fmt.Sprintf("source code file %s already exists", filename))
		}

		// assemble subcommand attributes
		subcmd := Subcommand{
			Name:             name,
			CamelName:        strcase.ToCamel(name),
			LowerCamelName:   strcase.ToLowerCamel(name),
			Parent:           parent,
			CamelParent:      strcase.ToCamel(parent),
			LowerCamelParent: strcase.ToLowerCamel(parent),
			LoadConfig:       loadConfig,
			OptionalConfig:   optionalConfig,
		}

		// parse template and write source code file
		tmpl := template.New("subcommand")
		t, err := tmpl.Parse(subcommandTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse source code template: %w", err)
		}
		sourceFile, err := os.Create(filename)
		if err != nil {
			fmt.Printf("Error: failed to create source file: %s\n", err)
			os.Exit(1)
		}
		if err := t.Execute(sourceFile, subcmd); err != nil {
			fmt.Printf("Error: failed template source code: %s\n", err)
			os.Exit(1)
		}

		if loadConfig {
			fmt.Printf("Note: the source code generated includes a reference to `config.%sConfig` which you will have to create.\n", strcase.ToCamel(name))
		}

		fmt.Printf("Complete: subcommand source code file written to %s\n", filename)

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&loadConfig, "load-config", "l", false, "Include config loading code")
	rootCmd.Flags().BoolVarP(&optionalConfig, "optional-config", "o", false, "Make config optional for the tptctl user")
}

const subcommandTemplate = `/*
Copyright © 2023 Threeport admin@threeport.io
*/
package cmd

import (
	"fmt"
	"os"
{{- if .LoadConfig }}
	"io/ioutil"
{{- end }}

	"github.com/spf13/cobra"
{{- if .LoadConfig }}
	"gopkg.in/yaml.v2"

	"github.com/threeport/tptctl/internal/api"
	qout "github.com/threeport/tptctl/internal/output"
{{- end }}
)

{{ if .LoadConfig -}}
var {{ .LowerCamelParent }}{{ .CamelName }}ConfigPath string
{{- end }}

// {{ .LowerCamelParent }}{{ .CamelName }}Cmd represents the {{ .Parent }} {{ .Name }} command
var {{ .LowerCamelParent }}{{ .CamelName }}Cmd = &cobra.Command{
	Use: "{{ .Name }}",
{{- if .LoadConfig }}
	Example: "tptctl {{ .Parent }} {{ .Name }} -c /path/to/config.yaml",
{{- else }}
	Example: "tptctl {{ .Parent }} {{ .Name }}",
{{- end }}
	Short: "A brief description of your command",
	Long: ` + "`" + `A long description of your command.` + "`" + `,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
{{- if .LoadConfig }}
		// load config
		configContent, err := ioutil.ReadFile({{ .LowerCamelParent }}{{ .CamelName }}ConfigPath)
		if err != nil {
			qout.Error("failed to read config file", err)
			os.Exit(1)
		}
		var {{ .LowerCamelName }} api.{{ .CamelName }}Config
		if err := yaml.Unmarshal(configContent, &{{ .LowerCamelName }}); err != nil {
			qout.Error("failed to unmarshal config file yaml content", err)
			os.Exit(1)
		}

		qout.Info(fmt.Sprintf("%+v\n", {{ .LowerCamelName }}))
{{ end }}
		qout.Complete("{{ .Parent }} {{ .Name }} called")
	},
}

func init() {
	{{ .LowerCamelParent }}Cmd.AddCommand({{ .LowerCamelParent }}{{ .CamelName }}Cmd)

	{{ if .LoadConfig -}}
	{{ .LowerCamelParent }}{{ .CamelName }}Cmd.Flags().StringVarP(&{{ .LowerCamelParent }}{{ .CamelName }}ConfigPath, "config", "c", "", "path to file with {{ .Name }} config")
	{{- if not .OptionalConfig }}
	{{ .LowerCamelParent }}{{ .CamelName }}Cmd.MarkFlagRequired("config")
	{{- end -}}
	{{- end }}
}
`
