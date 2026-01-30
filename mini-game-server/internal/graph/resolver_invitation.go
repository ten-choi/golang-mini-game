package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"draw-and-guess-server/internal/common"
	"draw-and-guess-server/internal/graph/model"
	"draw-and-guess-server/internal/valkey"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// ========================================
// Invitation Mutations (Not Implemented)
// ========================================
// Note: Invitation feature is planned for future release
// Currently, users join rooms directly without invitations

// InviteUser is the resolver for the inviteUser field.
func (r *mutationResolver) InviteUser(ctx context.Context, roomID string, inviteeUsername string) (*model.Invitation, error) {
	// Get inviter info from context (would need auth middleware)
	// For now, we'll use a placeholder
	inviterUserID := "" // TODO: Get from auth context

	// Find user by username to get userId
	inviteeUser, err := r.UserService.GetUserByName(ctx, inviteeUsername)
	if err != nil {
		return nil, fmt.Errorf("user not found: %s", inviteeUsername)
	}
	inviteeUserID := inviteeUser.ID.Hex()

	// Check invitee's status
	statusKey := common.RedisKeyUserStatus + inviteeUserID
	status, err := valkey.Client.Get(ctx, statusKey).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get user status: %w", err)
	}

	// Reject if user is in game
	if status == common.UserStatusInGame {
		return nil, fmt.Errorf("user is currently in a game")
	}

	// Reject if user is offline (no status in Redis)
	if status == "" || status == common.UserStatusOffline {
		return nil, fmt.Errorf("user is offline")
	}

	// Get room info
	room, exists := GetGameRoom(roomID)
	if !exists {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	// Check if room is full
	if len(room.Users) >= int(room.MaxUsers) {
		return nil, fmt.Errorf("room is full")
	}

	// Create invitation
	invitationID := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	invitation := &model.Invitation{
		ID:        invitationID,
		RoomID:    roomID,
		InviterID: inviterUserID,
		InviteeID: inviteeUserID,
		Status:    model.InviteStatusPending,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	// Store invitation in Redis
	invitationData := map[string]interface{}{
		"id":        invitationID,
		"roomId":    roomID,
		"roomName":  room.Name,
		"gameType":  room.GameType,
		"inviterId": inviterUserID,
		"inviteeId": inviteeUserID,
		"status":    "PENDING",
		"createdAt": now,
		"expiresAt": expiresAt,
	}

	invitationJSON, err := json.Marshal(invitationData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal invitation: %w", err)
	}

	invitationKey := common.RedisKeyInvitation + invitationID
	err = valkey.Client.Set(ctx, invitationKey, invitationJSON, time.Duration(common.RedisInvitationTTL)*time.Second).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store invitation: %w", err)
	}

	// Publish invitation to invitee's personal channel
	channel := common.ChannelUserPrefix + inviteeUserID
	
	// Add type field to invitation data
	invitationData["type"] = "INVITATION"

	msgBytes, err := json.Marshal(invitationData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal invitation message: %w", err)
	}

	err = valkey.Client.Publish(ctx, channel, msgBytes).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to publish invitation: %w", err)
	}

	logger := common.GetLogger()
	logger.Info("[Invitation] Sent invitation %s to user %s (status: %s) for room %s via channel %s", invitationID, inviteeUserID, status, roomID, channel)

	return invitation, nil
}

// InviteUsers is the resolver for the inviteUsers field.
func (r *mutationResolver) InviteUsers(ctx context.Context, roomID string, inviteeUsernames []string) ([]*model.Invitation, error) {
	invitations := make([]*model.Invitation, 0, len(inviteeUsernames))

	for _, username := range inviteeUsernames {
		invitation, err := r.InviteUser(ctx, roomID, username)
		if err != nil {
			// Log error but continue with other invitations
			logger := common.GetLogger()
			logger.Warn("[Invitation] Failed to invite user %s: %v", username, err)
			continue
		}
		invitations = append(invitations, invitation)
	}

	if len(invitations) == 0 {
		return nil, fmt.Errorf("all invitations failed")
	}

	return invitations, nil
}

// AcceptInvite is the resolver for the acceptInvite field.
func (r *mutationResolver) AcceptInvite(ctx context.Context, invitationID string) (*model.GameRoom, error) {
	// Get invitation from Redis
	invitationKey := common.RedisKeyInvitation + invitationID
	invitationJSON, err := valkey.Client.Get(ctx, invitationKey).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("invitation not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	// Parse invitation data
	var invitationData map[string]interface{}
	if err := json.Unmarshal([]byte(invitationJSON), &invitationData); err != nil {
		return nil, fmt.Errorf("failed to parse invitation: %w", err)
	}

	// Extract fields
	roomID, _ := invitationData["roomId"].(string)
	inviteeID, _ := invitationData["inviteeId"].(string)
	status, _ := invitationData["status"].(string)

	if roomID == "" || inviteeID == "" {
		return nil, fmt.Errorf("invalid invitation data")
	}

	// Check if invitation is still pending
	if status != "PENDING" {
		return nil, fmt.Errorf("invitation is no longer valid (status: %s)", status)
	}

	// Update invitation status to ACCEPTED
	invitationData["status"] = "ACCEPTED"
	updatedJSON, _ := json.Marshal(invitationData)
	valkey.Client.Set(ctx, invitationKey, updatedJSON, time.Duration(common.RedisInvitationTTL)*time.Second)

	// Join the game room
	room, err := r.JoinGameRoom(ctx, roomID, inviteeID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to join room: %w", err)
	}

	logger := common.GetLogger()
	logger.Info("[Invitation] User %s accepted invitation %s and joined room %s", inviteeID, invitationID, roomID)

	return room, nil
}

// RejectInvite is the resolver for the rejectInvite field.
func (r *mutationResolver) RejectInvite(ctx context.Context, invitationID string) (bool, error) {
	return false, fmt.Errorf("invitation feature is not yet implemented")
}

// ========================================
// Invitation Queries
// ========================================

// MyInvitations is the resolver for the myInvitations field.
func (r *queryResolver) MyInvitations(ctx context.Context, userID string, status *model.InviteStatus) ([]*model.Invitation, error) {
	// For now, return empty list. Full implementation would scan Redis for user's invitations
	// In a production system, you'd want to maintain an index of invitations per user
	return []*model.Invitation{}, nil
}

// Invitation is the resolver for the invitation field.
func (r *queryResolver) Invitation(ctx context.Context, id string) (*model.Invitation, error) {
	// Get invitation from Redis
	invitationKey := common.RedisKeyInvitation + id
	invitationJSON, err := valkey.Client.Get(ctx, invitationKey).Result()
	if err == redis.Nil {
		return nil, nil // Invitation not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	// Parse invitation data
	var invitationData map[string]interface{}
	if err := json.Unmarshal([]byte(invitationJSON), &invitationData); err != nil {
		return nil, fmt.Errorf("failed to parse invitation: %w", err)
	}

	// Convert to Invitation model
	invitation := &model.Invitation{
		ID:        id,
		RoomID:    invitationData["roomId"].(string),
		InviterID: invitationData["inviterId"].(string),
		InviteeID: invitationData["inviteeId"].(string),
		Status:    model.InviteStatus(invitationData["status"].(string)),
	}

	return invitation, nil
}

// ========================================
// Invitation Subscriptions
// ========================================

// InvitationReceived is the resolver for the invitationReceived field.
func (r *subscriptionResolver) InvitationReceived(ctx context.Context, userID string) (<-chan *model.Invitation, error) {
	invitations := make(chan *model.Invitation, 1)

	go func() {
		defer close(invitations)

		// Subscribe to user's personal invitation channel
		channel := common.ChannelUserPrefix + userID
		pubsub := valkey.Client.Subscribe(ctx, channel)
		defer pubsub.Close()

		ch := pubsub.Channel()

		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				if msg == nil {
					return
				}

				// Parse message
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(msg.Payload), &data); err != nil {
					continue
				}

				// Check if it's an invitation message
				if msgType, ok := data["type"].(string); ok && msgType == "INVITATION" {
					if invData, ok := data["data"].(map[string]interface{}); ok {
						invitation := &model.Invitation{
							ID:        invData["id"].(string),
							RoomID:    invData["roomId"].(string),
							InviterID: invData["inviterId"].(string),
							InviteeID: invData["inviteeId"].(string),
							Status:    model.InviteStatusPending,
						}

						select {
						case invitations <- invitation:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}()

	return invitations, nil
}
