package role

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/iancoleman/orderedmap"
	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRole_BuildPermissionsFromAPIResponse_Empty(t *testing.T) {
	role := &Role{}
	err := role.BuildPermissionsFromAPIResponse(&sdk.Role{
		Permissions: &orderedmap.OrderedMap{},
	})

	assert.NoError(t, err)
	assert.Nil(t, role.Permission)
}

func TestRole_BuildPermissionsFromAPIResponse_Error(t *testing.T) {
	permissions := orderedmap.New()
	permissions.Set("ContentModel", 1)

	role := &Role{}
	err := role.BuildPermissionsFromAPIResponse(&sdk.Role{
		Permissions: permissions,
	})

	assert.Error(t, err)
}

func TestRole_BuildPermissionsFromAPIResponse_ErrorSlice(t *testing.T) {
	permissions := orderedmap.New()
	permissions.Set("ContentModel", []interface{}{1, 2})

	role := &Role{}
	err := role.BuildPermissionsFromAPIResponse(&sdk.Role{
		Permissions: permissions,
	})

	assert.Error(t, err)
}

func TestRole_BuildPermissionsFromAPIResponse_All(t *testing.T) {
	permissions := orderedmap.New()
	permissions.Set("ContentModel", "all")

	r := &Role{}
	err := r.BuildPermissionsFromAPIResponse(&sdk.Role{
		Permissions: permissions,
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, len(r.Permission))
	assert.Equal(t, r.Permission, []Permission{
		{
			ID:     types.StringValue("ContentModel"),
			Value:  types.StringValue("all"),
			Values: nil,
		},
	})
}

func TestRole_BuildPermissionsFromAPIResponse_Values(t *testing.T) {
	permissions := orderedmap.New()
	permissions.Set("ContentModel", []string{"read"})

	r := &Role{}
	err := r.BuildPermissionsFromAPIResponse(&sdk.Role{
		Permissions: permissions,
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, len(r.Permission))
	assert.Equal(t, r.Permission, []Permission{
		{
			ID:    types.StringValue("ContentModel"),
			Value: types.StringNull(),
			Values: []basetypes.StringValue{
				types.StringValue("read"),
			},
		},
	})
}

func TestRole_BuildPermissionsFromAPIResponse_Combination(t *testing.T) {
	permissions := orderedmap.New()
	permissions.Set("ContentModel", []string{"read"})
	permissions.Set("Settings", []string{})
	permissions.Set("ContentDelivery", "all")

	r := &Role{}
	err := r.BuildPermissionsFromAPIResponse(&sdk.Role{
		Permissions: permissions,
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(r.Permission))
	assert.Equal(t, r.Permission, []Permission{
		{
			ID:    types.StringValue("ContentModel"),
			Value: types.StringNull(),
			Values: []basetypes.StringValue{
				types.StringValue("read"),
			},
		},
		{
			ID:     types.StringValue("ContentDelivery"),
			Value:  types.StringValue("all"),
			Values: nil,
		},
	})
}
