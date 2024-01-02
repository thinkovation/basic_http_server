package main

import (
	"context"
	"fmt"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// "github.com/gorilla/handlers"
type myServer struct {
	http.Server
	shutdownReq chan bool
}

// NewServer - this is the init function for the server process
func NewServer(port string) *myServer {

	//create server
	s := &myServer{
		Server: http.Server{
			Addr:         "localhost:" + port,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		shutdownReq: make(chan bool),
	}

	router := mux.NewRouter()

	//register handlers
	router.HandleFunc("/msauth", s.RootHandler)

	// HTML Assets
	sh := http.StripPrefix("/msauth/", http.FileServer(http.Dir("./assets/html_public/")))
	router.PathPrefix("/msauth/").Handler(sh)

	// Main endpoint
	//router.HandleFunc("/msauth/V01", ingest.Handler).Methods("PUT")

	// CORS stuff
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "X-API-KEY", "X-Request-Token", "Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})
	s.Handler = handlers.CORS(headersOk, originsOk, methodsOk)(router)

	return s
}

func (s *myServer) WaitShutdown() {
	irqSig := make(chan os.Signal, 1)
	signal.Notify(irqSig, syscall.SIGINT, syscall.SIGTERM)

	//Wait interrupt or shutdown request through /shutdown
	select {
	case sig := <-irqSig:
		Logger.Info(fmt.Sprintf("Shutdown request (signal: %v)", sig))
	case sig := <-s.shutdownReq:
		Logger.Info(fmt.Sprintf("Shutdown request (/shutdown %v)", sig))
	}
	Logger.Info("Stopping API server ...")

	//Create shutdown context with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//shutdown the server
	err := s.Shutdown(ctx)
	if err != nil {
		Logger.Error(fmt.Sprintf("Shutdown request error: %v", err))
	}

}

func (s *myServer) RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("msauth - see /ingest/V01/swaggerui/ for documentation\n"))
}

func GetTokenFromRequest(r *http.Request) string {
	var tmptoken string
	tmptoken = r.Header.Get("X-API-KEY")
	if tmptoken != "" {
		return tmptoken
	}
	tmptoken = r.URL.Query().Get("authtoken")
	if tmptoken != "" {
		return tmptoken
	}

	tmptoken = r.Header.Get("wf-tkn")
	if tmptoken == "" {
		tmptoken = r.URL.Query().Get("wf_tkn")
	}

	return tmptoken
}
