package gcp

import (
	"context"
	"fmt"
	"github.com/ttauveron/gcp-iam-dumper/pkg/model"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/cloudidentity/v1"
	"google.golang.org/api/option"
	"log"
	"strings"
	"sync"
)

type GroupWithMembers struct {
	Group   model.Principal
	Members []model.PrincipalRelationship
	Users   []model.Principal
}

func FetchGroupsMembership(ctx context.Context, groups []model.Principal, projectId string) ([]model.PrincipalRelationship, []model.Principal, error) {
	service, err := cloudidentity.NewService(ctx, option.WithQuotaProject(projectId))
	membershipsService := cloudidentity.NewGroupsMembershipsService(service)
	if err != nil {
		log.Fatalf("cloudidentity.NewService: %v", err)
	}

	var wg sync.WaitGroup
	groupsWithMembersChan := make(chan GroupWithMembers, len(groups))

	for _, group := range groups {

		wg.Add(1)
		go func(group model.Principal) {
			defer wg.Done()

			req := membershipsService.List(group.ID)
			resp, err := req.Do()
			if err != nil {
				log.Printf("Error listing members for group %s: %v", group.Name, err)
				return
			}

			var principalRelationships []model.PrincipalRelationship
			var principals []model.Principal
			for _, membership := range resp.Memberships {
				parts := strings.Split(membership.Name, "/")
				groupID := parts[0] + "/" + parts[1]
				memberID := parts[3]
				memberType := "user"
				// If the member is a group, its ID contains letters, else it's a user.
				if strings.ContainsAny(memberID, "abcdefghijklmnopqrstuvwxyz") {
					memberID = "groups/" + memberID
					memberType = "group"
				}
				principalRelationships = append(principalRelationships, model.PrincipalRelationship{ParentID: groupID, ChildID: memberID})
				principals = append(principals, model.Principal{ID: memberID, Name: membership.PreferredMemberKey.Id, Type: memberType})
			}

			groupsWithMembersChan <- GroupWithMembers{Group: group, Members: principalRelationships, Users: principals}
		}(group)
	}

	wg.Wait()
	close(groupsWithMembersChan)

	var principalRelationships []model.PrincipalRelationship
	var principals []model.Principal
	for gm := range groupsWithMembersChan {
		principalRelationships = append(principalRelationships, gm.Members...)
		principals = append(principals, gm.Users...)
	}

	return principalRelationships, principals, nil
}

func FetchGroups(ctx context.Context, projectId, organization string) ([]model.Principal, error) {
	service, err := cloudidentity.NewService(ctx, option.WithQuotaProject(projectId))
	if err != nil {
		log.Fatalf("cloudidentity.NewService: %v", err)
	}
	groupsService := cloudidentity.NewGroupsService(service)
	req := groupsService.List().Parent(fmt.Sprintf("customers/%s", organization))
	resp, err := req.Do()
	if err != nil {
		return nil, err
	}

	var principals []model.Principal
	for _, group := range resp.Groups {
		principals = append(principals, model.Principal{ID: group.Name, Name: group.GroupKey.Id, Type: "group"})
	}

	return principals, nil
}

func FetchUsers(ctx context.Context, projectId, customerID string) ([]model.Principal, error) {
	service, err := admin.NewService(ctx, option.WithQuotaProject(projectId))
	if err != nil {
		return nil, fmt.Errorf("admin.NewService: %v", err)
	}

	var users []model.Principal
	// Call the Admin SDK Directory API
	req := service.Users.List().Customer(customerID).MaxResults(500)
	err = req.Pages(ctx, func(page *admin.Users) error {
		for _, user := range page.Users {
			users = append(users, model.Principal{
				ID:   user.Id,
				Name: user.PrimaryEmail,
				Type: "user",
			})
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("Users.List: %v", err)
	}

	return users, nil
}
