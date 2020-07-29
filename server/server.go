package main

import (
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/acme/autocert"

	"github.com/gorilla/websocket"
	"github.com/kr/pty"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Mostly adapted from checkSameOrigin in gorilla/websocket
		origin := r.Header["Origin"]
		if len(origin) == 0 {
			return true
		}
		u, err := url.Parse(origin[0])
		if err != nil {
			return false
		}
		if strings.HasPrefix(u.Host, "localhost:") && strings.HasPrefix(r.Host, "localhost:") {
			return true
		}
		return u.Host == r.Host
	},
}

func shellHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	wrapper := &websocketWrapper{conn}

	// Disable carves table due to potential for file exfiltration. We also
	// use an AppArmor config on the server to prevent malicious activity
	// from being carried out through osqueryd. Dear reader, do you see any
	// vulnerabilities here? Please let us know.
	cmd := exec.Command("osqueryd", "-S", "--disable_tables=carves,curl")

	// TODO: Expose errors appropriately
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		// Cleanup both pty and osqueryd process
		_ = ptmx.Close()
		_ = cmd.Process.Kill()
		// Wait must be called in order to remvoe the zombie process
		_ = cmd.Wait()
	}()

	// Ensure that either a termination of the websocket or the osqueryd
	// process causes this function to return and the rest of the cleanup
	// to take place.
	waitchan := make(chan struct{}, 1)
	go func() {
		_, _ = io.Copy(wrapper, ptmx)
		waitchan <- struct{}{}
	}()
	go func() {
		_, _ = io.Copy(ptmx, wrapper)
		waitchan <- struct{}{}
	}()
	<-waitchan
}

func redirectHTTP(w http.ResponseWriter, r *http.Request) {
	url := "https://" + r.Host + r.URL.String()
	http.Redirect(w, r, url, http.StatusFound)
}

func main() {
	var (
		addr = *flag.String("addr", os.Getenv("ADDR"), "Address for server to bind")
		cert = *flag.String("cert", os.Getenv("CERT"), "Path to TLS certificate")
		key  = *flag.String("key", os.Getenv("KEY"), "Path to TLS private Key")
		env  = *flag.String("environment", os.Getenv("ENVIRONMENT"), "Set to PRODUCTION to enable auto-tls with letsencrypt")
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
	} else if env == "PRODUCTION" {
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("osquery.live"),
			Cache:      autocert.DirCache("/etc/letsencrypt/autocert"),
		}

		server := &http.Server{
			Addr: addr,
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
			},
		}

		go func() {
			log.Fatal(http.ListenAndServe(":http", certManager.HTTPHandler(nil)))
		}()

		log.Fatal(server.ListenAndServeTLS("", ""))

	} else {
		log.Printf("Starting server listening at http://%s", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
	}
}
