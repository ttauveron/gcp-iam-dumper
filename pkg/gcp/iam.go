package gcp

import (
	admin "cloud.google.com/go/iam/admin/apiv1"
	"cloud.google.com/go/iam/admin/apiv1/adminpb"
	"context"
	"fmt"
	"github.com/ttauveron/gcp-iam-dumper/pkg/model"
)

// listBuiltInRoles fetches all built-in roles in GCP.

func FetchAllRoles(ctx context.Context, scope string) ([]model.Role, error) {
	predefinedRoles, err := fetchPredefinedRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch predefined roles: %v", err)
	}
	customRoles, err := fetchCustomRoles(ctx, scope)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch predefined roles: %v", err)
	}
	return append(predefinedRoles, customRoles...), nil
}

func fetchPredefinedRoles(ctx context.Context) ([]model.Role, error) {
	iamClient, err := admin.NewIamClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create IAM client: %v", err)
	}
	defer iamClient.Close()

	var rolesBatch []*adminpb.Role
	var roles []model.Role
	nextPageToken := ""
	for {
		req := &adminpb.ListRolesRequest{
			//Parent: "organizations/0", // Use "organizations/0" for Google-managed rolesBatch
			PageToken: nextPageToken,
			PageSize:  1000,
			View:      adminpb.RoleView_FULL,
		}

		resp, err := iamClient.ListRoles(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to list rolesBatch: %v", err)
		}

		rolesBatch = append(rolesBatch, resp.Roles...)
		nextPageToken = resp.NextPageToken
		if nextPageToken == "" {
			break // No more pages
		}
	}

	for _, role := range rolesBatch {
		roles = append(roles, model.Role{
			ID:          role.Name,
			Title:       role.Title,
			Permissions: role.IncludedPermissions,
		})

	}

	return roles, nil
}
