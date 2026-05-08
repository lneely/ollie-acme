// ollie-acme: acme interface for ollie AI sessions
package main

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"9fans.net/go/acme"
)

var ollie string

func init() {
	ollie = os.Getenv("OLLIE")
	if ollie == "" {
		ollie = filepath.Join(os.Getenv("HOME"), "mnt", "ollie")
	}
}

func main() {
	acme.AutoExit(true)
	if err := newSessionList(); err != nil {
		log.Fatal(err)
	}
	select {}
}

// --- session list window ---

type sessionList struct {
	win       *acme.Win
	allSIDs   []string
	mu        sync.Mutex
}

func newSessionList() error {
	w, err := acme.New()
	if err != nil {
		return err
	}
	w.Name("ollie/sessions")
	sl := &sessionList{win: w}
	sl.refresh()
	w.Ctl("cleartag")
	w.Write("tag", []byte(" New Kill Refresh"))
	go sl.eventLoop()
	return nil
}

func (sl *sessionList) refresh() {
	data, err := os.ReadFile(filepath.Join(ollie, "s", "idx"))
	if err != nil {
		sl.win.Errf("read idx: %v", err)
		return
	}
	sl.win.Addr(",")
	sl.win.Write("data", nil)

	raw := strings.TrimSpace(string(data))
	if raw == "" {
		sl.win.Ctl("clean")
		return
	}

	lines := strings.Split(raw, "\n")
	sort.Strings(lines)

	sl.mu.Lock()
	sl.allSIDs = nil
	for _, l := range lines {
		if s := strings.TrimSpace(l); s != "" {
			id, _, _ := strings.Cut(s, "\t")
			sl.allSIDs = append(sl.allSIDs, id)
		}
	}
	sl.mu.Unlock()

	sl.win.Write("body", []byte(raw+"\n"))
	sl.win.Ctl("clean")
}

func (sl *sessionList) resolveFullSID(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	sl.mu.Lock()
	defer sl.mu.Unlock()
	for _, sid := range sl.allSIDs {
		parts := strings.Split(sid, "__")
		if parts[len(parts)-1] == text {
			return sid
		}
	}
	return text
}

func (sl *sessionList) eventLoop() {
	for e := range sl.win.EventChan() {
		switch e.C2 {
		case 'x', 'X':
			cmd := strings.TrimSpace(string(e.Text))
			if cmd == "New" {
				openFile(filepath.Join(ollie, "s", "new"))
				continue
			}
			if cmd == "Kill" {
				arg := strings.TrimSpace(string(e.Arg))
				sid := sl.resolveFullSID(arg)
				if sid != "" {
					os.WriteFile(filepath.Join(ollie, "s", sid, "ctl"), []byte("kill"), 0644)
					sl.refresh()
				}
				continue
			}
			if cmd == "Refresh" {
				sl.refresh()
				continue
			}
			sid := sl.resolveFullSID(cmd)
			if sid != "" && sessionExists(sid) {
				openChat(sid)
			} else {
				sl.win.WriteEvent(e)
			}
		case 'l', 'L':
			sl.win.WriteEvent(e)
		}
	}
	os.Exit(0)
}

func sessionExists(sid string) bool {
	_, err := os.Stat(filepath.Join(ollie, "s", sid))
	return err == nil
}

// --- chat window ---

type chatWin struct {
	win    *acme.Win
	sid    string
	offset int
	mu     sync.Mutex
}

var chatWindows sync.Map

func openChat(sid string) {
	if v, ok := chatWindows.Load(sid); ok {
		v.(*chatWin).win.Ctl("show")
		return
	}

	w, err := acme.New()
	if err != nil {
		log.Printf("open chat %s: %v", sid, err)
		return
	}
	w.Name("%s", "ollie/sessions/"+sid)
	w.Ctl("cleartag")
	w.Fprintf("tag", " Prompt Stop Ctl")

	cw := &chatWin{win: w, sid: sid}
	chatWindows.Store(sid, cw)

	// Load existing chat.
	chatPath := filepath.Join(ollie, "s", sid, "chat")
	if data, err := os.ReadFile(chatPath); err == nil && len(data) > 0 {
		w.Write("body", data)
		cw.offset = len(data)
	}
	w.Ctl("clean")

	go cw.tail()
	go cw.eventLoop()
}

func (cw *chatWin) eventLoop() {
	defer chatWindows.Delete(cw.sid)
	for e := range cw.win.EventChan() {
		switch e.C2 {
		case 'x', 'X':
			cmd := strings.TrimSpace(string(e.Text))
			switch cmd {
			case "Prompt":
				openFile(filepath.Join(ollie, "s", cw.sid, "prompt"))
			case "Stop":
				os.WriteFile(filepath.Join(ollie, "s", cw.sid, "ctl"), []byte("stop"), 0644)
			case "Ctl":
				openFile(filepath.Join(ollie, "s", cw.sid, "ctl"))
			default:
				cw.win.WriteEvent(e)
			}
		case 'l', 'L':
			cw.win.WriteEvent(e)
		}
	}
}

// tail polls the chat file for new content and streams it to the window.
func (cw *chatWin) tail() {
	chatPath := filepath.Join(ollie, "s", cw.sid, "chat")

	for {
		fi, err := os.Stat(chatPath)
		if err != nil {
			if os.IsNotExist(err) {
				return
			}
			time.Sleep(50 * time.Millisecond)
			continue
		}
		size := int(fi.Size())
		cw.mu.Lock()
		off := cw.offset
		cw.mu.Unlock()
		if size > off {
			f, err := os.Open(chatPath)
			if err == nil {
				buf := make([]byte, size-off)
				n, _ := f.ReadAt(buf, int64(off))
				f.Close()
				if n > 0 {
					cw.mu.Lock()
					cw.win.Addr("$")
					cw.win.Write("data", buf[:n])
					cw.offset += n
					cw.win.Ctl("clean")
					cw.mu.Unlock()
				}
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// openFile opens a 9P/filesystem path as a plain acme editor window.
func openFile(path string) {
	w, err := acme.New()
	if err != nil {
		log.Printf("open %s: %v", path, err)
		return
	}
	w.Name("%s", path)
	w.Ctl("get")
}
