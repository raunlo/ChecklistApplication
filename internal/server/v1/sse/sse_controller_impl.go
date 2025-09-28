package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	notification "com.raunlo.checklist/internal/core/notification"
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

	ch := s.broker.Subscribe(request.ChecklistId)
	defer s.broker.Unsubscribe(request.ChecklistId, ch)

	// Send a comment to establish the stream
	_, _ = w.Write([]byte(":ok\n\n"))
	flusher.Flush()

	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return nil, nil
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
