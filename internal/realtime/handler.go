package realtime

import (
	"context"
	"net/http"
	"time"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/apihelper"
	"github.com/Dimes2212/5th_semester_course_work_backend/internal/requestctx"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const websocketWriteTimeout = 5 * time.Second

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := requestctx.CurrentUserFromContext(r.Context())
	if !ok {
		apihelper.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	connection, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer connection.CloseNow()

	client := NewClient(currentUser.UserID)
	h.hub.Register(client)
	defer h.hub.Unregister(client)

	connectionContext := connection.CloseRead(r.Context())

	for {
		select {
		case <-connectionContext.Done():
			return
		case <-h.hub.Done():
			_ = connection.Close(
				websocket.StatusGoingAway,
				"server shutting down",
			)
			return

		case event := <-client.send:
			writeContext, cancel := context.WithTimeout(connectionContext, websocketWriteTimeout)

			err := wsjson.Write(writeContext, connection, event)
			cancel()

			if err != nil {
				return
			}
		}
	}
}
