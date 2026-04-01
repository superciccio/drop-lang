package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/superciccio/drop-lang/internal/editor"
	"github.com/superciccio/drop-lang/internal/interpreter"
	"github.com/superciccio/drop-lang/internal/lexer"
	"github.com/superciccio/drop-lang/internal/parser"
)

const helpText = `Drop — a zero-ceremony prototyping language

Usage:
  drop <file.drop>    Run a Drop program

Editor support:
  drop --editor vscode     Install syntax highlighting for VS Code
  drop --editor vim        Install for Vim / Neovim
  drop --editor sublime    Install for Sublime Text
  drop --editor zed        Install for Zed

Options:
  --help, -h    Show this help

Examples:
  drop hello.drop     Run a script
  drop server.drop    Start a web app (with hot reload)

Learn more: https://github.com/superciccio/drop-lang`

func main() {
	if len(os.Args) < 2 {
		fmt.Println(helpText)
		os.Exit(0)
	}

	if os.Args[1] == "--help" || os.Args[1] == "-h" {
		fmt.Println(helpText)
		os.Exit(0)
	}

	if os.Args[1] == "--editor" {
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: drop --editor <vscode|vim|sublime|zed>")
			os.Exit(1)
		}
		var err error
		switch os.Args[2] {
		case "vscode", "code":
			err = editor.InstallVSCode()
		case "vim", "neovim", "nvim":
			err = editor.InstallVim()
		case "sublime":
			err = editor.InstallSublime()
		case "zed":
			err = editor.InstallZed()
		default:
			fmt.Fprintf(os.Stderr, "Unknown editor: %s\nSupported: vscode, vim, sublime, zed\n", os.Args[2])
			os.Exit(1)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	filename := os.Args[1]

	interp, isServer, err := run(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Non-server scripts: run once and exit
	if !isServer {
		return
	}

	// Server apps: watch for changes and hot reload
	watchAndReload(filename, interp)
}

// run reads, lexes, parses, and executes a .drop file.
// Returns the interpreter, whether it started a server, and any error.
func run(filename string) (*interpreter.Interpreter, bool, error) {
	source, err := os.ReadFile(filename)
	if err != nil {
		return nil, false, fmt.Errorf("Error: could not read %s: %v", filename, err)
	}

	// Lex
	lex := lexer.New(string(source))
	tokens, err := lex.Tokenize()
	if err != nil {
		return nil, false, fmt.Errorf("Syntax error: %v", err)
	}

	// Parse
	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		return nil, false, fmt.Errorf("Parse error: %v", err)
	}

	// Run
	interp := interpreter.New()

	// For server apps, Run() blocks in server.Start().
	// Run in a goroutine so we can detect whether it's a server or a script.
	errCh := make(chan error, 1)
	go func() {
		errCh <- interp.Run(program)
	}()

	// Wait briefly: scripts finish fast, servers keep running
	select {
	case err := <-errCh:
		if err != nil {
			return nil, false, fmt.Errorf("Runtime error: %v", err)
		}
		return interp, false, nil
	case <-time.After(100 * time.Millisecond):
		// Still running — it's a server app
		go func() {
			if err := <-errCh; err != nil {
				fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			}
		}()
		return interp, true, nil
	}
}

func watchAndReload(filename string, current *interpreter.Interpreter) {
	lastMod := getModTime(filename)

	// Handle SIGINT/SIGTERM for clean shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sig:
			fmt.Println("\nShutting down...")
			current.Stop()
			os.Exit(0)

		case <-ticker.C:
			mod := getModTime(filename)
			if mod == lastMod {
				continue
			}
			lastMod = mod

			fmt.Println("\nReloading...")

			// Stop the current server
			current.Stop()

			// Small delay to let the port free up
			time.Sleep(50 * time.Millisecond)

			// Try to start the new version
			interp, isServer, err := run(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				fmt.Fprintln(os.Stderr, "Fix the error and save again.")
				continue
			}

			if !isServer {
				fmt.Fprintln(os.Stderr, "Warning: no 'serve' statement found. Waiting for changes...")
				continue
			}

			current = interp
		}
	}
}

func getModTime(filename string) time.Time {
	info, err := os.Stat(filename)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}
