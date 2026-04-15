package consumer

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestConsumerStruct(t *testing.T) {
	// We can test if the struct holds the correct fields
	// and if NewConsumer returns the correct error when nats is down
	consumer, err := NewConsumer("nats://invalid:4222", nil)
	assert.Error(t, err)
	assert.Nil(t, consumer)
}
