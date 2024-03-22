package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationRemoveRoleFromUsersExecuted event.Type = "admin_api.mutation.remove_role_from_users.executed"
)

type AdminAPIMutationRemoveRoleFromUsersExecutedEventPayload struct {
	AffectedUserIDs []string `json:"-"`
}

func (e *AdminAPIMutationRemoveRoleFromUsersExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationRemoveRoleFromUsersExecuted
}

func (e *AdminAPIMutationRemoveRoleFromUsersExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationRemoveRoleFromUsersExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationRemoveRoleFromUsersExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationRemoveRoleFromUsersExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationRemoveRoleFromUsersExecutedEventPayload) ForAudit() bool {
	// FIXME(tung): Should be true
	return false
}

func (e *AdminAPIMutationRemoveRoleFromUsersExecutedEventPayload) RequireReindexUserIDs() []string {
	return e.AffectedUserIDs
}

func (e *AdminAPIMutationRemoveRoleFromUsersExecutedEventPayload) DeletedUserIDs() []string {
	return []string{}
}

var _ event.NonBlockingPayload = &AdminAPIMutationRemoveRoleFromUsersExecutedEventPayload{}
