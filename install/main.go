package install

import (
	"bufio"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed all:app all:variants
var scaffold embed.FS

const (
	defaultFrameworkModule  = "github.com/golibry/go-web-skeleton"
	defaultFrameworkVersion = "latest"
)

type options struct {
	AppName           string
	ModulePath        string
	TargetDir         string
	DatabaseDriver    string
	MigrationsEnabled bool
	DockerEnabled     bool
	FrameworkVersion  string
	FrameworkReplace  string
	Force             bool
	SkipGoModTidy     bool
}

func Run(args []string) error {
	return run(args)
}

func run(args []string) error {
	fmt.Fprintln(os.Stdout, "Golibry web installer")

	opts, err := readOptions(args)
	if err != nil {
		return err
	}

	if err := ensureTargetDir(opts); err != nil {
		return err
	}
	if err := cleanupObsoleteGeneratedFiles(opts.TargetDir); err != nil {
		return err
	}

	replacements := buildReplacements(opts)
	if err := copyTree("app", opts.TargetDir, replacements, copyOptions{}); err != nil {
		return err
	}
	if err := copyTree(
		filepath.ToSlash(filepath.Join("variants", "database", opts.DatabaseDriver)),
		opts.TargetDir,
		replacements,
		copyOptions{SkipDocker: !opts.DockerEnabled},
	); err != nil {
		return err
	}
	migrationsVariant := "disabled"
	if opts.MigrationsEnabled {
		migrationsVariant = "enabled"
	}
	if err := copyTree(
		filepath.ToSlash(filepath.Join("variants", "migrations", migrationsVariant)),
		opts.TargetDir,
		replacements,
		copyOptions{},
	); err != nil {
		return err
	}
	scriptsVariant := "migrations_disabled"
	if opts.MigrationsEnabled {
		scriptsVariant = "migrations_enabled"
	}
	if err := copyTree(
		filepath.ToSlash(filepath.Join("variants", "scripts", opts.DatabaseDriver, scriptsVariant)),
		opts.TargetDir,
		replacements,
		copyOptions{},
	); err != nil {
		return err
	}

	if err := gofmtGeneratedFiles(opts.TargetDir); err != nil {
		return err
	}
	if err := initGoModule(opts); err != nil {
		return err
	}
	if !opts.SkipGoModTidy {
		if err := runCommand(opts.TargetDir, "go", "mod", "tidy"); err != nil {
			return err
		}
	}

	printNextSteps(opts)
	return nil
}

func readOptions(args []string) (options, error) {
	defaultTarget, err := os.Getwd()
	if err != nil {
		return options{}, err
	}
	migrationsFromEnv := envExists("MIGRATIONS_ENABLED")
	dockerFromEnv := envExists("DOCKER_DEV_SERVICES")

	opts := options{
		TargetDir:        envString("TARGET_DIR", defaultTarget),
		FrameworkVersion: envString("FRAMEWORK_VERSION", defaultFrameworkVersion),
		FrameworkReplace: strings.TrimSpace(os.Getenv("FRAMEWORK_REPLACE")),
		Force:            envBool("FORCE", false),
		SkipGoModTidy:    envBool("SKIP_GO_MOD_TIDY", false),
	}

	flags := flag.NewFlagSet("install", flag.ContinueOnError)
	flags.StringVar(
		&opts.AppName,
		"app-name",
		strings.TrimSpace(os.Getenv("APP_NAME")),
		"application name",
	)
	flags.StringVar(
		&opts.ModulePath,
		"module-path",
		strings.TrimSpace(os.Getenv("MODULE_PATH")),
		"Go module path",
	)
	flags.StringVar(&opts.TargetDir, "target-dir", opts.TargetDir, "target directory")
	flags.StringVar(
		&opts.DatabaseDriver,
		"database-driver",
		envString("DATABASE_DRIVER", os.Getenv("MIGRATIONS_DRIVER")),
		"database driver: mysql or postgres",
	)
	flags.BoolVar(
		&opts.MigrationsEnabled,
		"migrations",
		envBool("MIGRATIONS_ENABLED", true),
		"install migrations command",
	)
	flags.BoolVar(
		&opts.DockerEnabled,
		"docker",
		envBool("DOCKER_DEV_SERVICES", true),
		"install local Docker database setup",
	)
	flags.StringVar(
		&opts.FrameworkVersion,
		"framework-version",
		opts.FrameworkVersion,
		"framework module version",
	)
	flags.StringVar(
		&opts.FrameworkReplace,
		"framework-replace",
		opts.FrameworkReplace,
		"local framework replacement path for generated go.mod",
	)
	flags.BoolVar(
		&opts.Force,
		"force",
		opts.Force,
		"overwrite generated files in a non-empty target directory",
	)
	flags.BoolVar(&opts.SkipGoModTidy, "skip-go-mod-tidy", opts.SkipGoModTidy, "skip go mod tidy")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	flagsProvided := providedFlags(flags)

	opts.TargetDir, err = filepath.Abs(opts.TargetDir)
	if err != nil {
		return options{}, err
	}

	defaultAppName := filepath.Base(opts.TargetDir)
	opts.AppName = askString("Application name", opts.AppName, defaultAppName)
	appSlug := slugify(opts.AppName)
	opts.ModulePath = askString("Go module path", opts.ModulePath, "example.com/"+appSlug)
	opts.DatabaseDriver = normalizeDriver(
		askString(
			"Database driver",
			opts.DatabaseDriver,
			"mysql",
		),
	)
	if !migrationsFromEnv && !flagsProvided["migrations"] {
		opts.MigrationsEnabled = askBool("Enable migrations?", true)
	}
	if !dockerFromEnv && !flagsProvided["docker"] {
		opts.DockerEnabled = askBool("Install local Docker database?", true)
	}

	if opts.DatabaseDriver != "mysql" && opts.DatabaseDriver != "postgres" {
		return options{}, fmt.Errorf("unsupported database driver %q", opts.DatabaseDriver)
	}
	if opts.FrameworkVersion == "" {
		opts.FrameworkVersion = defaultFrameworkVersion
	}

	return opts, nil
}

func providedFlags(flags *flag.FlagSet) map[string]bool {
	provided := make(map[string]bool)
	flags.Visit(
		func(f *flag.Flag) {
			provided[f.Name] = true
		},
	)
	return provided
}

func ensureTargetDir(opts options) error {
	if err := os.MkdirAll(opts.TargetDir, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(opts.TargetDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Name() == ".git" {
			continue
		}
		if !opts.Force {
			return fmt.Errorf(
				"target directory is not empty: %s; pass -force or set FORCE=true to overwrite generated files",
				opts.TargetDir,
			)
		}
		break
	}

	return nil
}

func cleanupObsoleteGeneratedFiles(targetDir string) error {
	for _, relativePath := range []string{
		filepath.Join("scripts", ".database.env"),
		filepath.Join("scripts", ".migrations.env"),
	} {
		path := filepath.Join(targetDir, relativePath)
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	return nil
}

type copyOptions struct {
	SkipDocker bool
}

func copyTree(
	sourceRoot, targetRoot string,
	replacements map[string]string,
	opts copyOptions,
) error {
	return fs.WalkDir(
		scaffold, sourceRoot, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path == sourceRoot {
				return nil
			}

			relative := strings.TrimPrefix(path, sourceRoot+"/")
			if opts.SkipDocker && (relative == ".docker" || strings.HasPrefix(
				relative,
				".docker/",
			)) {
				if entry.IsDir() {
					return fs.SkipDir
				}
				return nil
			}

			target := filepath.Join(targetRoot, filepath.FromSlash(relative))
			if entry.IsDir() {
				return os.MkdirAll(target, 0755)
			}

			content, err := scaffold.ReadFile(path)
			if err != nil {
				return err
			}
			rendered := render(string(content), replacements)
			if strings.HasSuffix(path, ".go") {
				rendered = stripScaffoldBuildTag(rendered)
			}

			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			mode := fs.FileMode(0644)
			if strings.HasSuffix(target, ".sh") {
				mode = 0755
			}
			return os.WriteFile(target, []byte(rendered), mode)
		},
	)
}

func buildReplacements(opts options) map[string]string {
	return map[string]string{
		"{{MODULE_PATH}}": opts.ModulePath,
	}
}

func render(content string, replacements map[string]string) string {
	for placeholder, value := range replacements {
		content = strings.ReplaceAll(content, placeholder, value)
	}
	return content
}

func stripScaffoldBuildTag(content string) string {
	content = strings.TrimPrefix(content, "//go:build ignore\n\n")
	content = strings.TrimPrefix(content, "//go:build ignore\r\n\r\n")
	return content
}

func gofmtGeneratedFiles(targetDir string) error {
	files := make([]string, 0)
	for _, root := range []string{
		"cli",
		"config",
		filepath.Join("infrastructure", "database"),
		"presentation",
	} {
		rootPath := filepath.Join(targetDir, root)
		if _, err := os.Stat(rootPath); errors.Is(err, os.ErrNotExist) {
			continue
		}

		if err := filepath.WalkDir(
			rootPath, func(path string, entry os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if entry.IsDir() {
					return nil
				}
				if strings.HasSuffix(path, ".go") {
					files = append(files, path)
				}
				return nil
			},
		); err != nil {
			return err
		}
	}
	if len(files) == 0 {
		return nil
	}

	args := append([]string{"-w"}, files...)
	return runCommand(targetDir, "gofmt", args...)
}

func initGoModule(opts options) error {
	if _, err := os.Stat(filepath.Join(opts.TargetDir, "go.mod")); errors.Is(err, os.ErrNotExist) {
		if err := runCommand(opts.TargetDir, "go", "mod", "init", opts.ModulePath); err != nil {
			return err
		}
	}
	if err := syncGoWork(opts.TargetDir); err != nil {
		return err
	}

	if opts.FrameworkReplace != "" {
		if err := runCommand(
			opts.TargetDir,
			"go",
			"mod",
			"edit",
			"-require="+defaultFrameworkModule+"@v0.0.0",
		); err != nil {
			return err
		}
		return runCommand(
			opts.TargetDir,
			"go",
			"mod",
			"edit",
			"-replace="+defaultFrameworkModule+"="+opts.FrameworkReplace,
		)
	}

	return runCommand(opts.TargetDir, "go", "get", defaultFrameworkModule+"@"+opts.FrameworkVersion)
}

func syncGoWork(targetDir string) error {
	if _, err := os.Stat(filepath.Join(targetDir, "go.work")); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return runCommand(targetDir, "go", "work", "use", ".")
}

func runCommand(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func printNextSteps(opts options) {
	fmt.Printf("\nGolibry web app installed in %s\n", opts.TargetDir)
	fmt.Println("Next commands:")
	fmt.Println("  scripts/app.sh test ./...")
	fmt.Println("  scripts/app.sh build")
	fmt.Println("  scripts/app.sh run http:start")
	if opts.MigrationsEnabled {
		fmt.Println("  scripts/app.sh migrations status")
	}
}

func envString(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func envExists(name string) bool {
	_, ok := os.LookupEnv(name)
	return ok
}

func envBool(name string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(name)))
	switch value {
	case "":
		return fallback
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func askString(label, value, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}

	input := promptFromTerminal(label, fallback)
	if input == "" {
		return fallback
	}
	return input
}

func askBool(label string, fallback bool) bool {
	defaultLabel := "yes"
	if !fallback {
		defaultLabel = "no"
	}

	for {
		input := strings.ToLower(promptFromTerminal(label, defaultLabel))
		switch input {
		case "", "1", "true", "yes", "y", "on":
			return true
		case "0", "false", "no", "n", "off":
			return false
		}
		fmt.Fprintln(os.Stderr, "Please answer yes or no.")
	}
}

func promptFromTerminal(label, fallback string) string {
	reader := bufio.NewReader(os.Stdin)
	output := os.Stdout
	if runtime.GOOS != "windows" {
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err == nil {
			defer tty.Close()
			reader = bufio.NewReader(tty)
			output = tty
		}
	}

	fmt.Fprintf(output, "%s [%s]: ", label, fallback)
	_ = output.Sync()
	line, err := reader.ReadString('\n')
	if err != nil {
		return fallback
	}
	return strings.TrimSpace(line)
}

func normalizeDriver(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "postgresql", "pgx":
		return "postgres"
	default:
		return strings.ToLower(strings.TrimSpace(driver))
	}
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastDash := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && builder.Len() > 0 {
			builder.WriteRune('-')
			lastDash = true
		}
	}
	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "app"
	}
	return result
}
