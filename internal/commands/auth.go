package commands

import (
	"fmt"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
	"hhx/internal/api"
	"hhx/internal/config"
	"hhx/internal/models"
	"os"
	"syscall"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to the remote server",
	Long:  "Authenticate with the remote server to enable pushing files and other operations",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalConfigDir, err := config.GetGlobalConfigDir()
		if err != nil {
			fmt.Println("error getting global config directory:", err)
			return nil
		}

		if err := os.MkdirAll(globalConfigDir, 0755); err != nil {
			fmt.Println("error creating global config directory:", err)
			return nil
		}

		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			fmt.Println("error loading global config:", err)
			return nil
		}

		tokenStore := models.NewTokenStore(globalConfigDir)
		serverURL := globalConfig.ServerURL
		if serverURL == "" {
			fmt.Println("Error: server URL not configured")
			return nil
		}

		var email string
		fmt.Print("Email: ")
		_, err = fmt.Scanln(&email)
		if err != nil {
			fmt.Println("error reading email:", err)
			return nil
		}

		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(uintptr(syscall.Stdin))
		if err != nil {
			fmt.Println("error reading password:", err)
			return nil
		}
		fmt.Println() // Add a newline after password input

		password := string(passwordBytes)
		client := api.NewClient(serverURL, tokenStore)
		authResult, err := client.Login(email, password)
		if err != nil {
			fmt.Println("login failed:", err)
			return nil
		}

		globalConfig.AuthToken = authResult.Token
		globalConfig.UserID = authResult.UserID
		globalConfig.Email = authResult.Email

		if err := config.SaveGlobalConfig(globalConfig); err != nil {
			fmt.Println("error saving global config:", err)
			return nil
		}

		fmt.Printf("Successfully logged in as %s\n", email)
		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from the remote server",
	Long:  "Remove saved authentication credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalConfigDir, err := config.GetGlobalConfigDir()
		if err != nil {
			fmt.Println("error getting global config directory:", err)
			return nil
		}

		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			fmt.Println("error loading global config:", err)
			return nil
		}

		tokenStore := models.NewTokenStore(globalConfigDir)
		client := api.NewClient(globalConfig.ServerURL, tokenStore)
		if err := client.Logout(); err != nil {
			fmt.Println("error during logout:", err)
			return nil
		}

		globalConfig.AuthToken = ""
		globalConfig.UserID = ""
		globalConfig.Email = ""

		if err := config.SaveGlobalConfig(globalConfig); err != nil {
			fmt.Println("error saving global config:", err)
			return nil
		}

		fmt.Println("Successfully logged out")
		return nil
	},
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current user information",
	Long:  "Display information about the currently logged in user",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalConfigDir, err := config.GetGlobalConfigDir()
		if err != nil {
			fmt.Println("error getting global config directory:", err)
			return nil
		}

		if _, err := os.Stat(globalConfigDir); os.IsNotExist(err) {
			fmt.Println("You are not logged in")
			return nil
		}

		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			fmt.Println("error loading global config:", err)
			return nil
		}

		tokenStore := models.NewTokenStore(globalConfigDir)
		token, tokenErr := tokenStore.GetToken()
		if globalConfig.AuthToken == "" && (tokenErr != nil || token == "") {
			fmt.Println("You are not logged in")
			return nil
		}

		// If token is in store but not in config, update config
		if globalConfig.AuthToken == "" && token != "" {
			globalConfig.AuthToken = token
			// We don't have user info, but token is available
			fmt.Println("You are logged in, but user details are not available")
			fmt.Printf("Server: %s\n", globalConfig.ServerURL)
			return nil
		}

		fmt.Printf("Logged in as: %s\n", globalConfig.Email)
		fmt.Printf("User ID: %s\n", globalConfig.UserID)
		fmt.Printf("Server: %s\n", globalConfig.ServerURL)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(whoamiCmd)
}
