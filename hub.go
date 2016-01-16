package hub

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	ReadBufferSize  int = 1024
	WriteBufferSize int = 1024
)

type Hub struct {
	sync.Mutex                              // protects connections
	connections map[*connection]*subscriber // simple to close or remove connections
	subscribers map[string]*subscriber      // simple to find subscriber based on identifier string

	log            *log.Logger
	allowedOrigins []string
	wsConnFactory  websocket.Upgrader // websocket connection upgrader
	register       chan *connection   // register new connection on this channel
	unregister     chan *connection   // unregister connection channel
	subscribe      chan *subscription // subscribe as user

	Broadcast chan *Message     // fan out message to all connections
	Mailbox   chan *MailMessage // fan out message to subscriber

	SubscriptionTokenizer Tokenizer // for user subscription token validation
}

func New(logOutput io.Writer, origins ...string) *Hub {
	h := &Hub{
		allowedOrigins: origins,
		register:       make(chan *connection),
		unregister:     make(chan *connection),
		connections:    make(map[*connection]*subscriber),
		subscribers:    make(map[string]*subscriber),
		subscribe:      make(chan *subscription),
		Broadcast:      make(chan *Message),
		Mailbox:        make(chan *MailMessage),
	}

	factory := websocket.Upgrader{
		ReadBufferSize:  ReadBufferSize,
		WriteBufferSize: WriteBufferSize,
		CheckOrigin:     h.checkOrigin,
	}

	h.wsConnFactory = factory

	if nil == logOutput {
		logOutput = ioutil.Discard
	}
	h.log = log.New(leveledLogWriter(logOutput), "", log.LstdFlags)

	return h
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	ws, err := h.wsConnFactory.Upgrade(w, r, nil)
	if err != nil {
		h.log.Println("[ERROR] failed to upgrade connection:", err)
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws, hub: h}
	h.register <- c
	go c.listenWrite()
	c.listenRead()
}

func (h *Hub) Connections() int {
	h.Lock()
	defer h.Unlock()

	return len(h.connections)
}

func (h *Hub) doRegister(c *connection) {
	h.Lock()
	defer h.Unlock()

	h.connections[c] = nil
}

func (h *Hub) doMailbox(m *MailMessage) {
	h.Lock()
	defer h.Unlock()

	s, ok := h.subscribers[m.Username]
	if !ok {
		h.log.Println("[DEBUG] there are no subscriptions from:", m.Username)
		return
	}

	bytes, err := m.Message.bytes()
	if err != nil {
		h.log.Printf("[WARN] failed to marshal message: %+v, reason: %s\n", m, err)
		return
	}

	h.log.Println("[DEBUG] subscriber connection count:", len(s.connections))
	for c := range s.connections {
		c.send <- bytes
	}
}

func (h *Hub) doBroadcast(m *Message) {
	h.Lock()
	defer h.Unlock()

	bytes, err := m.bytes()
	if err != nil {
		h.log.Printf("[WARN] failed to marshal message: %+v, reason: %s\n", m, err)
		return
	}
	for c := range h.connections {
		c.send <- bytes
	}
}

func (h *Hub) doSubscribe(s *subscription) {
	h.Lock()
	defer h.Unlock()

	if h.SubscriptionTokenizer == nil {
		h.log.Println("[DEBUG] subscription tokenizer is not set, cannot validate subscriptions")
		return
	}

	token := h.SubscriptionTokenizer.Tokenize(s.Username)
	if token != s.Token {
		h.log.Printf("[WARN] username [%s], token [%s] does not match given: [%s]\n", s.Username, token, s.Token)
		return
	}

	// check if there already is a subscriber
	ns, alreadyAvailable := h.subscribers[s.Username]
	if !alreadyAvailable {
		ns = &subscriber{
			Username:    s.Username,
			connections: make(map[*connection]bool),
		}
	}

	ns.connections[s.connection] = true
	h.connections[s.connection] = ns
	h.subscribers[ns.Username] = ns
	h.log.Println("[DEBUG] subscribed as:", s.Username)
}

func (h *Hub) doUnregister(c *connection) {
	h.Lock()
	defer h.Unlock()

	s, ok := h.connections[c]
	if !ok {
		h.log.Println("[WARN] cannot unregister connection, it is not registered.")
		return
	}

	if s != nil {
		delete(s.connections, c)
		h.log.Printf("[DEBUG] unregistering one of subscribers: %s connections\n", s.Username)
		if len(s.connections) == 0 {
			// there are no more open connections for this subscriber
			h.log.Printf("[DEBUG] unsubscribe: %s, no more open connections\n", s.Username)
			delete(h.subscribers, s.Username)
		}
	}

	h.log.Println("[DEBUG] unregistering socket connection")
	c.close()
	delete(h.connections, c)
}

func (h *Hub) checkOrigin(r *http.Request) bool {
	origin := r.Header["Origin"]
	if len(origin) == 0 {
		return true
	}
	u, err := url.Parse(origin[0])
	if err != nil {
		return false
	}
	var allow bool
	for _, o := range h.allowedOrigins {
		if o == u.Host {
			allow = true
			break
		}
		if o == "*" {
			allow = true
			break
		}
	}
	if !allow {
		h.log.Printf("[DEBUG] none of allowed origins: %s matched: %s\n", strings.Join(h.allowedOrigins, ", "), u.Host)
	}
	return allow
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.doRegister(c)
		case c := <-h.unregister:
			h.doUnregister(c)
		case m := <-h.Broadcast:
			h.doBroadcast(m)
		case m := <-h.Mailbox:
			h.doMailbox(m)
		case s := <-h.subscribe:
			h.doSubscribe(s)
		}
	}
}
