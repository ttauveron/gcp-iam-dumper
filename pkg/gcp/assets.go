package gcp

import (
	asset "cloud.google.com/go/asset/apiv1"
	"cloud.google.com/go/asset/apiv1/assetpb"
	"context"
	"fmt"
	"github.com/ttauveron/gcp-iam-dumper/pkg/model"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/structpb"
	"log"
	"strings"
)

func FetchServiceAccounts(ctx context.Context, scope string) ([]model.Principal, error) {
	client, err := asset.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("asset.NewClient: %v", err)
	}
	defer client.Close()

	req := &assetpb.SearchAllResourcesRequest{
		Scope: scope, // e.g., "organizations/123456789"
		AssetTypes: []string{
			"iam.googleapis.com/ServiceAccount",
		},
	}
	var serviceAccounts []model.Principal

	it := client.SearchAllResources(ctx, req)
	for {
		serviceAccount, err := it.Next()

		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		segments := strings.Split(serviceAccount.Name, "/")
		serviceAccountEmail := segments[len(segments)-1]

		serviceAccounts = append(serviceAccounts, model.Principal{
			ID:   serviceAccountEmail,
			Name: serviceAccountEmail,
			Type: "serviceAccount",
		})
	}

	return serviceAccounts, nil
}

func FetchHierarchies(ctx context.Context, scope string) ([]model.Hierarchy, error) {
	client, err := asset.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("asset.NewClient: %v", err)
	}
	defer client.Close()

	req := &assetpb.SearchAllResourcesRequest{
		Scope: scope, // e.g., "organizations/123456789"
		AssetTypes: []string{
			"cloudresourcemanager.googleapis.com/Folder",
			"cloudresourcemanager.googleapis.com/Project",
			"cloudresourcemanager.googleapis.com/Organization",
		},
	}
	var hierarchies []model.Hierarchy

	it := client.SearchAllResources(ctx, req)
	for {
		hierarchy, err := it.Next()

		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		var id, name string

		switch hierarchy.AssetType {
		case "cloudresourcemanager.googleapis.com/Project":
			id = hierarchy.Project
			name = strings.TrimPrefix(hierarchy.Name, "//cloudresourcemanager.googleapis.com/projects/")
		case "cloudresourcemanager.googleapis.com/Folder", "cloudresourcemanager.googleapis.com/Organization":
			id = strings.TrimPrefix(hierarchy.Name, "//cloudresourcemanager.googleapis.com/")
			name = hierarchy.DisplayName
		default:
			return nil, fmt.Errorf("unknown hierarchy type: %s", hierarchy.AssetType)
		}
		hierarchies = append(hierarchies, model.Hierarchy{
			ID:       id,
			Name:     name,
			Type:     strings.ToLower(strings.TrimPrefix(hierarchy.AssetType, "cloudresourcemanager.googleapis.com/")),
			ParentID: strings.TrimPrefix(hierarchy.ParentFullResourceName, "//cloudresourcemanager.googleapis.com/"),
		})
	}

	return hierarchies, nil
}

func FetchAssetIAMPolicy(ctx context.Context, scope string) ([]model.ResourceIAMPermission, error) {
	client, err := asset.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("asset.NewClient: %v", err)
	}
	defer client.Close()

	req := &assetpb.SearchAllIamPoliciesRequest{
		Scope: scope, // e.g., "organizations/123456789"
		Query: "memberTypes=(group OR user OR allUsers OR serviceAccount) OR memberTypes:deleted",
	}
	var resourcePolicies []model.ResourceIAMPermission
	it := client.SearchAllIamPolicies(ctx, req)
	for {
		policy, err := it.Next()

		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		var hierarchyID string
		if policy.Project != "" {
			hierarchyID = policy.Project
		} else if len(policy.Folders) == 0 {
			hierarchyID = policy.Organization
		} else {
			hierarchyID = policy.Folders[0]
		}

		for _, binding := range policy.Policy.Bindings {
			condition := ""
			if binding.Condition != nil {
				condition = binding.Condition.Title
			}
			for _, member := range binding.Members {
				if strings.HasPrefix(member, "project") {
					continue
				}
				principalEmail := member
				parts := strings.SplitN(member, ":", 2)
				if len(parts) == 2 {
					principalEmail = parts[1]
				}
				resourcePolicies = append(resourcePolicies, model.ResourceIAMPermission{
					ResourceID:  policy.Resource,
					PrincipalID: principalEmail,
					RoleID:      binding.Role,
					Conditional: condition,
					AssetType:   policy.AssetType,
					HierarchyID: hierarchyID,
				})
			}
		}
	}

	return resourcePolicies, nil
}

func fetchCustomRoles(ctx context.Context, scope string) ([]model.Role, error) {
	client, err := asset.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("asset.NewClient: %v", err)
	}
	defer client.Close()

	req := &assetpb.SearchAllResourcesRequest{
		Scope: scope, // e.g., "organizations/123456789"
		AssetTypes: []string{
			"iam.googleapis.com/Role",
		},
	}
	var customRoles []model.Role

	it := client.SearchAllResources(ctx, req)
	for {
		role, err := it.Next()

		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		var permissions []string
		if attr, exists := role.AdditionalAttributes.Fields["includedPermissions"]; exists {
			if lv, ok := attr.GetKind().(*structpb.Value_ListValue); ok {
				for _, v := range lv.ListValue.Values {
					if perm, ok := v.GetKind().(*structpb.Value_StringValue); ok {
						permissions = append(permissions, perm.StringValue)
					}
				}
			}
		}
		customRoles = append(customRoles, model.Role{
			ID:          strings.TrimPrefix(role.Name, "//iam.googleapis.com/"),
			Title:       role.DisplayName,
			Permissions: permissions,
		})
	}
	return customRoles, nil
}
