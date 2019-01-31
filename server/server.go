package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gorilla/websocket"
	"github.com/kr/pty"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func shellHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	wrapper := &websocketWrapper{conn}
	_ = wrapper

	// Disable carves table due to potential for file exfiltration
	cmd := exec.Command("osqueryi", "--disable_tables=carves")

	// TODO: Add error handling
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Println(err)
		return
	}
	defer func() { _ = ptmx.Close() }()

	go func() { _, _ = io.Copy(ptmx, wrapper) }()
	_, _ = io.Copy(wrapper, ptmx)

	ptmx.Close()
	cmd.Wait()
}

func redirectHTTP(w http.ResponseWriter, r *http.Request) {
	url := "https://" + r.Host + r.URL.String()
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func main() {
	var (
		addr = *flag.String("addr", os.Getenv("ADDR"), "Address for server to bind")
		cert = *flag.String("cert", os.Getenv("CERT"), "Path to TLS certificate")
		key  = *flag.String("key", os.Getenv("KEY"), "Path to TLS private Key")
	)

	http.HandleFunc("/shell", shellHandler)
	static := http.FileServer(http.Dir("build"))
	http.Handle("/", static)

	if cert != "" && key != "" {
		// Redirect HTTP to HTTPS
		go func() {
			log.Fatal(http.ListenAndServe(":80", http.HandlerFunc(redirectHTTP)))
		}()

		log.Printf("Starting server listening at https://%s", addr)
		log.Fatal(http.ListenAndServeTLS(addr, cert, key, nil))
	} else {
		log.Printf("Starting server listening at http://%s", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
	}
}
