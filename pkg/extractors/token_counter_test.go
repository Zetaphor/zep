package extractors

import (
	"context"
	"testing"

	"github.com/danielchalef/zep/internal"

	"github.com/danielchalef/zep/pkg/memorystore"
	"github.com/danielchalef/zep/pkg/models"
	"github.com/danielchalef/zep/test"
	"github.com/stretchr/testify/assert"
)

func TestTokenCountExtractor(t *testing.T) {
	ctx := context.Background()

	db := memorystore.NewPostgresConn(test.TestDsn)
	defer db.Close()
	memorystore.CleanDB(t, db)

	cfg, err := test.NewTestConfig()
	assert.NoError(t, err)

	appState := &models.AppState{Config: cfg}

	store, err := memorystore.NewPostgresMemoryStore(appState, db)
	assert.NoError(t, err)

	appState.MemoryStore = store

	sessionID, err := test.GenerateRandomSessionID(16)
	assert.NoError(t, err)

	err = store.PutMemory(
		ctx,
		appState,
		sessionID,
		&models.Memory{Messages: test.TestMessages[:5]},
	)
	assert.NoError(t, err)

	memories, err := store.GetMemory(ctx, appState, sessionID, 0)
	assert.NoError(t, err)

	messages := memories.Messages

	messageEvent := &models.MessageEvent{
		SessionID: sessionID,
		Messages:  messages,
	}

	tokenCountExtractor := NewTokenCountExtractor()

	err = tokenCountExtractor.Extract(ctx, appState, messageEvent)
	assert.NoError(t, err)

	memory, err := appState.MemoryStore.GetMemory(ctx, appState, messageEvent.SessionID, 0)
	assert.NoError(t, err)
	assert.Equal(t, len(memory.Messages), len(messages))

	// reverse order since select orders LIFO
	internal.ReverseSlice(memory.Messages)

	for i := range memory.Messages {
		assert.NotZero(t, memory.Messages[i].TokenCount)
		assert.True(t, memory.Messages[i].TokenCount > 0)
	}
}