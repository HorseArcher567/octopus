package main

import (
	"fmt"
	"os"

	"github.com/HorseArcher567/octopus/cmd/octopus-cli/generator"

	"github.com/spf13/cobra"
)

var version = "1.0.0"

var rootCmd = &cobra.Command{
	Use:   "octopus-cli",
	Short: "Octopus RPC framework code generator",
	Long:  `A code generation tool for Octopus RPC framework`,
}

var newCmd = &cobra.Command{
	Use:   "new [service-name]",
	Short: "Create a new RPC service",
	Long:  `Create a new RPC service with standard project structure`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]

		// 获取标志
		module, _ := cmd.Flags().GetString("module")
		dir, _ := cmd.Flags().GetString("dir")

		// 生成项目
		if err := generator.Generate(serviceName, module, dir); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n✅ Service '%s' created successfully!\n", serviceName)
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  cd %s\n", serviceName)
		fmt.Printf("  make proto\n")
		fmt.Printf("  make run\n")
	},
}

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new monorepo project",
	Long:  `Initialize a new monorepo project with standard structure for multiple services`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]

		// Get flags
		module, _ := cmd.Flags().GetString("module")
		dir, _ := cmd.Flags().GetString("dir")

		// Generate monorepo project
		if err := generator.InitMonorepo(projectName, module, dir); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n✅ Monorepo project '%s' initialized successfully!\n", projectName)
	},
}

var addCmd = &cobra.Command{
	Use:   "add [app-name]",
	Short: "Add a new application to monorepo",
	Long:  `Add a new application to an existing monorepo project`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]

		// Get flags
		port, _ := cmd.Flags().GetInt("port")
		dir, _ := cmd.Flags().GetString("dir")

		// Add app to monorepo
		if err := generator.AddApp(appName, port, dir); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n✅ Application '%s' added successfully!\n", appName)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("octopus-cli version %s\n", version)
	},
}

func init() {
	// new 命令的标志
	newCmd.Flags().StringP("module", "m", "", "Go module name (default: service-name)")
	newCmd.Flags().StringP("dir", "d", ".", "Output directory")

	// init 命令的标志
	initCmd.Flags().StringP("module", "m", "", "Go module name (default: project-name)")
	initCmd.Flags().StringP("dir", "d", ".", "Output directory")

	// add 命令的标志
	addCmd.Flags().IntP("port", "p", 9000, "Service port")
	addCmd.Flags().StringP("dir", "d", ".", "Monorepo root directory")

	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
