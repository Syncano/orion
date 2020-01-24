package serializers

import (
	"github.com/Syncano/orion/app/models"
)

type UserSerializer struct {
	Class *models.Class
}

func (s UserSerializer) Response(i interface{}) interface{} {
	o := i.(*models.User)
	base := map[string]interface{}{
		"id":         o.ID,
		"username":   o.Username,
		"user_key":   o.Key,
		"created_at": &o.Profile.CreatedAt,
		"updated_at": &o.Profile.UpdatedAt,
		"revision":   o.Profile.Revision,
	}

	processDataObjectFields(s.Class, o.Profile, base)

	return base
}

func (s UserSerializer) ResponseWithGroup(i interface{}) interface{} {
	o := i.(*models.User)
	gSerializer := UserGroupSerializer{}
	base := s.Response(i).(map[string]interface{})

	groups := make([]interface{}, 0)
	for _, group := range o.Groups {
		groups = append(groups, gSerializer.ShortResponse(group))
	}

	base["groups"] = groups

	return base
}
