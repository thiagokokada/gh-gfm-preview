package cmd

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
	"github.com/thiagokokada/gh-gfm-preview/internal/server"
	"github.com/thiagokokada/gh-gfm-preview/internal/utils"
)

var logLevel = new(slog.LevelVar)

var rootCmd = &cobra.Command{
	Use:     "gh-gfm-preview",
	Short:   "GitHub CLI extension to preview Markdown",
	Run:     run,
	Args:    cobra.RangeArgs(0, 1),
	Version: version(),
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntP("port", "p", 3333, "TCP port number of this server")
	rootCmd.Flags().StringP("host", "H", "localhost", "hostname this server will bind")
	rootCmd.Flags().BoolP("disable-reload", "R", false, "disable live reloading")
	rootCmd.Flags().BoolP("markdown-mode", "m", false, "force \"markdown\" mode (rather than default \"gfm\")")
	rootCmd.Flags().BoolP("disable-auto-open", "A", false, "disable auto opening your browser")
	rootCmd.Flags().BoolP("verbose", "v", false, "show verbose output")
	rootCmd.Flags().BoolP("light-mode", "l", false, "force light mode")
	rootCmd.Flags().BoolP("dark-mode", "d", false, "force dark mode")
	rootCmd.Flags().BoolP("directory-listing", "D", false, "enable directory browsing mode")
	rootCmd.Flags().StringP("directory-listing-show-extensions", "", ".md,.txt", "file extensions to show in directory listing (comma-separated, use '*' for all files)")
	rootCmd.Flags().StringP("directory-listing-text-extensions", "", ".md,.txt", "text file extensions for preview (comma-separated, others will be served as binary)")
}

func detectStdin(filename string) (bool, string) {
	switch filename {
	case "-":
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("Error reading stdin: %v", err)
		}

		return true, string(data)
	case "":
		if fi, _ := os.Stdin.Stat(); (fi.Mode() & os.ModeCharDevice) == 0 {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				log.Fatalf("Error reading stdin: %v", err)
			}

			return true, string(data)
		}
	}

	return false, ""
}

func run(cmd *cobra.Command, args []string) {
	filename := ""
	if len(args) > 0 {
		filename = args[0]
	}

	flags := cmd.Flags()

	verbose := utils.Must(flags.GetBool("verbose"))
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	slog.SetDefault(slog.New(h))

	if verbose {
		logLevel.Set(slog.LevelDebug)
	}

	host := utils.Must(flags.GetString("host"))
	port := utils.Must(flags.GetInt("port"))
	httpServer := server.Server{Host: host, Port: port}

	disableReload := utils.Must(flags.GetBool("disable-reload"))

	forceLightMode := utils.Must(flags.GetBool("light-mode"))
	forceDarkMode := utils.Must(flags.GetBool("dark-mode"))

	markdownMode := utils.Must(flags.GetBool("markdown-mode"))

	disableAutoOpen := utils.Must(flags.GetBool("disable-auto-open"))

	// Detect stdin usage
	useStdin, stdinContent := detectStdin(filename)

	directoryListing := utils.Must(flags.GetBool("directory-listing"))
	directoryListingShowExtensions := utils.Must(flags.GetString("directory-listing-show-extensions"))
	directoryListingTextExtensions := utils.Must(flags.GetString("directory-listing-text-extensions"))

	param := &server.Param{
		Filename:                       filename,
		MarkdownMode:                   markdownMode,
		Reload:                         !disableReload,
		ForceLightMode:                 forceLightMode,
		ForceDarkMode:                  forceDarkMode,
		AutoOpen:                       !disableAutoOpen,
		UseStdin:                       useStdin,
		StdinContent:                   stdinContent,
		DirectoryListing:               directoryListing,
		DirectoryListingShowExtensions: directoryListingShowExtensions,
		DirectoryListingTextExtensions: directoryListingTextExtensions,
	}

	err := httpServer.Serve(param)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func version() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	return buildInfo.Main.Version
}
