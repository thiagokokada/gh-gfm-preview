package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/thiagokokada/gh-gfm-preview/internal/server"
	"github.com/thiagokokada/gh-gfm-preview/internal/utils"
)

var rootCmd = &cobra.Command{
	Use:   "gh-gfm-preview",
	Short: "GitHub CLI extension to preview Markdown",
	Run:   run,
	Args:  cobra.RangeArgs(0, 1),
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
	rootCmd.Flags().BoolP("disable-reload", "", false, "disable live reloading")
	rootCmd.Flags().BoolP("markdown-mode", "", false, "force \"markdown\" mode (rather than default \"gfm\")")
	rootCmd.Flags().BoolP("disable-auto-open", "", false, "disable auto opening your browser")
	rootCmd.Flags().BoolP("verbose", "", false, "show verbose output")
	rootCmd.Flags().BoolP("light-mode", "", false, "force light mode")
	rootCmd.Flags().BoolP("dark-mode", "", false, "force dark mode")
}

func run(cmd *cobra.Command, args []string) {
	filename := ""
	if len(args) > 0 {
		filename = args[0]
	}

	flags := cmd.Flags()

	verbose := utils.Must(flags.GetBool("verbose"))
	utils.SetVerbose(verbose)

	host := utils.Must(flags.GetString("host"))
	port := utils.Must(flags.GetInt("port"))
	httpServer := server.Server{Host: host, Port: port}

	disableReload := utils.Must(flags.GetBool("disable-reload"))

	forceLightMode := utils.Must(flags.GetBool("light-mode"))
	forceDarkMode := utils.Must(flags.GetBool("dark-mode"))

	markdownMode := utils.Must(flags.GetBool("markdown-mode"))

	disableAutoOpen := utils.Must(flags.GetBool("disable-auto-open"))

	param := &server.Param{
		Filename:       filename,
		MarkdownMode:   markdownMode,
		Reload:         !disableReload,
		ForceLightMode: forceLightMode,
		ForceDarkMode:  forceDarkMode,
		AutoOpen:       !disableAutoOpen,
	}

	err := httpServer.Serve(param)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
