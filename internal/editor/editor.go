package editor

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

//go:embed drop.tmLanguage.json
var tmGrammar []byte

//go:embed drop.vim
var vimSyntax []byte

func InstallVSCode() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// VS Code local extension
	extDir := filepath.Join(home, ".vscode", "extensions", "drop-lang")
	synDir := filepath.Join(extDir, "syntaxes")

	if err := os.MkdirAll(synDir, 0755); err != nil {
		return err
	}

	// Write grammar
	if err := os.WriteFile(filepath.Join(synDir, "drop.tmLanguage.json"), tmGrammar, 0644); err != nil {
		return err
	}

	// Write package.json
	pkg := map[string]interface{}{
		"name":        "drop-lang",
		"displayName": "Drop Language",
		"description": "Syntax highlighting for the Drop prototyping language",
		"version":     "0.1.0",
		"engines":     map[string]string{"vscode": "^1.60.0"},
		"categories":  []string{"Programming Languages"},
		"contributes": map[string]interface{}{
			"languages": []map[string]interface{}{
				{
					"id":         "drop",
					"aliases":    []string{"Drop", "drop"},
					"extensions": []string{".drop"},
				},
			},
			"grammars": []map[string]interface{}{
				{
					"language":  "drop",
					"scopeName": "source.drop",
					"path":      "./syntaxes/drop.tmLanguage.json",
				},
			},
		},
	}
	pkgJSON, _ := json.MarshalIndent(pkg, "", "  ")
	if err := os.WriteFile(filepath.Join(extDir, "package.json"), pkgJSON, 0644); err != nil {
		return err
	}

	fmt.Println("Installed Drop syntax for VS Code")
	fmt.Println("  Restart VS Code to activate")
	fmt.Printf("  Extension: %s\n", extDir)
	return nil
}

func InstallVim() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Support both vim and neovim
	vimDir := filepath.Join(home, ".vim")
	synDir := filepath.Join(vimDir, "syntax")
	ftDir := filepath.Join(vimDir, "ftdetect")

	if err := os.MkdirAll(synDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(ftDir, 0755); err != nil {
		return err
	}

	// Write syntax file
	if err := os.WriteFile(filepath.Join(synDir, "drop.vim"), vimSyntax, 0644); err != nil {
		return err
	}

	// Write ftdetect
	ftdetect := []byte("autocmd BufRead,BufNewFile *.drop setfiletype drop\n")
	if err := os.WriteFile(filepath.Join(ftDir, "drop.vim"), ftdetect, 0644); err != nil {
		return err
	}

	fmt.Println("Installed Drop syntax for Vim")
	fmt.Printf("  Syntax:   %s\n", filepath.Join(synDir, "drop.vim"))
	fmt.Printf("  Ftdetect: %s\n", filepath.Join(ftDir, "drop.vim"))

	// Also install for Neovim if config exists
	nvimDir := neovimConfigDir(home)
	if _, err := os.Stat(nvimDir); err == nil {
		nvimSynDir := filepath.Join(nvimDir, "syntax")
		nvimFtDir := filepath.Join(nvimDir, "ftdetect")
		os.MkdirAll(nvimSynDir, 0755)
		os.MkdirAll(nvimFtDir, 0755)
		os.WriteFile(filepath.Join(nvimSynDir, "drop.vim"), vimSyntax, 0644)
		os.WriteFile(filepath.Join(nvimFtDir, "drop.vim"), ftdetect, 0644)
		fmt.Println("  Also installed for Neovim")
	}

	return nil
}

func InstallSublime() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Sublime Text packages directory varies by OS
	var pkgDir string
	switch runtime.GOOS {
	case "darwin":
		pkgDir = filepath.Join(home, "Library", "Application Support", "Sublime Text", "Packages", "Drop")
	case "linux":
		pkgDir = filepath.Join(home, ".config", "sublime-text", "Packages", "Drop")
	case "windows":
		pkgDir = filepath.Join(os.Getenv("APPDATA"), "Sublime Text", "Packages", "Drop")
	}

	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(pkgDir, "drop.tmLanguage.json"), tmGrammar, 0644); err != nil {
		return err
	}

	fmt.Println("Installed Drop syntax for Sublime Text")
	fmt.Printf("  Package: %s\n", pkgDir)
	return nil
}

func InstallZed() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var extDir string
	switch runtime.GOOS {
	case "darwin":
		extDir = filepath.Join(home, ".config", "zed", "extensions", "drop")
	case "linux":
		extDir = filepath.Join(home, ".config", "zed", "extensions", "drop")
	default:
		extDir = filepath.Join(home, ".config", "zed", "extensions", "drop")
	}

	gramDir := filepath.Join(extDir, "grammars")
	if err := os.MkdirAll(gramDir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(gramDir, "drop.tmLanguage.json"), tmGrammar, 0644); err != nil {
		return err
	}

	fmt.Println("Installed Drop syntax for Zed")
	fmt.Printf("  Extension: %s\n", extDir)
	return nil
}

func neovimConfigDir(home string) string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "nvim")
	}
	return filepath.Join(home, ".config", "nvim")
}
