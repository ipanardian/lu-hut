// Package terminal provides terminal-related utilities like help display.
package terminal

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	cobra.AddTemplateFunc("luHeader", luHeader)
	cobra.AddTemplateFunc("luCmd", luCmd)
	cobra.AddTemplateFunc("luDesc", luDesc)
}

func luHeader(s string) string {
	return color.New(color.FgWhite, color.Bold).Sprint(s)
}

func luCmd(s string) string {
	return color.New(color.FgYellow, color.Bold).Sprint(s)
}

func luDesc(s string) string {
	return color.New(color.FgHiWhite).Sprint(s)
}

// UsageTemplate returns a Cobra usage template that wraps the default
// sections (Usage, Available Commands, Flags, etc.) in lu-hut's color
// scheme. Register it on the root command so every subcommand inherits
// the same layout via Cobra's template inheritance.
func UsageTemplate() string {
	return `{{luHeader "Usage:"}}{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

{{luHeader "Available Commands:"}}{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{luCmd (rpad .Name .NamePadding) }} {{luDesc .Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{luCmd (rpad .Name .NamePadding) }} {{luDesc .Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

{{luHeader "Additional Commands:"}}{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{luCmd (rpad .Name .NamePadding) }} {{luDesc .Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{luHeader "Flags:"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{luHeader "Global Flags:"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
}

// Banner returns a colored "lu-hut <version>" prefix suitable for embedding
// in a Cobra command's Long description.
func Banner(version, tagline string) string {
	return color.New(color.FgCyan, color.Bold).Sprint("lu-hut "+version) +
		"\n" + color.New(color.FgHiWhite).Sprint(tagline)
}
