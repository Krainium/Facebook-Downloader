// Facebook video and reel downloader.
//
// A share link is resolved by following it with a full browser header set.
// Facebook varies on Sec-Fetch-Site and Sec-Fetch-Mode and answers 400 when
// they are missing, which is why a plain fetch fails. The resolved page embeds
// the media urls as browser_native_hd_url and browser_native_sd_url, so no API
// key and no login are involved. Downloads stream back through this server.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const browserUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
	"(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

var (
	hdURL   = regexp.MustCompile(`"browser_native_hd_url"\s*:\s*"([^"]+)"`)
	sdURL   = regexp.MustCompile(`"browser_native_sd_url"\s*:\s*"([^"]+)"`)
	ogImage = regexp.MustCompile(`<meta property="og:image" content="([^"]+)"`)
	ogTitle = regexp.MustCompile(`<meta property="og:title" content="([^"]+)"`)
	videoID = regexp.MustCompile(`/(?:reel|videos|watch)/(?:\?v=)?(\d{6,})`)
	objID   = regexp.MustCompile(`"(?:video_id|id)"\s*:\s*"?(\d{14,17})"?`)
)

// wire types
type Variant struct {
	URL      string `json:"url"`
	Label    string `json:"label"` // HD or SD
	Filename string `json:"filename"`
}

type Media struct {
	Type     string    `json:"type"` // video
	URL      string    `json:"url"`  // best rendition
	Thumb    string    `json:"thumb"`
	Ext      string    `json:"ext"`
	Filename string    `json:"filename"`
	Variants []Variant `json:"variants,omitempty"`
}

type Post struct {
	ID      string  `json:"id"`
	Title   string  `json:"title"`
	PageURL string  `json:"pageUrl"`
	Media   []Media `json:"media"`
}

// http
func newClient() *http.Client {
	tr := &http.Transport{
		MaxIdleConns:        64,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 15 * time.Second,
	}
	if p := os.Getenv("FDL_PROXY"); p != "" {
		if u, err := url.Parse(p); err == nil {
			tr.Proxy = http.ProxyURL(u)
			log.Printf("upstream proxy: %s", u.Redacted())
		}
	}
	return &http.Client{Transport: tr, Timeout: 90 * time.Second}
}

var client = newClient()

// browserHeaders is what makes Facebook answer at all. Without the Sec-Fetch
// trio it returns a 400 error page regardless of the url or the source address.
func browserHeaders(req *http.Request) {
	req.Header.Set("User-Agent", browserUA)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
}

// extraction
func validLink(raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return errors.New("that does not look like a Facebook link")
	}
	h := strings.ToLower(u.Hostname())
	ok := strings.HasSuffix(h, "facebook.com") || strings.HasSuffix(h, "fb.watch") ||
		strings.HasSuffix(h, "fb.com")
	if !ok {
		return errors.New("only facebook.com and fb.watch links are supported")
	}
	return nil
}

// unescapeJSON turns an embedded JSON string literal back into a usable url.
func unescapeJSON(s string) string {
	if out, err := strconv.Unquote(`"` + s + `"`); err == nil {
		return out
	}
	// Fall back to the two escapes that actually appear in these payloads.
	s = strings.ReplaceAll(s, `\/`, "/")
	s = strings.ReplaceAll(s, `%`, "%")
	return s
}

func safeName(s string) string {
	s = strings.Map(func(r rune) rune {
		if strings.ContainsRune(`\/:*?"<>|`, r) || r < 32 {
			return '_'
		}
		return r
	}, s)
	if len(s) > 60 {
		s = s[:60]
	}
	s = strings.TrimSpace(s)
	if s == "" {
		s = "facebook"
	}
	return s
}

func firstMatch(re *regexp.Regexp, body []byte) string {
	if m := re.FindSubmatch(body); m != nil {
		// og: meta content is html escaped, so an apostrophe arrives as
		// &#x2019; and would otherwise be shown to the reader verbatim.
		return html.UnescapeString(unescapeJSON(string(m[1])))
	}
	return ""
}

// enclosingObject returns the byte range of the JSON object that directly holds
// position p. It honours string quoting so a brace inside a caption or url does
// not throw the scan off.
func enclosingObject(b []byte, p int) (int, int) {
	end := len(b)
	depth, inStr, esc := 0, false, false
	for i := p; i < len(b); i++ {
		c := b[i]
		if inStr {
			if esc {
				esc = false
			} else if c == '\\' {
				esc = true
			} else if c == '"' {
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{':
			depth++
		case '}':
			if depth == 0 {
				end = i + 1
				i = len(b) // stop
				continue
			}
			depth--
		}
	}

	// The opening brace is the last unmatched '{' before p, found by a
	// string aware forward pass so quotes are respected the same way.
	start, inStr, esc := 0, false, false
	var stack []int
	for i := 0; i < p; i++ {
		c := b[i]
		if inStr {
			if esc {
				esc = false
			} else if c == '\\' {
				esc = true
			} else if c == '"' {
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{':
			stack = append(stack, i)
		case '}':
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	if len(stack) > 0 {
		start = stack[len(stack)-1]
	}
	return start, end
}

// objOwnsID reports whether the given object carries the requested video id.
func objOwnsID(obj []byte, wantID string) bool {
	for _, m := range objID.FindAllSubmatch(obj, -1) {
		if string(m[1]) == wantID {
			return true
		}
	}
	return false
}

// mediaForID returns the playback urls that belong to the requested reel.
//
// A Facebook page embeds the requested video plus a feed of recommendations,
// each an identical media object. Taking the first url in the document happens
// to work today only because Facebook renders the requested video first, which
// is exactly the fragile assumption to avoid. Instead each url is tied back to
// the reel id carried inside its own JSON object, so the wrong clip cannot be
// returned even if the ordering changes.
func mediaForID(body []byte, wantID string) (hd, sd string) {
	if wantID == "" {
		return "", ""
	}
	for _, m := range hdURL.FindAllSubmatchIndex(body, -1) {
		s, e := enclosingObject(body, m[0])
		obj := body[s:e]
		if objOwnsID(obj, wantID) {
			hd = html.UnescapeString(unescapeJSON(string(body[m[2]:m[3]])))
			if sm := sdURL.FindSubmatch(obj); sm != nil {
				sd = html.UnescapeString(unescapeJSON(string(sm[1])))
			}
			return hd, sd
		}
	}
	// The reel may only ship an SD rendition; locate it the same way.
	for _, m := range sdURL.FindAllSubmatchIndex(body, -1) {
		s, e := enclosingObject(body, m[0])
		if objOwnsID(body[s:e], wantID) {
			return "", html.UnescapeString(unescapeJSON(string(body[m[2]:m[3]])))
		}
	}
	return "", ""
}

func extract(rawURL string) (*Post, error) {
	if err := validLink(rawURL); err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("GET", strings.TrimSpace(rawURL), nil)
	browserHeaders(req)
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach Facebook: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Facebook returned %d, the post may be private or removed", res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 24<<20))
	if err != nil {
		return nil, err
	}

	final := res.Request.URL.String()
	id := "facebook"
	if m := videoID.FindStringSubmatch(final); m != nil {
		id = m[1]
	}

	// Bind extraction to the requested reel. Only if that fails, e.g. the page
	// shape is one this does not recognise, fall back to document order.
	hd, sd := mediaForID(body, id)
	if hd == "" && sd == "" {
		hd = firstMatch(hdURL, body)
		sd = firstMatch(sdURL, body)
	}
	if hd == "" && sd == "" {
		return nil, errors.New("no video found, the post may be private, a photo, or login walled")
	}

	title := firstMatch(ogTitle, body)
	base := safeName(id)

	var variants []Variant
	if hd != "" {
		variants = append(variants, Variant{URL: hd, Label: "HD", Filename: base + "_hd.mp4"})
	}
	if sd != "" {
		variants = append(variants, Variant{URL: sd, Label: "SD", Filename: base + "_sd.mp4"})
	}

	best := variants[0]
	return &Post{
		ID:      id,
		Title:   title,
		PageURL: final,
		Media: []Media{{
			Type:     "video",
			URL:      best.URL,
			Thumb:    firstMatch(ogImage, body),
			Ext:      "mp4",
			Filename: base + ".mp4",
			Variants: variants,
		}},
	}, nil
}

// handlers
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{
		"ok":      true,
		"service": "facebook",
		"proxied": os.Getenv("FDL_PROXY") != "",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

func handleExtract(w http.ResponseWriter, r *http.Request) {
	post, err := extract(r.URL.Query().Get("url"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, post)
}

// Only Facebook's own media hosts may be fetched, so this cannot be used as an
// open proxy for arbitrary urls.
func allowedMediaHost(h string) bool {
	h = strings.ToLower(h)
	for _, s := range []string{".fbcdn.net", "fbcdn.net", ".facebook.com", "facebook.com", ".fbsbx.com"} {
		if strings.HasSuffix(h, s) {
			return true
		}
	}
	return false
}

func proxyMedia(w http.ResponseWriter, r *http.Request, attachment bool) {
	target := r.URL.Query().Get("url")
	name := r.URL.Query().Get("filename")
	if target == "" {
		writeJSON(w, 400, map[string]string{"error": "url is required"})
		return
	}
	u, err := url.Parse(target)
	if err != nil || !allowedMediaHost(u.Hostname()) {
		writeJSON(w, 403, map[string]string{"error": "only Facebook media hosts are allowed"})
		return
	}

	req, _ := http.NewRequest("GET", target, nil)
	req.Header.Set("User-Agent", browserUA)
	req.Header.Set("Referer", "https://www.facebook.com/")
	req.Header.Set("Accept", "*/*")
	if rng := r.Header.Get("Range"); rng != "" && !attachment {
		req.Header.Set("Range", rng)
	}

	res, err := client.Do(req)
	if err != nil {
		writeJSON(w, 502, map[string]string{"error": "could not fetch media"})
		return
	}
	defer res.Body.Close()

	for _, h := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges"} {
		if v := res.Header.Get(h); v != "" {
			w.Header().Set(h, v)
		}
	}
	if attachment {
		if name == "" {
			name = "facebook.mp4"
		}
		w.Header().Set("Content-Disposition",
			fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(safeName(name))))
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
	} else {
		w.Header().Set("Cache-Control", "private, max-age=600")
		w.WriteHeader(res.StatusCode)
	}
	_, _ = io.Copy(w, res.Body)
}

func handleDownload(w http.ResponseWriter, r *http.Request) { proxyMedia(w, r, true) }
func handleMedia(w http.ResponseWriter, r *http.Request)    { proxyMedia(w, r, false) }

func staticHandler(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if strings.HasPrefix(r.URL.Path, "/api/") {
			log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
		}
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "4447"
	}
	dist := os.Getenv("FDL_DIST")
	if dist == "" {
		dist = "./public"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/extract", handleExtract)
	mux.HandleFunc("/api/download", handleDownload)
	mux.HandleFunc("/api/media", handleMedia)
	mux.Handle("/", staticHandler(dist))

	srv := &http.Server{Addr: ":" + port, Handler: withLogging(mux), ReadHeaderTimeout: 15 * time.Second}
	log.Printf("facebook-downloader listening on :%s (serving %s)", port, dist)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
