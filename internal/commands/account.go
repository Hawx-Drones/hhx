package commands

import (
	"bufio"
	"fmt"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
	"hhx/internal/api"
	"hhx/internal/config"
	"hhx/internal/models"
	"os"
	"syscall"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Manage your account",
	Long:  "Manage your account, including login, logout, registration, and viewing account details",
}

var accountCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new account",
	Long:  "Register a new account with the remote server",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalConfigDir, err := config.GetGlobalConfigDir()
		if err != nil {
			fmt.Println("Error getting global config directory:", err)
			return nil
		}

		if err := os.MkdirAll(globalConfigDir, 0755); err != nil {
			fmt.Println("Error creating global config directory:", err)
			return nil
		}

		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			fmt.Println("Error loading global config:", err)
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
			fmt.Println("Error reading email:", err)
			return nil
		}

		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(uintptr(syscall.Stdin))
		if err != nil {
			fmt.Println("Error reading password:", err)
			return nil
		}
		fmt.Println() // Add a newline after password input

		fmt.Print("Confirm password: ")
		confirmPasswordBytes, err := term.ReadPassword(uintptr(syscall.Stdin))
		if err != nil {
			fmt.Println("Error reading password confirmation:", err)
			return nil
		}
		fmt.Println() // Add a newline after password input

		password := string(passwordBytes)
		confirmPassword := string(confirmPasswordBytes)

		if password != confirmPassword {
			fmt.Println("Error: Passwords do not match")
			return nil
		}

		name := ""
		fmt.Print("Name (optional): ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			name = scanner.Text()
		}

		phone := ""
		fmt.Print("Phone (optional): ")
		if scanner.Scan() {
			phone = scanner.Text()
		}

		client := api.NewClient(serverURL, tokenStore)
		authResult, err := client.Register(email, password, name, phone)
		if err != nil {
			fmt.Println("Account creation failed:", err)
			return nil
		}

		globalConfig.AuthToken = authResult.Token
		globalConfig.UserID = authResult.UserID
		globalConfig.Email = authResult.Email

		if err := config.SaveGlobalConfig(globalConfig); err != nil {
			fmt.Println("Error saving global config:", err)
			return nil
		}

		fmt.Printf("Account successfully created for %s\n", email)
		fmt.Println("You are now logged in")
		return nil
	},
}

var accountLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to your account",
	Long:  "Authenticate with the remote server to enable pushing files and other operations",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalConfigDir, err := config.GetGlobalConfigDir()
		if err != nil {
			fmt.Println("Error getting global config directory:", err)
			return nil
		}

		if err := os.MkdirAll(globalConfigDir, 0755); err != nil {
			fmt.Println("Error creating global config directory:", err)
			return nil
		}

		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			fmt.Println("Error loading global config:", err)
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
			fmt.Println("Error reading email:", err)
			return nil
		}

		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(uintptr(syscall.Stdin))
		if err != nil {
			fmt.Println("Error reading password:", err)
			return nil
		}
		fmt.Println() // Add a newline after password input

		password := string(passwordBytes)
		client := api.NewClient(serverURL, tokenStore)
		authResult, err := client.Login(email, password)
		if err != nil {
			fmt.Println("Login failed:", err)
			return nil
		}

		globalConfig.AuthToken = authResult.Token
		globalConfig.UserID = authResult.UserID
		globalConfig.Email = authResult.Email

		if err := config.SaveGlobalConfig(globalConfig); err != nil {
			fmt.Println("Error saving global config:", err)
			return nil
		}

		fmt.Printf("Successfully logged in as %s\n", email)
		return nil
	},
}

var accountLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from your account",
	Long:  "Remove saved authentication credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalConfigDir, err := config.GetGlobalConfigDir()
		if err != nil {
			fmt.Println("Error getting global config directory:", err)
			return nil
		}

		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			fmt.Println("Error loading global config:", err)
			return nil
		}

		tokenStore := models.NewTokenStore(globalConfigDir)
		client := api.NewClient(globalConfig.ServerURL, tokenStore)
		if err := client.Logout(); err != nil {
			fmt.Println("Error during logout:", err)
			return nil
		}

		globalConfig.AuthToken = ""
		globalConfig.UserID = ""
		globalConfig.Email = ""

		if err := config.SaveGlobalConfig(globalConfig); err != nil {
			fmt.Println("Error saving global config:", err)
			return nil
		}

		fmt.Println("Successfully logged out")
		return nil
	},
}

var accountInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show current account information",
	Long:  "Display basic information about the currently logged in user",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalConfigDir, err := config.GetGlobalConfigDir()
		if err != nil {
			fmt.Println("Error getting global config directory:", err)
			return nil
		}

		if _, err := os.Stat(globalConfigDir); os.IsNotExist(err) {
			fmt.Println("You are not logged in")
			return nil
		}

		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			fmt.Println("Error loading global config:", err)
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

var accountDetailsCmd = &cobra.Command{
	Use:   "details",
	Short: "Show detailed account information",
	Long:  "Display comprehensive information about the current account including subscription status",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalConfigDir, err := config.GetGlobalConfigDir()
		if err != nil {
			fmt.Println("Error getting global config directory:", err)
			return nil
		}

		if _, err := os.Stat(globalConfigDir); os.IsNotExist(err) {
			fmt.Println("You are not logged in")
			return nil
		}

		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			fmt.Println("Error loading global config:", err)
			return nil
		}

		tokenStore := models.NewTokenStore(globalConfigDir)
		token, tokenErr := tokenStore.GetToken()
		if globalConfig.AuthToken == "" && (tokenErr != nil || token == "") {
			fmt.Println("You are not logged in")
			return nil
		}

		client := api.NewClient(globalConfig.ServerURL, tokenStore)
		userDetails, err := client.GetUserDetails()
		if err != nil {
			fmt.Println("Error fetching account details:", err)
			return nil
		}

		fmt.Println("Account Details")
		fmt.Println("---------------")
		fmt.Printf("Email: %s\n", userDetails.Email)
		fmt.Printf("User ID: %s\n", userDetails.UserID)
		fmt.Printf("Name: %s\n", userDetails.Name)
		fmt.Printf("Phone: %s\n", userDetails.Phone)
		fmt.Println()
		fmt.Println("Subscription Information")
		fmt.Println("------------------------")
		fmt.Printf("Plan: %s\n", userDetails.Subscription.Plan)
		fmt.Printf("Status: %s\n", userDetails.Subscription.Status)
		fmt.Printf("Renewal Date: %s\n", userDetails.Subscription.RenewalDate)
		fmt.Printf("Member Since: %s\n", userDetails.Subscription.MemberSince)

		return nil
	},
}

var accountUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update account information",
	Long:  "Update your account details such as name, email, phone, or password",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalConfigDir, err := config.GetGlobalConfigDir()
		if err != nil {
			fmt.Println("Error getting global config directory:", err)
			return nil
		}

		if _, err := os.Stat(globalConfigDir); os.IsNotExist(err) {
			fmt.Println("You are not logged in")
			return nil
		}

		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			fmt.Println("Error loading global config:", err)
			return nil
		}

		tokenStore := models.NewTokenStore(globalConfigDir)
		token, tokenErr := tokenStore.GetToken()
		if globalConfig.AuthToken == "" && (tokenErr != nil || token == "") {
			fmt.Println("You are not logged in")
			return nil
		}

		client := api.NewClient(globalConfig.ServerURL, tokenStore)

		userDetails, err := client.GetUserDetails()
		if err != nil {
			fmt.Println("Error fetching account details:", err)
			return nil
		}

		// Determine what to update based on flags
		updateName := cmd.Flags().Changed("name")
		updateEmail := cmd.Flags().Changed("email")
		updatePhone := cmd.Flags().Changed("phone")
		updatePassword := cmd.Flags().Changed("password")

		// If no specific flags are set, update everything interactively
		if !updateName && !updateEmail && !updatePhone && !updatePassword {
			updateName = true
			updateEmail = true
			updatePhone = true
			updatePassword = true
		}

		updateRequest := &models.UserDetails{}
		if updateName {
			nameStr, _ := cmd.Flags().GetString("name")
			if nameStr != "" {
				updateRequest.Name = nameStr
			} else {
				fmt.Printf("Name [%s]: ", userDetails.Name)
				var name string
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					name = scanner.Text()
				}
				if name != "" {
					updateRequest.Name = name
				} else {
					updateRequest.Name = userDetails.Name
				}
			}
		}

		if updateEmail {
			emailStr, _ := cmd.Flags().GetString("email")
			if emailStr != "" {
				updateRequest.Email = emailStr
			} else {
				fmt.Printf("Email [%s]: ", userDetails.Email)
				var email string
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					email = scanner.Text()
				}
				if email != "" {
					updateRequest.Email = email
				} else {
					updateRequest.Email = userDetails.Email
				}
			}
		}

		if updatePhone {
			phoneStr, _ := cmd.Flags().GetString("phone")
			if phoneStr != "" {
				updateRequest.Phone = phoneStr
			} else {
				fmt.Printf("Phone [%s]: ", userDetails.Phone)
				var phone string
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					phone = scanner.Text()
				}
				if phone != "" {
					updateRequest.Phone = phone
				} else {
					updateRequest.Phone = userDetails.Phone
				}
			}
		}

		if updatePassword {
			fmt.Print("New Password (leave blank to keep current): ")
			passwordBytes, err := term.ReadPassword(uintptr(syscall.Stdin))
			if err != nil {
				fmt.Println("\nError reading password:", err)
				return nil
			}
			fmt.Println() // Add newline after password input

			password := string(passwordBytes)
			if password != "" {
				fmt.Print("Confirm Password: ")
				confirmBytes, err := term.ReadPassword(uintptr(syscall.Stdin))
				if err != nil {
					fmt.Println("\nError reading password confirmation:", err)
					return nil
				}
				fmt.Println() // Add newline after password input

				confirmPassword := string(confirmBytes)
				if password != confirmPassword {
					fmt.Println("Error: Passwords do not match")
					return nil
				}

				updateRequest.Password = password
			}
		}

		result, err := client.UpdateAccount(updateRequest)
		if err != nil {
			fmt.Println("Error updating account:", err)
			return nil
		}

		if updateEmail && updateRequest.Email != "" && updateRequest.Email != userDetails.Email {
			globalConfig.Email = result.Email
			if err := config.SaveGlobalConfig(globalConfig); err != nil {
				fmt.Println("Error saving updated email to global config:", err)
			}
		}

		fmt.Println("Account successfully updated")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(accountCmd)

	accountCmd.AddCommand(accountCreateCmd)
	accountCmd.AddCommand(accountLoginCmd)
	accountCmd.AddCommand(accountLogoutCmd)
	accountCmd.AddCommand(accountInfoCmd)
	accountCmd.AddCommand(accountDetailsCmd)
	accountCmd.AddCommand(accountUpdateCmd)

	accountUpdateCmd.Flags().String("name", "", "Update user name")
	accountUpdateCmd.Flags().String("email", "", "Update email address")
	accountUpdateCmd.Flags().String("phone", "", "Update phone number")
	accountUpdateCmd.Flags().Bool("password", false, "Update password")
}
