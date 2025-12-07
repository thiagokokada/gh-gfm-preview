package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/lmittmann/tint"
	"github.com/spf13/pflag"
	"github.com/thiagokokada/gh-gfm-preview/internal/server"
)

var logLevel = new(slog.LevelVar)

func Execute() {
	fs := pflag.NewFlagSet("gh-gfm-preview", pflag.ExitOnError)
	fs.SortFlags = false

	port := fs.IntP("port", "p", 3333, "TCP port number of this server")
	host := fs.StringP("host", "H", "localhost", "hostname this server will bind")
	disableReload := fs.BoolP("disable-reload", "R", false, "disable live reloading")
	disableAutoOpen := fs.BoolP("disable-auto-open", "A", false, "disable auto opening your browser")
	lightMode := fs.BoolP("light-mode", "l", false, "force light mode")
	darkMode := fs.BoolP("dark-mode", "d", false, "force dark mode")
	markdownMode := fs.BoolP("markdown-mode", "m", false, `force "markdown" mode (rather than default "gfm")`)
	directoryListing := fs.BoolP("directory-listing", "D", false, "enable directory browsing mode")
	directoryListingShowExtensions := fs.StringP("directory-listing-show-extensions", "", ".md,.txt", "file extensions to show in directory listing (comma-separated, use '*' for all files)")
	directoryListingTextExtensions := fs.StringP("directory-listing-text-extensions", "", ".md,.txt", "text file extensions for preview (comma-separated, others will be served as binary)")
	noColor := fs.BoolP("no-color", "", false, "disable color for logs")
	verbose := fs.BoolP("verbose", "v", false, "show verbose output")
	version := fs.BoolP("version", "", false, "show program version")

	_ = fs.Parse(os.Args[1:])

	if *version {
		fmt.Println(getVersion())
		os.Exit(0)
	}

	filename := ""
	if fs.NArg() > 0 {
		filename = fs.Arg(0)
	}

	if *verbose {
		logLevel.Set(slog.LevelDebug)
	}

	h := slog.New(tint.NewHandler(
		os.Stdout,
		&tint.Options{
			Level:   logLevel,
			NoColor: *noColor || os.Getenv("NO_COLOR") == "1",
		},
	))
	slog.SetDefault(h)

	// Detect stdin usage
	useStdin, stdinContent := detectStdin(filename)

	param := &server.Param{
		Filename:                       filename,
		MarkdownMode:                   *markdownMode,
		Reload:                         !*disableReload,
		ForceLightMode:                 *lightMode,
		ForceDarkMode:                  *darkMode,
		AutoOpen:                       !*disableAutoOpen,
		UseStdin:                       useStdin,
		StdinContent:                   stdinContent,
		DirectoryListing:               *directoryListing,
		DirectoryListingShowExtensions: *directoryListingShowExtensions,
		DirectoryListingTextExtensions: *directoryListingTextExtensions,
	}

	httpServer := server.Server{Host: *host, Port: *port}

	err := httpServer.Serve(param)
	if err != nil {
		slog.Error("Error while starting HTTP server", "error", err)
		os.Exit(1)
	}
}

func detectStdin(filename string) (bool, string) {
	switch filename {
	case "-":
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			slog.Error("Error while reading stdin", "error", err)
			os.Exit(1)
		}

		return true, string(data)
	case "":
		if fi, _ := os.Stdin.Stat(); (fi.Mode() & os.ModeCharDevice) == 0 {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				slog.Error("Error while reading stdin", "error", err)
				os.Exit(1)
			}

			return true, string(data)
		}
	}

	return false, ""
}

func getVersion() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	return buildInfo.Main.Version
}
