package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var osscriptsCmd = &cobra.Command{
	Use:   "osscripts",
	Short: "Run OS-provided scripts and recipes",
	Long:  `Browse and execute just recipes and shell scripts provided by your OS distribution.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return osscriptsMenu()
	},
}

func init() {
	rootCmd.AddCommand(osscriptsCmd)
}

// recipeFile represents a discovered recipe or script
type recipeFile struct {
	justfile    string // path to the justfile
	recipeName  string // name of the recipe within the justfile
	displayName string // friendly name for display
	fileType    string // "just" or "bash"
	scriptPath  string // for bash scripts, the full path
}

// osscriptsAvailable checks if any OS scripts are available
func osscriptsAvailable() bool {
	recipes := discoverOSScripts()
	return len(recipes) > 0
}

func discoverOSScripts() []recipeFile {
	var recipes []recipeFile

	// Check if just command is available
	if _, err := exec.LookPath("just"); err != nil {
		return recipes
	}

	// Scan /usr/share/*/just/ directories
	shareDir := "/usr/share"
	entries, err := os.ReadDir(shareDir)
	if err != nil {
		return recipes
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		justDir := filepath.Join(shareDir, entry.Name(), "just")
		if _, err := os.Stat(justDir); err != nil {
			continue
		}

		// Walk the just directory to find .just and .sh files
		filepath.Walk(justDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}

			ext := filepath.Ext(path)
			basename := filepath.Base(path)

			if ext == ".just" || basename == "justfile" || basename == "Justfile" {
				// Parse justfile to extract individual recipes
				justRecipes := parseJustfileRecipes(path)

				for _, recipeName := range justRecipes {
					// Skip the default recipe
					if recipeName == "default" {
						continue
					}

					// For just recipes, display only the recipe name
					displayName := recipeName
					recipes = append(recipes, recipeFile{
						justfile:    path,
						recipeName:  recipeName,
						displayName: displayName,
						fileType:    "just",
					})
				}
			} else if ext == ".sh" {
				// Bash scripts are executed directly
				relPath, _ := filepath.Rel(justDir, path)
				displayName := fmt.Sprintf("%s/%s", entry.Name(), relPath)

				recipes = append(recipes, recipeFile{
					scriptPath:  path,
					displayName: displayName,
					fileType:    "bash",
				})
			}
			return nil
		})
	}

	return recipes
}

// parseJustfileRecipes reads a justfile and extracts recipe names
func parseJustfileRecipes(justfilePath string) []string {
	var recipes []string

	file, err := os.Open(justfilePath)
	if err != nil {
		return recipes
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Recipe definitions start with a name followed by optional parameters and a colon
		// Format: recipe-name param1 param2: or just recipe-name:
		if strings.Contains(line, ":") && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			// This looks like a recipe definition
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 0 {
				recipeLine := strings.TrimSpace(parts[0])
				// Extract just the recipe name (first word)
				recipeWords := strings.Fields(recipeLine)
				if len(recipeWords) > 0 {
					recipeName := recipeWords[0]
					// Skip if it starts with @ (hidden recipe) or _ (private recipe)
					if !strings.HasPrefix(recipeName, "@") && !strings.HasPrefix(recipeName, "_") {
						recipes = append(recipes, recipeName)
					}
				}
			}
		}
	}

	return recipes
}

func osscriptsMenu() error {
	recipes := discoverOSScripts()
	if len(recipes) == 0 {
		return fmt.Errorf("no OS scripts found in /usr/share/*/just/")
	}

	// Build options from discovered recipes
	opts := make([]huh.Option[string], 0, len(recipes))
	recipeMap := make(map[string]recipeFile)

	for i, recipe := range recipes {
		key := fmt.Sprintf("recipe_%d", i)
		icon := "📜"
		if recipe.fileType == "bash" {
			icon = "🐚"
		}
		opts = append(opts, huh.NewOption(fmt.Sprintf("%s %s", icon, recipe.displayName), key))
		recipeMap[key] = recipe
	}

	var selected string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a script or recipe to run").
				Description("Choose from available OS-provided scripts").
				Options(opts...).
				Value(&selected),
		),
	).Run(); err != nil {
		return err
	}

	recipe, ok := recipeMap[selected]
	if !ok {
		return fmt.Errorf("recipe not found")
	}

	// Execute the selected recipe
	var cmd *exec.Cmd
	if recipe.fileType == "bash" {
		cmd = exec.Command("bash", recipe.scriptPath)
	} else {
		// Execute the specific recipe from the justfile
		cmd = exec.Command("just", "-f", recipe.justfile, recipe.recipeName)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
