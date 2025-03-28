package commands

import (
	"encoding/json"
	"fmt"
	"hhx/internal/config"
	"hhx/internal/models"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// collectionCmd represents the collection command
var collectionCmd = &cobra.Command{
	Use:   "collection",
	Short: "Manage collections (buckets and tables)",
	Long:  `Create, list, and manage collections of data for buckets and tables.`,
}

// collectionListCmd represents the collection list command
var collectionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all collections",
	Long:  `List all collections in the repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find repository root
		_, err := findRepoRoot()
		if err != nil {
			return err
		}

		// Load repository config
		repoConfig, err := config.LoadRepoConfig()
		if err != nil {
			return fmt.Errorf("error loading repository config: %w", err)
		}

		// Load index
		index, err := models.LoadIndex(repoConfig.IndexPath)
		if err != nil {
			return fmt.Errorf("error loading index: %w", err)
		}

		// Get all collections
		collections := index.GetCollections()

		// Display collections
		if len(collections) == 0 {
			fmt.Println("No collections found.")
			fmt.Println("Use 'hhx collection create' to create a new collection.")
			return nil
		}

		fmt.Println("Collections:")
		fmt.Println("-------------")

		for _, collection := range collections {
			if index.DefaultCollection == collection.Name {
				color.Green("* %s (%s) - %s [DEFAULT]\n", collection.Name, collection.Type, collection.Path)
			} else {
				fmt.Printf("  %s (%s) - %s\n", collection.Name, collection.Type, collection.Path)
			}

			// Print schema for tables
			if collection.Type == models.CollectionTypeTable && collection.Schema != nil {
				fmt.Println("  Schema:")
				for _, column := range collection.Schema.Columns {
					var attributes []string
					if column.PrimaryKey {
						attributes = append(attributes, "PRIMARY KEY")
					}
					if column.Nullable {
						attributes = append(attributes, "NULL")
					}
					if len(attributes) > 0 {
						fmt.Printf("    - %s (%s) [%s]\n", column.Name, column.Type, strings.Join(attributes, ", "))
					} else {
						fmt.Printf("    - %s (%s)\n", column.Name, column.Type)
					}
				}
			}

			fmt.Println()
		}

		return nil
	},
}

// collectionCreateCmd represents the collection create command
var collectionCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new collection",
	Long:  `Create a new collection (bucket or table) in the repository.`,
	Example: `  hhx collection create my-models --type=bucket --path=models/
  hhx collection create experiment-results --type=table --schema-file=schema.json
  hhx collection create metrics --type=table --columns="id:string:pk,timestamp:datetime,value:float"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Get flags
		collType, _ := cmd.Flags().GetString("type")
		path, _ := cmd.Flags().GetString("path")
		schemaFile, _ := cmd.Flags().GetString("schema-file")
		columns, _ := cmd.Flags().GetString("columns")
		setDefault, _ := cmd.Flags().GetBool("default")

		// Validate collection type
		if collType != "bucket" && collType != "table" {
			return fmt.Errorf("invalid collection type: %s (must be 'bucket' or 'table')", collType)
		}

		// If path is not specified, use name
		if path == "" {
			path = name
		}

		// Create collection
		collection := &models.Collection{
			Name: name,
			Type: models.CollectionType(collType),
			Path: path,
		}

		// Parse schema if this is a table
		if collType == "table" {
			schema := &models.Schema{
				Columns: []*models.Column{},
			}

			// Parse schema from file if provided
			if schemaFile != "" {
				data, err := os.ReadFile(schemaFile)
				if err != nil {
					return fmt.Errorf("error reading schema file: %w", err)
				}

				if err := json.Unmarshal(data, schema); err != nil {
					return fmt.Errorf("error parsing schema file: %w", err)
				}
			} else if columns != "" {
				// Parse columns format: "name:type[:pk][:null],name2:type2[:pk][:null]"
				columnDefs := strings.Split(columns, ",")
				for _, colDef := range columnDefs {
					parts := strings.Split(colDef, ":")
					if len(parts) < 2 {
						return fmt.Errorf("invalid column definition: %s", colDef)
					}

					col := &models.Column{
						Name: parts[0],
						Type: parts[1],
					}

					// Check for optional attributes
					for i := 2; i < len(parts); i++ {
						switch parts[i] {
						case "pk":
							col.PrimaryKey = true
						case "null":
							col.Nullable = true
						default:
							return fmt.Errorf("unknown column attribute: %s", parts[i])
						}
					}

					schema.Columns = append(schema.Columns, col)
				}
			} else {
				return fmt.Errorf("schema is required for table collections. Use --schema-file or --columns")
			}

			collection.Schema = schema
		}

		// Find repository root
		_, err := findRepoRoot()
		if err != nil {
			return err
		}

		// Load repository config
		repoConfig, err := config.LoadRepoConfig()
		if err != nil {
			return fmt.Errorf("error loading repository config: %w", err)
		}

		// Load index
		index, err := models.LoadIndex(repoConfig.IndexPath)
		if err != nil {
			return fmt.Errorf("error loading index: %w", err)
		}

		// Add collection
		if err := index.AddCollection(collection); err != nil {
			return fmt.Errorf("error creating collection: %w", err)
		}

		// Set as default if requested
		if setDefault {
			if err := index.SetDefaultCollection(name); err != nil {
				return fmt.Errorf("error setting default collection: %w", err)
			}
		}

		// Save index
		if err := index.Save(repoConfig.IndexPath); err != nil {
			return fmt.Errorf("error saving index: %w", err)
		}

		fmt.Printf("Collection '%s' created successfully.\n", name)
		if setDefault {
			fmt.Printf("Set as default collection.\n")
		}
		return nil
	},
}

// collectionRemoveCmd represents the collection remove command
var collectionRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a collection",
	Long:  `Remove a collection from the repository.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Find repository root
		_, err := findRepoRoot()
		if err != nil {
			return err
		}

		// Load repository config
		repoConfig, err := config.LoadRepoConfig()
		if err != nil {
			return fmt.Errorf("error loading repository config: %w", err)
		}

		// Load index
		index, err := models.LoadIndex(repoConfig.IndexPath)
		if err != nil {
			return fmt.Errorf("error loading index: %w", err)
		}

		// Get the collection first to ensure it exists
		if _, err := index.GetCollection(name); err != nil {
			return fmt.Errorf("collection not found: %s", name)
		}

		// Remove collection
		if err := index.RemoveCollection(name); err != nil {
			return fmt.Errorf("error removing collection: %w", err)
		}

		// Save index
		if err := index.Save(repoConfig.IndexPath); err != nil {
			return fmt.Errorf("error saving index: %w", err)
		}

		fmt.Printf("Collection '%s' removed successfully.\n", name)
		return nil
	},
}

// collectionSetDefaultCmd represents the collection set-default command
var collectionSetDefaultCmd = &cobra.Command{
	Use:   "set-default [name]",
	Short: "Set the default collection",
	Long:  `Set the default collection to use when no collection is specified.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Find repository root
		_, err := findRepoRoot()
		if err != nil {
			return err
		}

		// Load repository config
		repoConfig, err := config.LoadRepoConfig()
		if err != nil {
			return fmt.Errorf("error loading repository config: %w", err)
		}

		// Load index
		index, err := models.LoadIndex(repoConfig.IndexPath)
		if err != nil {
			return fmt.Errorf("error loading index: %w", err)
		}

		// Set default collection
		if err := index.SetDefaultCollection(name); err != nil {
			return fmt.Errorf("error setting default collection: %w", err)
		}

		// Save index
		if err := index.Save(repoConfig.IndexPath); err != nil {
			return fmt.Errorf("error saving index: %w", err)
		}

		fmt.Printf("Default collection set to '%s'.\n", name)
		return nil
	},
}

// collectionShowCmd represents the collection show command
var collectionShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show details of a collection",
	Long:  `Show detailed information about a collection.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Find repository root
		_, err := findRepoRoot()
		if err != nil {
			return err
		}

		// Load repository config
		repoConfig, err := config.LoadRepoConfig()
		if err != nil {
			return fmt.Errorf("error loading repository config: %w", err)
		}

		// Load index
		index, err := models.LoadIndex(repoConfig.IndexPath)
		if err != nil {
			return fmt.Errorf("error loading index: %w", err)
		}

		// Get collection
		collection, err := index.GetCollection(name)
		if err != nil {
			return fmt.Errorf("collection not found: %s", name)
		}

		// Display collection details
		fmt.Printf("Collection: %s\n", collection.Name)
		fmt.Printf("Type: %s\n", collection.Type)
		fmt.Printf("Path: %s\n", collection.Path)

		if index.DefaultCollection == collection.Name {
			fmt.Println("Default: Yes")
		} else {
			fmt.Println("Default: No")
		}

		// Print schema for tables
		if collection.Type == models.CollectionTypeTable && collection.Schema != nil {
			fmt.Println("\nSchema:")
			for _, column := range collection.Schema.Columns {
				var attributes []string
				if column.PrimaryKey {
					attributes = append(attributes, "PRIMARY KEY")
				}
				if column.Nullable {
					attributes = append(attributes, "NULL")
				}
				if len(attributes) > 0 {
					fmt.Printf("  - %s (%s) [%s]\n", column.Name, column.Type, strings.Join(attributes, ", "))
				} else {
					fmt.Printf("  - %s (%s)\n", column.Name, column.Type)
				}
			}
		}

		// Print metadata if any
		if len(collection.Metadata) > 0 {
			fmt.Println("\nMetadata:")
			for key, value := range collection.Metadata {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}

		return nil
	},
}

func init() {
	// Add collection commands to collection command
	collectionCmd.AddCommand(collectionListCmd)
	collectionCmd.AddCommand(collectionCreateCmd)
	collectionCmd.AddCommand(collectionRemoveCmd)
	collectionCmd.AddCommand(collectionSetDefaultCmd)
	collectionCmd.AddCommand(collectionShowCmd)

	// Add flags to collection create command
	collectionCreateCmd.Flags().String("type", "", "Type of collection (bucket or table)")
	collectionCreateCmd.Flags().String("path", "", "Path within the remote (default: same as name)")
	collectionCreateCmd.Flags().String("schema-file", "", "JSON file containing schema definition (for tables)")
	collectionCreateCmd.Flags().String("columns", "", "Column definitions for tables (format: 'name:type[:pk][:null],name2:type2')")
	collectionCreateCmd.Flags().Bool("default", false, "Set as default collection")

	// Make type flag required
	collectionCreateCmd.MarkFlagRequired("type")
}
