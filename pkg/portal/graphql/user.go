package graphql

import (
	"context"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"

	"github.com/authgear/authgear-server/pkg/portal/model"
)

const typeUser = "User"

var nodeUser = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name:        typeUser,
		Description: "Portal User",
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
		},
		Fields: graphql.Fields{
			"id": relay.GlobalIDField(typeUser, nil),
		},
	}),
	&model.User{},
	func(ctx context.Context, id string) (interface{}, error) {
		// FIXME(portal): check id
		gqlCtx := GQLContext(ctx)
		return gqlCtx.Viewer.Get()
	},
)
