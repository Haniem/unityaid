package auth

import "context"

type Authorizer struct {
	repository *Repository
}

func NewAuthorizer(repository *Repository) *Authorizer {
	return &Authorizer{repository: repository}
}

func (a *Authorizer) CanCreateOrganization(ctx context.Context, claims Claims) (bool, error) {
	return a.IsSuperAdmin(ctx, claims)
}

func (a *Authorizer) CanManageOrganization(ctx context.Context, claims Claims, organizationID string) (bool, error) {
	if organizationID == "" {
		return false, nil
	}
	if ok, err := a.IsSuperAdmin(ctx, claims); ok || err != nil {
		return ok, err
	}
	return a.repository.HasRoleInOrganization(ctx, claims.UserID, organizationID, "org_admin")
}

func (a *Authorizer) CanManageContent(ctx context.Context, claims Claims, organizationID string) (bool, error) {
	if organizationID == "" {
		return a.IsSuperAdmin(ctx, claims)
	}
	if ok, err := a.IsSuperAdmin(ctx, claims); ok || err != nil {
		return ok, err
	}
	return a.repository.HasRoleInOrganization(ctx, claims.UserID, organizationID, "org_admin", "coordinator")
}

func (a *Authorizer) CanAccessOrganization(ctx context.Context, claims Claims, organizationID string) (bool, error) {
	if organizationID == "" {
		return a.IsSuperAdmin(ctx, claims)
	}
	if ok, err := a.IsSuperAdmin(ctx, claims); ok || err != nil {
		return ok, err
	}
	return a.repository.HasRoleInOrganization(ctx, claims.UserID, organizationID, "org_admin", "coordinator", "volunteer")
}

func (a *Authorizer) CanManageAnyContent(ctx context.Context, claims Claims) (bool, error) {
	if ok, err := a.IsSuperAdmin(ctx, claims); ok || err != nil {
		return ok, err
	}
	return a.repository.HasAnyRole(ctx, claims.UserID, "org_admin", "coordinator")
}

func (a *Authorizer) ManageableOrganizationIDs(ctx context.Context, claims Claims) ([]string, error) {
	if ok, err := a.IsSuperAdmin(ctx, claims); ok || err != nil {
		if err != nil {
			return nil, err
		}
		if ok {
			return nil, nil
		}
	}
	return a.repository.OrganizationIDsForRoles(ctx, claims.UserID, "org_admin", "coordinator")
}

func (a *Authorizer) CanManageUsers(ctx context.Context, claims Claims) (bool, error) {
	return a.IsSuperAdmin(ctx, claims)
}

func (a *Authorizer) CanEditVolunteer(ctx context.Context, claims Claims, targetUserID string) (bool, error) {
	if targetUserID == "" {
		return false, nil
	}
	if claims.UserID == targetUserID {
		return true, nil
	}
	if ok, err := a.IsSuperAdmin(ctx, claims); ok || err != nil {
		return ok, err
	}
	return a.repository.SharesOrganizationWithRole(ctx, claims.UserID, targetUserID, "org_admin", "coordinator")
}

func (a *Authorizer) IsSuperAdmin(ctx context.Context, claims Claims) (bool, error) {
	if ok, err := a.repository.HasSystemRole(ctx, claims.UserID, "system_admin"); ok || err != nil {
		return ok, err
	}
	if claims.Role == "super_admin" {
		return true, nil
	}
	return a.repository.HasAnyRole(ctx, claims.UserID, "super_admin")
}
