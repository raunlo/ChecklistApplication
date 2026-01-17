package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	notification "com.raunlo.checklist/internal/core/notification"
	serverutils "com.raunlo.checklist/internal/server/server_utils"
	"github.com/gin-gonic/gin"
)

type ISSEController = StrictServerInterface

func NewSSEController(broker notification.IBroker) ISSEController {
	return &sseControllerImpl{broker: broker, mapper: NewChecklistItemUpdatesMapper()}
}

type sseControllerImpl struct {
	broker notification.IBroker
	mapper IChecklistItemUpdatesMapper
}

// GetEventsStreamForChecklistItems streams SSE events for the given checklistId.
// The generated strict handler forwards a *gin.Context as the ctx parameter,
// so we assert and use it to access the ResponseWriter and Request.
func (s *sseControllerImpl) GetEventsStreamForChecklistItems(ctx context.Context, request GetEventsStreamForChecklistItemsRequestObject) (GetEventsStreamForChecklistItemsResponseObject, error) {
	gctx, ok := ctx.(*gin.Context)
	if !ok {
		return nil, fmt.Errorf("expected *gin.Context in StrictServerInterface, got %T", ctx)
	}

	domainContext := serverutils.CreateContext(ctx) // to ensure any middleware has run

	w := gctx.Writer
	r := gctx.Request

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return nil, nil
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return nil, nil
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// attempt to read client id from Gin context (set by middleware)
	clientId := gctx.GetString("clientId")
	// middleware sets Gin context key "clientId"; use that value if present

	ch, err := s.broker.Subscribe(domainContext, request.ChecklistId)
	if err != nil {
		http.Error(w, "Failed to subscribe to events", http.StatusInternalServerError)
		return nil, nil
	}
	defer s.broker.Unsubscribe(domainContext, request.ChecklistId)

	// keep clientId variable to avoid compiler unused var error (for future use)
	_ = clientId

	// Send a comment to establish the stream
	_, _ = w.Write([]byte(":ok\n\n"))
	flusher.Flush()

	// Heartbeat to keep connection alive and detect dropped clients
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return nil, nil
		case <-heartbeat.C:
			// Send heartbeat comment to keep connection alive
			_, _ = w.Write([]byte(":heartbeat\n\n"))
			flusher.Flush()
		case msg, ok := <-ch:
			if !ok {
				return nil, nil
			}

			// marshal mapped event envelope to JSON and write as SSE data field
			mapped := s.mapper.Map(msg)
			b, err := json.Marshal(mapped)
			if err != nil {
				// on marshal error, send a comment and continue
				_, _ = w.Write([]byte(":error\n\n"))
				flusher.Flush()
				continue
			}

			_, _ = w.Write(append([]byte("data:"), b...))
			_, _ = w.Write([]byte("\n\n"))
			flusher.Flush()
		}
	}
}
