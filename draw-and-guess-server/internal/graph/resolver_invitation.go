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

// InviteUsers is the resolver for the inviteUsers field.
func (r *mutationResolver) InviteUsers(ctx context.Context, roomID string, inviteeUsernames []string) ([]*model.Invitation, error) {
	// TODO: Implement batch invitation logic with database
	invitations := make([]*model.Invitation, 0, len(inviteeUsernames))

	for _, username := range inviteeUsernames {
		invitation := &model.Invitation{
			ID:        uuid.New().String(),
			RoomID:    roomID,
			InviteeID: username,
			Status:    model.InviteStatusPending,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(5 * time.Minute),
		}
		invitations = append(invitations, invitation)
	}

	return invitations, nil
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

// InvitationReceived is the resolver for the invitationReceived field.
func (r *subscriptionResolver) InvitationReceived(ctx context.Context, userID string) (<-chan *model.Invitation, error) {
	// TODO: Implement invitation subscription
	ch := make(chan *model.Invitation)

	// Close channel when context is done
	go func() {
		<-ctx.Done()
		close(ch)
	}()

	return ch, nil
}

// Invitation is the resolver for the invitation query field.
func (r *queryResolver) Invitation(ctx context.Context, id string) (*model.Invitation, error) {
	// TODO: Implement invitation query from database
	return nil, fmt.Errorf("not implemented yet")
}

// MyInvitations is the resolver for the myInvitations query field.
func (r *queryResolver) MyInvitations(ctx context.Context, userID string, status *model.InviteStatus) ([]*model.Invitation, error) {
	// TODO: Implement myInvitations query from database
	return []*model.Invitation{}, nil
}
