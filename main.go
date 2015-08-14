package main

import (
	"fmt"
	"os"

	"github.com/calavera/gh-rel/db"
	"github.com/calavera/gh-rel/github"
	"github.com/spf13/cobra"
)

func fullMain() {
	var port uint
	var dbPath string
	var githubAuthToken string
	var adminPassword string

	cmdAdd := &cobra.Command{
		Use:   "add [owner/name]",
		Short: "Add a new project to the dashboard",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				fmt.Printf("Invalid arguments: %v\n", args)
				cmd.Usage()
				return
			}

			setup(dbPath, githubAuthToken)
			defer teardown()

			if err := addProject(args[0]); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	cmdServe := &cobra.Command{
		Use:   "serve",
		Short: "Start the web server",
		Run: func(cmd *cobra.Command, args []string) {
			setup(dbPath, githubAuthToken)
			defer teardown()
			startServer(port, adminPassword)
		},
	}
	cmdServe.Flags().UintVarP(&port, "port", "p", 8888, "port to serve the web application")
	cmdServe.Flags().StringVarP(&githubAuthToken, "auth", "a", "", "GitHub auth token")
	cmdServe.Flags().StringVarP(&adminPassword, "passwd", "", "passw0rd", "Admin password to add projects")

	rootCmd := &cobra.Command{Use: "gh-rel"}
	rootCmd.AddCommand(cmdAdd, cmdServe)
	rootCmd.PersistentFlags().StringVarP(&dbPath, "db", "d", db.DefaultPath, "path to the database file")

	rootCmd.Execute()
}

func setup(dbPath, githubAuthToken string) {
	if err := db.Open(dbPath); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	github.InitClient(githubAuthToken)
}

func teardown() {
	db.Close()
}

func main() {
	fullMain()
}
