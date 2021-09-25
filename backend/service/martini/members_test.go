package martini_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xhttp/marshal"
	v1 "github.com/martinisecurity/trusty/api/v1"
	"github.com/martinisecurity/trusty/backend/service/martini"
	"github.com/martinisecurity/trusty/internal/db"
	"github.com/martinisecurity/trusty/internal/db/orgsdb/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrgsMembers(t *testing.T) {
	svc := trustyServer.Service(martini.ServiceName).(*martini.Service)
	mp := strings.Replace(v1.PathForMartiniOrgMembers, ":org_id", "23456", 1)
	r, err := http.NewRequest(http.MethodGet, mp, nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	svc.GetOrgMembersHandler()(w, r, rest.Params{
		{
			Key:   "org_id",
			Value: "invalid",
		},
	})
	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, `{"code":"invalid_parameter","message":"invalid org_id"}`, w.Body.String())

	w = httptest.NewRecorder()
	svc.GetOrgMembersHandler()(w, r, rest.Params{
		{
			Key:   "org_id",
			Value: "1234567",
		},
	})
	require.Equal(t, http.StatusOK, w.Code)

	var mres v1.OrgMembersResponse
	require.NoError(t, marshal.Decode(w.Body, &mres))
	assert.Empty(t, mres.Members)
}

func TestOrgMemberHandler(t *testing.T) {
	ctx := context.Background()
	svc := trustyServer.Service(martini.ServiceName).(*martini.Service)

	orgdb := svc.Db()
	id, _ := orgdb.NextID()

	g := guid.MustCreate()
	org, err := orgdb.UpdateOrg(ctx, &model.Organization{
		ID:         id,
		ExternalID: strconv.FormatUint(id, 10),
		Provider:   v1.ProviderMartini,
		Login:      g,
		Email:      g + "@trusty.com",
	})
	require.NoError(t, err)

	user, err := orgdb.LoginUser(ctx, &model.User{
		Email:      "denis+test@ekspand.com",
		Name:       "test user",
		Login:      "denis+test@ekspand.com",
		ExternalID: "123456",
		Provider:   v1.ProviderGoogle,
	})
	require.NoError(t, err)

	ms, err := orgdb.AddOrgMember(ctx, org.ID, user.ID, v1.RoleOwner, v1.ProviderMartini)
	require.NoError(t, err)
	assert.Equal(t, v1.RoleOwner, ms.Role)
	assert.Equal(t, user.ID, ms.UserID)
	assert.Equal(t, org.ID, ms.OrgID)

	orgID := db.IDString(org.ID)

	var memberID string
	// ADD member
	{

		addReq := &v1.OrgMemberRequest{
			Action: "ADD",
			Email:  guid.MustCreate() + "@trusty.com",
			Role:   v1.RoleAdmin,
		}
		js, err := json.Marshal(addReq)
		require.NoError(t, err)

		r, err := http.NewRequest(http.MethodPost, v1.PathForMartiniOrgMembers, bytes.NewReader(js))
		require.NoError(t, err)
		r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

		w := httptest.NewRecorder()
		svc.OrgMemberHandler()(w, r, rest.Params{
			{
				Key:   "org_id",
				Value: orgID,
			},
		})
		require.Equal(t, http.StatusOK, w.Code)

		var mres v1.OrgMembersResponse
		require.NoError(t, marshal.Decode(w.Body, &mres))
		require.Len(t, mres.Members, 2)
		memberID = mres.Members[1].UserID
	}

	// REMOVE member
	{

		remReq := &v1.OrgMemberRequest{
			Action: "REMOVE",
			UserID: memberID,
		}
		js, err := json.Marshal(remReq)
		require.NoError(t, err)

		r, err := http.NewRequest(http.MethodPost, v1.PathForMartiniOrgMembers, bytes.NewReader(js))
		require.NoError(t, err)
		r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

		w := httptest.NewRecorder()
		svc.OrgMemberHandler()(w, r, rest.Params{
			{
				Key:   "org_id",
				Value: orgID,
			},
		})
		require.Equal(t, http.StatusOK, w.Code)

		var mres v1.OrgMembersResponse
		require.NoError(t, marshal.Decode(w.Body, &mres))
		assert.Len(t, mres.Members, 1)
	}
}
