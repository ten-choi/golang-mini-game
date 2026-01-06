package graph

import (
	"context"
	"fmt"
	"time"

	"draw-and-guess-server/internal/graph/model"

	"github.com/google/uuid"
)

// InviteUser is the resolver for the inviteUser field.
func (r *mutationResolver) InviteUser(ctx context.Context, roomID string, inviteeUsername string) (*model.Invitation, error) {
	// TODO: Implement invitation logic with database
	invitation := &model.Invitation{
		ID:        uuid.New().String(),
		RoomID:    roomID,
		InviteeID: inviteeUsername,
		Status:    model.InviteStatusPending,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	return invitation, nil
}

// AcceptInvite is the resolver for the acceptInvite field.
func (r *mutationResolver) AcceptInvite(ctx context.Context, invitationID string) (*model.GameRoom, error) {
	// TODO: Implement invitation acceptance logic
	return nil, fmt.Errorf("not implemented yet")
}

// RejectInvite is the resolver for the rejectInvite field.
func (r *mutationResolver) RejectInvite(ctx context.Context, invitationID string) (bool, error) {
	// TODO: Implement invitation rejection logic
	return false, fmt.Errorf("not implemented yet")
}
