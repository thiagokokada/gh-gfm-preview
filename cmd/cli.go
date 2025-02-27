package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var Version string

var verbose = false

type Param struct {
	filename       string
	markdownMode   bool
	reload         bool
	forceLightMode bool
	forceDarkMode  bool
	autoOpen       bool
}

var programName = "gh-gfm-preview"
var rootCmd = &cobra.Command{
	Use:   programName,
	Short: "GitHub CLI extension to preview Markdown",
	Run: func(cmd *cobra.Command, args []string) {

		showVerionFlag, _ := cmd.Flags().GetBool("version")
		if showVerionFlag {
			showVersion()
			os.Exit(0)
		}

		filename := ""
		if len(args) > 0 {
			filename = args[0]
		}

		verbose, _ = cmd.Flags().GetBool("verbose")

		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")

		server := Server{host: host, port: port}

		disableReload, _ := cmd.Flags().GetBool("disable-reload")
		reload := true
		if disableReload {
			reload = false
		}

		forceLightMode, _ := cmd.Flags().GetBool("light-mode")
		forceDarkMode, _ := cmd.Flags().GetBool("dark-mode")

		markdownMode, _ := cmd.Flags().GetBool("markdown-mode")

		disableAutoOpen, _ := cmd.Flags().GetBool("disable-auto-open")
		autoOpen := true
		if disableAutoOpen {
			autoOpen = false
		}

		param := &Param{
			filename:       filename,
			markdownMode:   markdownMode,
			reload:         reload,
			forceLightMode: forceLightMode,
			forceDarkMode:  forceDarkMode,
			autoOpen:       autoOpen,
		}

		err := server.Serve(param)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntP("port", "p", 3333, "TCP port number of this server")
	rootCmd.Flags().StringP("host", "", "localhost", "hostname this server will bind")
	rootCmd.Flags().BoolP("version", "", false, "show the version")
	rootCmd.Flags().BoolP("disable-reload", "", false, "disable live reloading")
	rootCmd.Flags().BoolP("markdown-mode", "", false, "force \"markdown\" mode (rather than default \"gfm\")")
	rootCmd.Flags().BoolP("disable-auto-open", "", false, "disable auto opening your browser")
	rootCmd.Flags().BoolP("verbose", "", false, "show verbose output")
	rootCmd.Flags().BoolP("light-mode", "", false, "force light mode")
	rootCmd.Flags().BoolP("dark-mode", "", false, "force dark mode")
}

func showVersion() {
	fmt.Printf("%s version %s\n", programName, Version)
}
