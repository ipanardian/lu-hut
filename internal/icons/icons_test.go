package icons

import (
	"testing"
)

func TestForFile_Directory(t *testing.T) {
	tests := []struct {
		name     string
		dirName  string
		isEmpty  bool
		expected rune
	}{
		{"git dir", ".git", false, iconFolderGit},
		{"github dir", ".github", false, iconFolderGithub},
		{"node_modules", "node_modules", false, iconFolderNpm},
		{"src dir", "src", false, rune(0xf08de)},
		{"generic non-empty dir", "myproject", false, iconFolder},
		{"empty dir fallback", "emptydir", true, iconFolderOpen},
		{"config dir", ".config", false, iconFolderConfig},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ForFile(tt.dirName, true, tt.isEmpty)
			if got != tt.expected {
				t.Errorf("ForFile(%q, dir, empty=%v) = %U, want %U", tt.dirName, tt.isEmpty, got, tt.expected)
			}
		})
	}
}

func TestForFile_ExactFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected rune
	}{
		{"Dockerfile", "Dockerfile", iconDocker},
		{"go.mod", "go.mod", iconGo},
		{"go.sum", "go.sum", iconGo},
		{"Cargo.toml", "Cargo.toml", iconRust},
		{"Cargo.lock", "Cargo.lock", iconRust},
		{".gitignore", ".gitignore", iconGit},
		{".gitconfig", ".gitconfig", iconGit},
		{"Makefile", "Makefile", iconMake},
		{"LICENSE", "LICENSE", iconLicense},
		{"README.md", "README.md", iconReadme},
		{"package.json", "package.json", iconNodejs},
		{"docker-compose.yml", "docker-compose.yml", iconDocker},
		{".zshrc", ".zshrc", iconShell},
		{".bashrc", ".bashrc", iconShell},
		{"Gemfile", "Gemfile", iconRuby},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ForFile(tt.filename, false, false)
			if got != tt.expected {
				t.Errorf("ForFile(%q) = %U, want %U", tt.filename, got, tt.expected)
			}
		})
	}
}

func TestForFile_Extension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected rune
	}{
		{"Go file", "main.go", iconGo},
		{"Rust file", "lib.rs", iconRust},
		{"Python file", "script.py", iconPython},
		{"JS file", "app.js", iconJavaScript},
		{"TS file", "types.ts", iconTypeScript},
		{"JSX file", "component.jsx", iconReact},
		{"TSX file", "page.tsx", iconReact},
		{"HTML file", "index.html", iconHTML5},
		{"CSS file", "style.css", iconCSS3},
		{"SCSS file", "theme.scss", iconSass},
		{"JSON file", "config.json", iconJSON},
		{"YAML file", "ci.yaml", iconYAML},
		{"YML file", "ci.yml", iconYAML},
		{"TOML file", "config.toml", iconToml},
		{"Markdown file", "notes.md", iconMarkdown},
		{"C file", "main.c", iconC},
		{"C++ file", "main.cpp", iconCpp},
		{"Java file", "Main.java", iconJava},
		{"Ruby file", "app.rb", iconRuby},
		{"PHP file", "index.php", iconPHP},
		{"Shell script", "deploy.sh", iconShellCmd},
		{"PNG image", "logo.png", iconImage},
		{"MP4 video", "demo.mp4", iconVideo},
		{"MP3 audio", "song.mp3", iconAudio},
		{"ZIP archive", "dist.zip", iconCompressed},
		{"tar.gz", "archive.gz", iconCompressed},
		{"SQLite db", "data.sqlite", iconSqlite},
		{"SQL file", "schema.sql", iconDatabase},
		{"PDF file", "doc.pdf", iconDocument},
		{"SVG file", "logo.svg", iconVector},
		{"Font TTF", "font.ttf", iconFont},
		{"Vue file", "App.vue", iconVue},
		{"Svelte file", "App.svelte", iconSvelte},
		{"Dart file", "main.dart", iconDart},
		{"Kotlin file", "Main.kt", iconKotlin},
		{"Swift file", "App.swift", iconSwift},
		{"Scala file", "Main.scala", iconScala},
		{"Lua file", "script.lua", iconLua},
		{"Haskell file", "Main.hs", iconHaskell},
		{"Elixir file", "app.ex", iconElixir},
		{"Terraform file", "main.tf", iconTerraform},
		{"GraphQL file", "schema.graphql", iconGraphql},
		{"Proto file", "service.proto", iconProto},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ForFile(tt.filename, false, false)
			if got != tt.expected {
				t.Errorf("ForFile(%q) = %U, want %U", tt.filename, got, tt.expected)
			}
		})
	}
}

func TestForFile_Fallbacks(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected rune
	}{
		{"no extension file", "Makefile_custom", iconFileUnknown},
		{"unknown extension", "file.xyz123", iconFile},
		{"uppercase extension", "file.GO", iconGo},
		{"mixed case extension", "file.Rs", iconRust},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ForFile(tt.filename, false, false)
			if got != tt.expected {
				t.Errorf("ForFile(%q) = %U, want %U", tt.filename, got, tt.expected)
			}
		})
	}
}

func TestShouldShow(t *testing.T) {
	tests := []struct {
		name     string
		mode     Mode
		wantBool bool
	}{
		{"always", ModeAlways, true},
		{"never", ModeNever, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldShow(tt.mode)
			if got != tt.wantBool {
				t.Errorf("ShouldShow(%q) = %v, want %v", tt.mode, got, tt.wantBool)
			}
		})
	}
}

func TestPrefix(t *testing.T) {
	t.Run("never returns empty string", func(t *testing.T) {
		p := Prefix("main.go", false, false, ModeNever)
		if p != "" {
			t.Errorf("Prefix with ModeNever = %q, want empty", p)
		}
	})

	t.Run("always returns icon+space", func(t *testing.T) {
		p := Prefix("main.go", false, false, ModeAlways)
		if len([]rune(p)) != 2 {
			t.Errorf("Prefix with ModeAlways = %q (%d runes), want 2 runes (icon+space)", p, len([]rune(p)))
		}
		if []rune(p)[1] != ' ' {
			t.Errorf("Prefix second rune = %U, want space", []rune(p)[1])
		}
	})
}

func TestIconWidth(t *testing.T) {
	if IconWidth != 2 {
		t.Errorf("IconWidth = %d, want 2", IconWidth)
	}
}
