package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/securecookie"
	"golang.org/x/time/rate"
)

var SECRET_KEY = make([]byte, 64)
var s = securecookie.New(SECRET_KEY, nil)

var index_tmpl = template.Must(template.ParseFiles("templates/index.html"))

type Visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var visitors = make(map[string]*Visitor)
var mu sync.Mutex

type DNSRequest struct {
	Domain string `json:"domain"`
	Type   string `json:"type"`
	Server string `json:"server"`
}

func main() {
/*
	file, f_err := os.Open("SECRET_KEY")
	if f_err != nil {
		_, e := rand.Read(SECRET_KEY)
		f, err := os.Create("SECRET_KEY")
		_, e1 := f.Write(SECRET_KEY)
		if e != nil || err != nil || e1 != nil {
			log.Fatal(e, err, e1)
		}

	} else {
		_, err2 := file.Read(SECRET_KEY)
		if err2 != nil {
			_, e := rand.Read(SECRET_KEY)
			f, err := os.Create("data/SECRET_KEY")
			_, e1 := f.Write(SECRET_KEY)
			if e != nil || err != nil || e1 != nil {
				log.Fatal(e)
			}

		}
	}
*/
	var host = read_file_as_str("HOST")

	if host == "" {
		host = "0.0.0.0:80"
		f, err := os.Create("HOST")
		_, e1 := f.Write([]byte(`0.0.0.0:80`))
		if err != nil || e1 != nil {
			log.Fatal(err, e1, err)
		}
	}

	s = securecookie.New(SECRET_KEY, nil)

	// file.Close()

	http.HandleFunc("/", index)
	http.Handle("/resolve/", rateLimiter(http.HandlerFunc(resolve_handle)))
	fs := http.FileServer(http.Dir("."))
	http.Handle("/static/", fs)
	go cleanupVisitors()
	fmt.Printf("\n\nRunning at %s\n\n", host)
	log.Fatal(http.ListenAndServe(host, nil))

}

func index(w http.ResponseWriter, r *http.Request) {
	index_tmpl = template.Must(template.ParseFiles("templates/index.html"))
	w.Header().Add("Server", "Dresol (golang net/http)")
	w.Header().Add("Privet", "chitatel!")
	index_tmpl.Execute(w, nil)
}

func resolve_handle(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}

	data, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(500)
		return
	}

	payload := DNSRequest{}
	err = json.Unmarshal(data, &payload)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(422)
		return
	}

	var result []byte
	result, err = resolve(payload.Domain, payload.Type, payload.Server)

	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")
	w.Write(result)
}

func read_file_as_str(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func() {
		if err = file.Close(); err != nil {

		}
	}()

	b, err1 := io.ReadAll(file)

	if err1 != nil {
		return ""
	}

	return string(b)
}

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]

	if !exists {
		limiter := rate.NewLimiter(1, 4)

		visitors[ip] = &Visitor{
			limiter:  limiter,
			lastSeen: time.Now(),
		}

		return limiter
	}

	v.lastSeen = time.Now()

	return v.limiter
}

func rateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Internal error", 500)
			return
		}

		limiter := getVisitor(ip)

		if !limiter.Allow() {
			http.Error(w, "Too many requests", 429)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func cleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		mu.Lock()

		for ip, visitor := range visitors {
			if time.Since(visitor.lastSeen) > 3*time.Minute {
				delete(visitors, ip)
			}
		}

		mu.Unlock()
	}
}
