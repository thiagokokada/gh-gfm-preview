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
	Run: func(cmd *cobra.Command, args []string) {
		filename := ""
		if len(args) > 0 {
			filename = args[0]
		}

		utils.Verbose = utils.Must(cmd.Flags().GetBool("verbose"))

		host := utils.Must(cmd.Flags().GetString("host"))
		port := utils.Must(cmd.Flags().GetInt("port"))
		httpServer := server.Server{Host: host, Port: port}

		disableReload := utils.Must(cmd.Flags().GetBool("disable-reload"))
		reload := true
		if disableReload {
			reload = false
		}

		forceLightMode := utils.Must(cmd.Flags().GetBool("light-mode"))
		forceDarkMode := utils.Must(cmd.Flags().GetBool("dark-mode"))

		markdownMode := utils.Must(cmd.Flags().GetBool("markdown-mode"))

		disableAutoOpen := utils.Must(cmd.Flags().GetBool("disable-auto-open"))

		autoOpen := true
		if disableAutoOpen {
			autoOpen = false
		}

		param := &server.Param{
			Filename:       filename,
			MarkdownMode:   markdownMode,
			Reload:         reload,
			ForceLightMode: forceLightMode,
			ForceDarkMode:  forceDarkMode,
			AutoOpen:       autoOpen,
		}

		err := httpServer.Serve(param)
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
	rootCmd.Flags().BoolP("disable-reload", "", false, "disable live reloading")
	rootCmd.Flags().BoolP("markdown-mode", "", false, "force \"markdown\" mode (rather than default \"gfm\")")
	rootCmd.Flags().BoolP("disable-auto-open", "", false, "disable auto opening your browser")
	rootCmd.Flags().BoolP("verbose", "", false, "show verbose output")
	rootCmd.Flags().BoolP("light-mode", "", false, "force light mode")
	rootCmd.Flags().BoolP("dark-mode", "", false, "force dark mode")
}
