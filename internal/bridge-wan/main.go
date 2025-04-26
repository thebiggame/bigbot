package bridge_wan

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/thebiggame/bigbot/internal/bridge-wan/web"
	"github.com/thebiggame/bigbot/internal/config"
	protodef "github.com/thebiggame/bigbot/proto"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type BridgeWAN struct {
	// The context given to us by the main bot.
	ctx *context.Context
	// The HTTP server.
	httpServer *http.Server

	// The authentication key to test WS connections against.
	wsKey string

	// Active event websocket connection.
	// Potentially nil if the event is not connected - do EventAvailable() on the bridge to determine status.
	wsConn *websocket.Conn

	// Stores response handlers for given request IDs. (the mutex MUST be held to interact with this map)
	wsResponseCh  map[string]chan *protodef.RPCResponse
	wsResponseMtx sync.Mutex
}

// logger stores the module's logger instance.
var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

var EventBridge *BridgeWAN

func generateRequestID() string {
	// Generate a unique request ID
	id, err := uuid.NewV4()
	if err != nil {
		// This should never happen.
		panic(err)
	}
	return id.String()
}

func BridgeIsAvailable() (up bool) {
	if EventBridge != nil {
		return EventBridge.EventAvailable()
	}
	return false
}

func New() (bridge *BridgeWAN, err error) {
	bridge = &BridgeWAN{
		httpServer:   &http.Server{Addr: config.RuntimeConfig.Bridge.Address},
		wsKey:        string(config.RuntimeConfig.Bridge.Key),
		wsResponseCh: make(map[string]chan *protodef.RPCResponse),
	}
	EventBridge = bridge
	return bridge, nil
}

func (mod *BridgeWAN) SetLogger(log *slog.Logger) {
	logger = log
}

func (mod *BridgeWAN) Run() (err error) {
	// Create app context (this is passed to modules).
	// The signal.NotifyContext is a special context that gets torn down when an interrupt / SIGTERM is received.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = mod.Start(ctx)

	if err != nil {
		return err
	}
	return
}

func (mod *BridgeWAN) Start(ctx context.Context) (err error) {
	if !config.RuntimeConfig.Bridge.Enabled {
		return nil
	}

	mod.ctx = &ctx

	tmpl, err := template.ParseFS(web.Templates, "*.html")
	if err != nil {
		return err
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := web.IndexPageData{
			Version: config.AppVersion,
		}
		err := tmpl.ExecuteTemplate(w, "index.html", data)
		if err != nil {
			logger.Error("error rendering template", slog.Any("err", err))
		}
	})

	// Websocket handler
	http.HandleFunc("/ws", mod.wsHandle)
	go func() {
		logger.Info(fmt.Sprintf("BIGbridge SERVER listening at %v", mod.httpServer.Addr))
		if err := mod.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("error with HTTP server", slog.Any("error", err))
		}
	}()

	for {
		select {
		// Spinloop here to make sure that we stay alive long enough for the context to get torn down properly.
		case <-ctx.Done():
			// Do clean shutdown.
			if err := mod.httpServer.Shutdown(context.Background()); err != nil {
				logger.Error("error shutting down http server", slog.Any("error", err))
			}
			return ctx.Err()
		}
	}
}
