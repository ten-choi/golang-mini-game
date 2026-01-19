package utils

import (
	"log"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	node     *snowflake.Node
	nodeOnce sync.Once
)

// InitSnowflake initializes the Snowflake ID generator
// nodeID should be unique across distributed systems (0-1023)
func InitSnowflake(nodeID int64) error {
	var err error
	nodeOnce.Do(func() {
		node, err = snowflake.NewNode(nodeID)
		if err != nil {
			log.Printf("Failed to initialize Snowflake node: %v", err)
			return
		}
		log.Printf("Snowflake ID generator initialized with node ID: %d", nodeID)
	})
	return err
}

// GenerateID generates a new Snowflake ID
func GenerateID() int64 {
	if node == nil {
		log.Panic("Snowflake node not initialized. Call InitSnowflake first.")
	}
	return node.Generate().Int64()
}

// GetNode returns the Snowflake node instance
func GetNode() *snowflake.Node {
	return node
}
