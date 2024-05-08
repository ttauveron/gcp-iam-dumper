package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/ttauveron/gcp-iam-dumper/pkg/db"
	"github.com/ttauveron/gcp-iam-dumper/pkg/gcp"
	"log"
	"os"
	"path/filepath"
)

func main() {

	binaryName := filepath.Base(os.Args[0])
	var rootCmd = &cobra.Command{
		Use: binaryName,
	}

	var cmdDump = &cobra.Command{
		Use:   "dump",
		Short: "Dump IAM data into a SQLite database",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Performing load operation")

			quotaProjectId, _ := cmd.Flags().GetString("quotaProjectId")
			workspaceOrgId, _ := cmd.Flags().GetString("workspaceOrgId")
			gcpOrgId, _ := cmd.Flags().GetString("gcpOrgId")
			sqliteFile, _ := cmd.Flags().GetString("sqliteFile")

			ctx := context.Background()
			database, err := db.InitDB(sqliteFile)
			if err != nil {
				log.Fatalf("Failed to initialize database: %v", err)
			}
			defer database.Close()

			fmt.Printf("Syncing Roles...\n")
			syncRoles(ctx, database, gcpOrgId)
			fmt.Printf("Syncing GroupAndMembers...\n")
			syncGroupAndMembers(ctx, database, quotaProjectId, workspaceOrgId)
			fmt.Printf("Syncing Hierarchy...\n")
			syncHierarchy(ctx, database, gcpOrgId)
			fmt.Printf("Syncing Service Accounts...\n")
			syncServiceAccounts(ctx, database, gcpOrgId)
			fmt.Printf("Syncing Bindings...\n")
			syncBindings(ctx, database, gcpOrgId)
		},
	}
	cmdDump.Flags().StringP("quotaProjectId", "", "", "The quota project ID used for Directory API/Cloud Identity API (mandatory)")
	cmdDump.Flags().StringP("workspaceOrgId", "", "", "Workspace organization ID (mandatory)")
	cmdDump.Flags().StringP("gcpOrgId", "", "", "GCP organization ID (mandatory)")
	cmdDump.Flags().StringP("sqliteFile", "", "./database.db", "Path to the SQLite file for CSV export")
	cmdDump.MarkFlagRequired("quotaProjectId")
	cmdDump.MarkFlagRequired("workspaceOrgId")
	cmdDump.MarkFlagRequired("gcpOrgId")

	var cmdExport = &cobra.Command{
		Use:   "export",
		Short: "Export SQLite database to CSV",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Performing export operation")
			sqliteFile, _ := cmd.Flags().GetString("sqliteFile")
			exportDir, _ := cmd.Flags().GetString("exportDir")
			err := db.ListTablesAndDump(sqliteFile, exportDir)
			if err != nil {
				log.Fatalf("Failed to export tables to csv files: %v", err)
			}
		},
	}
	cmdExport.Flags().StringP("sqliteFile", "", "./database.db", "Path to the SQLite file for CSV export")
	cmdExport.Flags().StringP("exportDir", "", "./export", "Directory used to dump CSV exports")

	var cmdUpload = &cobra.Command{
		Use:   "upload",
		Short: "Upload files to GCS",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Performing upload operation")
			srcPath, _ := cmd.Flags().GetString("srcPath")
			bucketName, _ := cmd.Flags().GetString("bucketName")
			ctx := context.Background()
			err := gcp.UploadFilesToGCS(ctx, bucketName, srcPath)
			if err != nil {
				log.Fatalf("Failed to upload files to GCS: %v", err)
			}
		},
	}
	cmdUpload.Flags().StringP("bucketName", "", "", "GCS Bucket name where files are uploaded")
	cmdUpload.Flags().StringP("srcPath", "", "./export", "Path to upload, can be a file or a directory (non-recursive)")
	cmdUpload.MarkFlagRequired("bucketName")

	rootCmd.AddCommand(cmdDump, cmdExport, cmdUpload)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func syncRoles(ctx context.Context, database *sql.DB, gcpOrganizationID string) {
	roles, err := gcp.FetchAllRoles(ctx, gcpOrganizationID)
	if err != nil {
		log.Fatalf("Failed to fetch roles: %v", err)
	}
	if err := db.InsertRoles(database, roles); err != nil {
		log.Fatalf("Failed to insert roles: %v", err)
	}
}

func syncBindings(ctx context.Context, database *sql.DB, gcpOrganizationID string) {
	bindings, err := gcp.FetchAssetIAMPolicy(ctx, gcpOrganizationID)
	if err != nil {
		log.Fatalf("Error listing GCP bindings: %v", err)
	}
	if err := db.InsertResourceIAMPermission(database, bindings); err != nil {
		log.Fatalf("Failed to insert bindings: %v", err)
	}
}

func syncServiceAccounts(ctx context.Context, database *sql.DB, gcpOrganizationID string) {
	serviceAccounts, err := gcp.FetchServiceAccounts(ctx, gcpOrganizationID)
	if err != nil {
		log.Fatalf("Error listing GCP service accounts: %v", err)
	}
	if err := db.InsertPrincipals(database, serviceAccounts); err != nil {
		log.Fatalf("Failed to insert service accounts: %v", err)
	}
}

func syncGroupAndMembers(ctx context.Context, database *sql.DB, projectId string, organizationID string) {
	users, err := gcp.FetchUsers(ctx, projectId, organizationID)
	if err != nil {
		log.Fatalf("Error listing Users: %v", err)
	}
	if err := db.InsertPrincipals(database, users); err != nil {
		log.Fatalf("Failed to insert users: %v", err)
	}

	groups, err := gcp.FetchGroups(ctx, projectId, organizationID)
	if err := db.InsertPrincipals(database, groups); err != nil {
		log.Fatalf("Failed to insert groups: %v", err)
	}

	principalRelationships, principals, err := gcp.FetchGroupsMembership(ctx, groups, projectId)
	if err := db.InsertPrincipalRelationships(database, principalRelationships); err != nil {
		log.Fatalf("Failed to insert principalRelationships: %v", err)
	}
	// In case there are external users, we need to track them too
	if err := db.InsertPrincipals(database, principals); err != nil {
		log.Fatalf("Failed to insert principals: %v", err)
	}

}

func syncHierarchy(ctx context.Context, database *sql.DB, gcpOrganizationID string) {
	hierarchies, err := gcp.FetchHierarchies(ctx, gcpOrganizationID)
	if err != nil {
		log.Fatalf("Error listing GCP hierarchies: %v", err)
	}
	if err := db.InsertHierarchies(database, hierarchies); err != nil {
		log.Fatalf("Failed to insert hierarchies: %v", err)
	}
}
