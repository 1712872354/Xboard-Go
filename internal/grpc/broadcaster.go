package grpc

import (
	"fmt"
	"sync"

	"xboard-go/pkg/logger"
)

// eventChannelBuffer is the buffered channel size per node.
const eventChannelBuffer = 16

// Broadcaster manages per-node event channels.
// When the panel needs to push config or user changes to a connected node,
// it calls the appropriate Broadcast* method, which enqueues a PanelMessage
// onto the node's channel.  If the channel is full the oldest message is
// dropped to avoid blocking the caller.
type Broadcaster struct {
	mu       sync.RWMutex
	channels map[uint32]chan *PanelMessage
}

// NewBroadcaster creates a new Broadcaster instance.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		channels: make(map[uint32]chan *PanelMessage),
	}
}

// Subscribe creates (or returns the existing) event channel for a node.
// The returned channel is read-only; the Stream handler reads from it.
func (b *Broadcaster) Subscribe(nodeID uint32) <-chan *PanelMessage {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch, exists := b.channels[nodeID]
	if !exists {
		ch = make(chan *PanelMessage, eventChannelBuffer)
		b.channels[nodeID] = ch
	}
	return ch
}

// Unsubscribe closes and removes the channel for a node.
func (b *Broadcaster) Unsubscribe(nodeID uint32) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if ch, exists := b.channels[nodeID]; exists {
		close(ch)
		delete(b.channels, nodeID)
	}
}

// sendNonBlocking attempts to send a message to the channel without blocking.
// If the channel is full, the oldest message is drained first.
func (b *Broadcaster) sendNonBlocking(nodeID uint32, msg *PanelMessage) {
	b.mu.RLock()
	ch, exists := b.channels[nodeID]
	b.mu.RUnlock()

	if !exists {
		return
	}

	// Try a non-blocking send first.
	select {
	case ch <- msg:
		return
	default:
	}

	// Channel full — drain the oldest message, then send.
	select {
	case <-ch:
		logger.Sugar().Debugf("Broadcaster: dropped oldest message for node %d (channel full)", nodeID)
	default:
	}

	select {
	case ch <- msg:
	default:
		// Still full after drain (race); log and drop.
		logger.Sugar().Warnf("Broadcaster: could not enqueue message for node %d", nodeID)
	}
}

// BroadcastConfig sends a ConfigUpdate to the specified node.
func (b *Broadcaster) BroadcastConfig(nodeID uint32, config *NodeConfig) {
	b.sendNonBlocking(nodeID, &PanelMessage{
		Payload: &ConfigUpdate{Config: config},
	})
	logger.Sugar().Debugf("Broadcaster: config update sent to node %d", nodeID)
}

// BroadcastUsers sends a full UserListUpdate to the specified node.
func (b *Broadcaster) BroadcastUsers(nodeID uint32, users []*User) {
	b.sendNonBlocking(nodeID, &PanelMessage{
		Payload: &UserListUpdate{Users: users},
	})
	logger.Sugar().Debugf("Broadcaster: user list update sent to node %d (%d users)", nodeID, len(users))
}

// BroadcastUserDelta sends an incremental UserDelta to the specified node.
func (b *Broadcaster) BroadcastUserDelta(nodeID uint32, added []*User, removed []uint32) {
	b.sendNonBlocking(nodeID, &PanelMessage{
		Payload: &UserDelta{Added: added, Removed: removed},
	})
	logger.Sugar().Debugf("Broadcaster: user delta sent to node %d (+%d -%d)", nodeID, len(added), len(removed))
}

// BroadcastPong sends a Pong heartbeat response to the specified node.
func (b *Broadcaster) BroadcastPong(nodeID uint32, timestamp int64) {
	b.sendNonBlocking(nodeID, &PanelMessage{
		Payload: &Pong{Timestamp: timestamp},
	})
}

// SubscriberCount returns the number of active subscribers (for diagnostics).
func (b *Broadcaster) SubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.channels)
}

// HasSubscriber checks whether a given node has an active stream.
func (b *Broadcaster) HasSubscriber(nodeID uint32) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, exists := b.channels[nodeID]
	return exists
}

// GetSubscriberNodeIDs returns the node IDs of all active subscribers.
func (b *Broadcaster) GetSubscriberNodeIDs() []uint32 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	ids := make([]uint32, 0, len(b.channels))
	for id := range b.channels {
		ids = append(ids, id)
	}
	return ids
}

// String implements fmt.Stringer for debugging.
func (b *Broadcaster) String() string {
	return fmt.Sprintf("Broadcaster(subscribers=%d)", b.SubscriberCount())
}
