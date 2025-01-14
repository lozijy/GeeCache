package http

import (
	"GeeCache/group"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

// HTTPPool implements PeerPicker for a pool of HTTP peers
type HTTPPool struct {
	self        string
	basePath    string
	mu          sync.Mutex
	peers       *Map
	httpGetters map[string]*httpGetter
}

// NewHTTPPool initializes an HTTP pool of peers
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log info with server name
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all http requests
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key>
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := group.GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

// Set updates the pool's list of peers
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = NewMap(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer picks a peer according to key
func (p *HTTPPool) PickPeer(key string) (group.PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

// ListenAndServe starts a HTTP server on the given address
func (p *HTTPPool) ListenAndServe() error {
	return ListenAndServe(p.self, p)
}

// PeerGetter is the interface that must be implemented by a peer
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}

// PeerPicker is the interface that must be implemented to locate
// the peer that owns a specific key
type PeerPicker interface {
	PickPeer(key string) (peer group.PeerGetter, ok bool)
}

var _ group.PeerGetter = (*httpGetter)(nil)
var _ group.PeerPicker = (*HTTPPool)(nil)

// Handle registers the handler for the given pattern
func Handle(pattern string, handler http.Handler) {
	http.Handle(pattern, handler)
}

// HandlerFunc type is an adapter to allow the use of ordinary functions as HTTP handlers
type HandlerFunc = http.HandlerFunc

// ResponseWriter interface is used by an HTTP handler to construct an HTTP response
type ResponseWriter = http.ResponseWriter

// Request represents an HTTP request received by a server
type Request = http.Request

// Error replies to the request with the specified error message and HTTP code
func Error(w ResponseWriter, error string, code int) {
	http.Error(w, error, code)
}

// StatusInternalServerError represents an internal server error (HTTP 500)
const StatusInternalServerError = http.StatusInternalServerError

// ListenAndServe starts a HTTP server
func ListenAndServe(addr string, handler http.Handler) error {
	// Remove http:// prefix if present
	if len(addr) > 7 && addr[:7] == "http://" {
		addr = addr[7:]
	}
	return http.ListenAndServe(addr, handler)
}
