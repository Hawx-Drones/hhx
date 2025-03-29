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
			return err
		}

		if err := os.MkdirAll(globalConfigDir, 0755); err != nil {
			return fmt.Errorf("error creating global config directory: %w", err)
		}

		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("error loading global config: %w", err)
		}

		tokenStore := models.NewTokenStore(globalConfigDir)
		serverURL := globalConfig.ServerURL
		if serverURL == "" {
			return fmt.Errorf("server URL not configured")
		}

		var email string
		fmt.Print("Email: ")
		fmt.Scanln(&email)

		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(uintptr(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("error reading password: %w", err)
		}
		fmt.Println() // Add a newline after password input

		password := string(passwordBytes)
		client := api.NewClient(serverURL, tokenStore)
		authResult, err := client.Login(email, password)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		globalConfig.AuthToken = authResult.Token
		globalConfig.UserID = authResult.UserID
		globalConfig.Email = authResult.Email

		if err := config.SaveGlobalConfig(globalConfig); err != nil {
			return fmt.Errorf("error saving global config: %w", err)
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
			return err
		}

		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("error loading global config: %w", err)
		}

		tokenStore := models.NewTokenStore(globalConfigDir)
		client := api.NewClient(globalConfig.ServerURL, tokenStore)
		if err := client.Logout(); err != nil {
			return fmt.Errorf("error during logout: %w", err)
		}

		globalConfig.AuthToken = ""
		globalConfig.UserID = ""
		globalConfig.Email = ""

		if err := config.SaveGlobalConfig(globalConfig); err != nil {
			return fmt.Errorf("error saving global config: %w", err)
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
			return err
		}

		if _, err := os.Stat(globalConfigDir); os.IsNotExist(err) {
			fmt.Println("You are not logged in")
			return nil
		}

		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("error loading global config: %w", err)
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
