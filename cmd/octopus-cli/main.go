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

	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
