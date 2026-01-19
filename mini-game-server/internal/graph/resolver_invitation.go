package graph

import (
	"context"
	"fmt"

	"draw-and-guess-server/internal/graph/model"
)

// ========================================
// Invitation Mutations (Not Implemented)
// ========================================
// Note: Invitation feature is planned for future release
// Currently, users join rooms directly without invitations

// InviteUser is the resolver for the inviteUser field.
func (r *mutationResolver) InviteUser(ctx context.Context, roomID string, inviteeUserID string) (*model.Invitation, error) {
	return nil, fmt.Errorf("invitation feature is not yet implemented - join rooms directly instead")
}

// InviteUsers is the resolver for the inviteUsers field.
func (r *mutationResolver) InviteUsers(ctx context.Context, roomID string, inviteeUserIDs []string) ([]*model.Invitation, error) {
	return nil, fmt.Errorf("invitation feature is not yet implemented - join rooms directly instead")
}

// AcceptInvite is the resolver for the acceptInvite field.
func (r *mutationResolver) AcceptInvite(ctx context.Context, invitationID string) (*model.GameRoom, error) {
	return nil, fmt.Errorf("invitation feature is not yet implemented")
}

// RejectInvite is the resolver for the rejectInvite field.
func (r *mutationResolver) RejectInvite(ctx context.Context, invitationID string) (bool, error) {
	return false, fmt.Errorf("invitation feature is not yet implemented")
}

// ========================================
// Invitation Subscriptions (Not Implemented)
// ========================================

// InvitationReceived is the resolver for the invitationReceived field.
func (r *subscriptionResolver) InvitationReceived(ctx context.Context, userID string) (<-chan *model.Invitation, error) {
	return nil, fmt.Errorf("invitation feature is not yet implemented")
}

// ========================================
// Invitation Queries (Not Implemented)
// ========================================

// Invitation is the resolver for the invitation query field.
func (r *queryResolver) Invitation(ctx context.Context, id string) (*model.Invitation, error) {
	return nil, fmt.Errorf("invitation feature is not yet implemented")
}

// MyInvitations is the resolver for the myInvitations query field.
func (r *queryResolver) MyInvitations(ctx context.Context, userID string, status *model.InviteStatus) ([]*model.Invitation, error) {
	return nil, fmt.Errorf("invitation feature is not yet implemented")
}
