package query

import (
	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/pkg/storage"
)

// UserMembershipManager represents User Membership manager.
type UserMembershipManager struct {
	*Manager
}

// NewUserMembershipManager creates and returns new User Membership manager.
func NewUserMembershipManager(c storage.DBContext) *UserMembershipManager {
	return &UserMembershipManager{Manager: NewTenantManager(c)}
}

// Q outputs objects query.
func (mgr *UserMembershipManager) Q(o interface{}) *orm.Query {
	return mgr.Query(o)
}

// ForUserQ outputs objects query filtered by user.
func (mgr *UserMembershipManager) ForUserQ(user *models.User, o interface{}) *orm.Query {
	return mgr.Q(o).Where("user_id = ?", user.ID)
}

// ForGroupQ outputs objects query filtered by group.
func (mgr *UserMembershipManager) ForGroupQ(group *models.UserGroup, o interface{}) *orm.Query {
	return mgr.Q(o).Where("group_id = ?", group.ID)
}

// ForUserAndGroupQ outputs objects query filtered by user and group.
func (mgr *UserMembershipManager) ForUserAndGroupQ(o *models.UserMembership) *orm.Query {
	return mgr.Q(o).Where("user_id = ?", o.UserID).Where("group_id = ?", o.GroupID)
}
