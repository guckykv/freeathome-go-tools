// sysapprobe measures how a System Access Point treats websocket clients:
// which keepalive it tolerates, how many connections it serves at once, and how
// long a handshake takes. Run it after a firmware update to check whether the
// assumptions the fahapi library builds on still hold.
//
// It deliberately speaks the protocol directly instead of going through the
// library, because the point is to test the layer the library sits on.
//
// Findings against software 2.6, which the library relies on:
//
//   - A connection with no keepalive stays open; pings are for detecting a dead
//     link, not for keeping the SysAP interested.
//   - Ping frames are answered reliably.
//   - A single text frame closes every websocket client of the SysAP, not just
//     the sender. The modes that send one are therefore disruptive and refuse
//     to run without -disruptive.
package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tkanos/gonfig"
)

type Configuration struct {
	Host     string `env:"FHAPI_HOST"`     // local IP of the SysAP
	Username string `env:"FHAPI_USER"`     // username comes from free@home app
	Password string `env:"FHAPI_PASSWORD"` // pw is the same like you have used in the free@home app
}

var (
	configFile = flag.String("c", "~/.fahapi-config.json", "configuration file")
	mode       = flag.String("mode", "ping", "none | ping | text | text-once | ladder | collateral")
	interval   = flag.Duration("interval", 20*time.Second, "keepalive interval for ping and text")
	duration   = flag.Duration("d", 60*time.Second, "how long to hold a connection")
	count      = flag.Int("n", 8, "ladder: how many connections to open")
	stagger    = flag.Duration("stagger", 15*time.Second, "ladder: delay between connections")
	disruptive = flag.Bool("disruptive", false, "allow modes that disturb every other client")

	configuration Configuration
)

// disruptiveModes send a text frame, which drops every websocket client the
// SysAP has -- the free@home app included.
var disruptiveModes = map[string]bool{"text": true, "text-once": true, "collateral": true}

func main() {
	flag.Usage = usage
	flag.Parse()
	initialize()

	switch *mode {
	case "ladder":
		runLadder()
	case "collateral":
		runCollateral()
	default:
		fmt.Printf("%-10s | %s\n", *mode, hold(*mode, *interval, *duration))
	}
}

func initialize() {
	if strings.HasPrefix(*configFile, "~/") {
		usr, err := user.Current()
		if err != nil {
			log.Fatalf("home directory: %v", err)
		}
		*configFile = filepath.Join(usr.HomeDir, (*configFile)[2:])
	}
	if err := gonfig.GetConf(*configFile, &configuration); err != nil {
		log.Fatalf("read config %s: %v", *configFile, err)
	}
	if configuration.Host == "" {
		log.Fatal("no Host configured")
	}
	if disruptiveModes[*mode] && !*disruptive {
		log.Fatalf("mode %q sends a text frame, which disconnects every websocket "+
			"client of this SysAP -- including the free@home app. Pass -disruptive "+
			"if that is acceptable right now.", *mode)
	}
}

// result records what became of one connection.
type result struct {
	handshake time.Duration
	lifetime  time.Duration
	messages  int
	pongs     int
	err       error
}

func (r result) String() string {
	reason := "-"
	if r.err != nil {
		reason = r.err.Error()
	}
	return fmt.Sprintf("handshake %5.1fs | held %6.1fs | messages %3d | pongs %3d | %s",
		r.handshake.Seconds(), r.lifetime.Seconds(), r.messages, r.pongs, reason)
}

// hold opens one connection, applies the given keepalive and reports what
// happened. It returns early if the SysAP closes the connection.
func hold(keepalive string, every, d time.Duration) result {
	header := http.Header{}
	header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString(
		[]byte(configuration.Username+":"+configuration.Password)))

	dialer := &websocket.Dialer{HandshakeTimeout: 30 * time.Second}

	dialStart := time.Now()
	conn, _, err := dialer.Dial("ws://"+configuration.Host+"/fhapi/v1/api/ws", header)
	if err != nil {
		return result{handshake: time.Since(dialStart), err: err}
	}
	defer conn.Close()

	var mu sync.Mutex
	res := result{handshake: time.Since(dialStart)}
	start := time.Now()

	conn.SetPongHandler(func(string) error {
		mu.Lock()
		defer mu.Unlock()
		res.pongs++
		return nil
	})

	closed := make(chan struct{})
	go func() {
		defer close(closed)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				mu.Lock()
				res.err = err
				mu.Unlock()
				return
			}
			mu.Lock()
			res.messages++
			mu.Unlock()
		}
	}()

	// One text frame, to tell "any text frame is fatal" from "this rate is".
	if keepalive == "text-once" {
		time.AfterFunc(2*time.Second, func() {
			conn.WriteMessage(websocket.TextMessage, []byte(time.Now().String()))
		})
	}

	var keepaliveTick <-chan time.Time
	if keepalive == "ping" || keepalive == "text" {
		ticker := time.NewTicker(every)
		defer ticker.Stop()
		keepaliveTick = ticker.C
	}

	finish := func() result {
		mu.Lock()
		defer mu.Unlock()
		res.lifetime = time.Since(start)
		return res
	}

	deadline := time.After(d)
	for {
		select {
		case <-deadline:
			return finish()
		case <-closed:
			return finish()
		case now := <-keepaliveTick:
			var err error
			if keepalive == "ping" {
				err = conn.WriteMessage(websocket.PingMessage, nil)
			} else {
				err = conn.WriteMessage(websocket.TextMessage, []byte(now.String()))
			}
			if err != nil {
				mu.Lock()
				res.err = err
				mu.Unlock()
				return finish()
			}
		}
	}
}

// runLadder opens connections one by one to find out whether the SysAP limits
// how many it serves, and whether reaching a limit rejects the new connection
// or drops an established one.
func runLadder() {
	fmt.Printf("ladder: up to %d connections, %s apart, holding %s\n", *count, *stagger, *duration)

	var wg sync.WaitGroup
	results := make([]result, *count)

	for i := 0; i < *count; i++ {
		remaining := *duration - time.Duration(i)*(*stagger)
		if remaining <= 0 {
			fmt.Printf("  connection %d skipped: -d is too short for %d connections %s apart\n",
				i+1, *count, *stagger)
			continue
		}
		wg.Add(1)
		go func(i int, d time.Duration) {
			defer wg.Done()
			results[i] = hold("ping", *interval, d)
		}(i, remaining)

		fmt.Printf("  + connection %d opened\n", i+1)
		time.Sleep(*stagger)
	}
	wg.Wait()

	fmt.Println("result:")
	for i, r := range results {
		fmt.Printf("  connection %d: %s\n", i+1, r)
	}
}

// runCollateral answers whether one client's text frame disturbs the others: an
// observer that only pings runs first, then a second connection sends one text
// frame.
func runCollateral() {
	var wg sync.WaitGroup
	var observer, sender result

	wg.Add(1)
	go func() {
		defer wg.Done()
		observer = hold("ping", 5*time.Second, 45*time.Second)
	}()

	time.Sleep(15 * time.Second)
	fmt.Println("  observer has been connected for 15s, now a second client sends one text frame")

	wg.Add(1)
	go func() {
		defer wg.Done()
		sender = hold("text-once", time.Second, 20*time.Second)
	}()
	wg.Wait()

	fmt.Printf("  observer (pings only): %s\n", observer)
	fmt.Printf("  sender   (one text):   %s\n", sender)
	if observer.err != nil {
		fmt.Println("  => the observer was taken down with it: a text frame disturbs every client")
	} else {
		fmt.Println("  => the observer survived: a text frame only affects its sender")
	}
}

func usage() {
	fmt.Printf("usage %s:\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Print(`
  Modes:
    none        hold a connection without any keepalive
    ping        keepalive with ping frames, the way the library does it
    text        keepalive with text frames, the way the library used to
    text-once   a single text frame after two seconds
    ladder      open -n connections one by one, looking for a limit
    collateral  does one client's text frame disturb the others?

  text, text-once and collateral disconnect every websocket client of the
  SysAP and need -disruptive.

  Example: sysapprobe -mode ping -interval 5s -d 60s
`)
}
