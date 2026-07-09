package main

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

//go:embed web
var webFS embed.FS

var (
	authUser  string
	authPass  string
	hmacKey   []byte
	loginMu   sync.Mutex
	failCount = make(map[string]*failEntry)
)

type failEntry struct {
	count        int
	last         time.Time
	blockedUntil time.Time
}

func sessionToken(user string) (string, error) {
	expiry := time.Now().Add(24 * time.Hour).Unix()
	payload := user + "@" + strconv.FormatInt(expiry, 10)
	mac := hmac.New(sha256.New, hmacKey)
	mac.Write([]byte(payload))
	sig := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func verifySession(token string) bool {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, hmacKey)
	mac.Write(payload)
	expected := mac.Sum(nil)
	if len(sig) != len(expected) || subtle.ConstantTimeCompare(sig, expected) != 1 {
		return false
	}
	payloadStr := string(payload)
	atIdx := strings.LastIndex(payloadStr, "@")
	if atIdx < 0 {
		return false
	}
	expiry, err := strconv.ParseInt(payloadStr[atIdx+1:], 10, 64)
	if err != nil || time.Now().Unix() > expiry {
		return false
	}
	return true
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Cache-Control", "no-store")
		p := r.URL.Path
		if p == "/login.html" || p == "/api/login" || p == "/api/logout" || p == "/api/auth/status" || p == "/health" {
			next.ServeHTTP(w, r)
			return
		}
		c, err := r.Cookie("session")
		if err != nil || !verifySession(c.Value) {
			if strings.HasPrefix(p, "/api/") {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login.html", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	ip := r.RemoteAddr
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}
	loginMu.Lock()
	fe := failCount[ip]
	now := time.Now()
	if fe != nil {
		if now.Sub(fe.last) > time.Minute {
			fe.count = 0
			fe.blockedUntil = time.Time{}
		}
		if now.Before(fe.blockedUntil) {
			loginMu.Unlock()
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"error": "too many attempts, try later"})
			return
		}
	}
	loginMu.Unlock()

	var creds struct {
		User string `json:"user"`
		Pass string `json:"pass"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil || creds.User == "" || creds.Pass == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "bad request"})
		return
	}

	userOk := subtle.ConstantTimeCompare([]byte(creds.User), []byte(authUser)) == 1
	passOk := subtle.ConstantTimeCompare([]byte(creds.Pass), []byte(authPass)) == 1

	if !userOk || !passOk {
		loginMu.Lock()
		if fe == nil || now.Sub(fe.last) > time.Minute {
			failCount[ip] = &failEntry{count: 1, last: now}
		} else {
			fe.count++
			fe.last = now
			if fe.count >= 5 {
				fe.blockedUntil = now.Add(30 * time.Second)
			}
		}
		loginMu.Unlock()
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid credentials"})
		return
	}

	token, err := sessionToken(creds.User)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
		MaxAge:   86400,
	})
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
		MaxAge:   -1,
	})
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type ContainerItem struct {
	ContainerID   string `json:"container_id"`
	ContainerName string `json:"container_name"`
	LocalDigest   string `json:"local_digest"`
	Status        string `json:"status"`
	AutoUpdate    bool   `json:"auto_update"`
	Error         string `json:"error,omitempty"`
}

type ImageGroup struct {
	Image        string          `json:"image"`
	RemoteDigest string          `json:"remote_digest"`
	Status       string          `json:"status"`
	Containers   []ContainerItem `json:"containers"`
}

type App struct {
	mu              sync.RWMutex
	images          []ImageGroup
	cooldowns       map[string]time.Time
	containerErrors map[string]string
	rateLimited     bool
	progress        sync.Map
	updating        sync.Map
	selfCID         string
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app := &App{
		cooldowns:       make(map[string]time.Time),
		containerErrors: make(map[string]string),
	}
	if host, err := os.Hostname(); err == nil && len(host) >= 12 {
		app.selfCID = host[:12]
		log.Printf("self CID: %s", app.selfCID)
	}

	authUser = os.Getenv("AUTH_USER")
	authPass = os.Getenv("AUTH_PASS")
	if authUser != "" && authPass != "" {
		hmacKey = make([]byte, 32)
		if _, err := rand.Read(hmacKey); err != nil {
			log.Fatalf("generate HMAC key: %v", err)
		}
		log.Println("auth enabled")
	}

	go func() {
		app.checkAll()
		checkInterval := getEnvDuration("CHECK_INTERVAL", 10*time.Minute)
		ticker := time.NewTicker(checkInterval)
		for range ticker.C {
			app.checkAll()
		}
	}()

	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("embedded web directory: %v", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("/login.html", func(w http.ResponseWriter, r *http.Request) {
		if authUser == "" || authPass == "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		http.ServeFileFS(w, r, sub, "login.html")
	})
	mux.HandleFunc("/api/auth/status", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"enabled": authUser != "" && authPass != ""})
	})
	mux.HandleFunc("/api/images", app.handleImages)
	mux.HandleFunc("/api/images/", app.handleImageAction)
	mux.HandleFunc("/api/groups/", app.handleGroupAction)
	mux.HandleFunc("/api/prune", app.handlePrune)
	mux.HandleFunc("/api/ratelimit", app.handleRateLimit)
	mux.HandleFunc("/api/login", handleLogin)
	mux.HandleFunc("/api/logout", handleLogout)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	var handler http.Handler = mux
	if authUser != "" && authPass != "" {
		handler = authMiddleware(mux)
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	go func() {
		log.Printf("Listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func (app *App) handleImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 405)
		return
	}
	app.mu.RLock()
	defer app.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app.images)
}

func (app *App) handleRateLimit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 405)
		return
	}
	app.mu.RLock()
	rateLimited := app.rateLimited
	app.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"rate_limited": rateLimited})
}

func (app *App) handlePrune(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	go func() {
		pruneUnusedImages()
		app.checkAll()
	}()
	w.Write([]byte(`{"status":"pruning"}`))
}

func (app *App) findContainer(cid string) *ContainerItem {
	for i := range app.images {
		for j := range app.images[i].Containers {
			if app.images[i].Containers[j].ContainerID == cid {
				return &app.images[i].Containers[j]
			}
		}
	}
	return nil
}

func (app *App) findImageByCID(cid string) (imageName, containerName string) {
	for i := range app.images {
		for j := range app.images[i].Containers {
			if app.images[i].Containers[j].ContainerID == cid {
				return app.images[i].Image, app.images[i].Containers[j].ContainerName
			}
		}
	}
	return "", ""
}

func (app *App) handleImageAction(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/images/"), "/")
	if len(parts) != 2 {
		http.Error(w, "bad path: /api/images/{id}/{action}", 400)
		return
	}
	cid, action := parts[0], parts[1]

	if action == "progress" {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", 405)
			return
		}
		v, ok := app.progress.Load(cid)
		if !ok {
			http.Error(w, "no progress", 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(v)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}

	app.mu.RLock()
	img := app.findContainer(cid)
	app.mu.RUnlock()
	if img == nil {
		http.Error(w, "container not found", 404)
		return
	}

	switch action {
	case "update":
		go app.updateContainer(cid)
		w.Write([]byte(`{"status":"updating"}`))
	case "auto-update":
		app.mu.RLock()
		c := app.findContainer(cid)
		on := c != nil && c.AutoUpdate
		app.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"auto_update": on})
	default:
		http.Error(w, "unknown action", 400)
	}
}

func (app *App) handleGroupAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/groups/")
	if path == "" {
		http.Error(w, "not found", 404)
		return
	}
	if path != "update" {
		http.Error(w, "unknown action", 400)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Image string `json:"image"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Image == "" {
		http.Error(w, "bad request", 400)
		return
	}
	go app.updateGroup(req.Image)
	w.Write([]byte(`{"status":"updating"}`))
}

func (app *App) updateGroup(image string) {
	seen := map[string]bool{}
	for {
		cid := ""
		app.mu.RLock()
		for _, g := range app.images {
			if g.Image == image {
				for _, c := range g.Containers {
					if c.Status == "outdated" && !seen[c.ContainerID] {
						cid = c.ContainerID
						seen[cid] = true
						goto found
					}
				}
				break
			}
		}
		app.mu.RUnlock()
		break
	found:
		app.mu.RUnlock()
		log.Printf("updating %s in group %s", cid, image)
		app.updateContainer(cid)
	}
}

func (app *App) checkAll() {
	containers, err := listContainers()
	if err != nil {
		log.Printf("list containers: %v", err)
		return
	}

	type groupEntry struct {
		cid      string
		name     string
		imgID    string
		imageTag string
		labels   map[string]string
	}
	groups := make(map[string][]groupEntry)
	for _, c := range containers {
		if strings.HasPrefix(c.Image, "sha256:") {
			continue
		}
		if app.selfCID != "" && c.ID[:12] == app.selfCID {
			continue
		}
		imageRef := c.Image
		if !strings.Contains(imageRef, "/") && !strings.Contains(imageRef, ":") {
			if r := resolveImageName(c.ImageID); r != "" {
				imageRef = r
			}
		}
		cleanImage := imageRef
		if i := strings.Index(cleanImage, "@"); i >= 0 {
			cleanImage = cleanImage[:i]
		}
		if _, _, ok := strings.Cut(cleanImage, ":"); !ok {
			cleanImage = cleanImage + ":latest"
		}
		entryName := c.ID[:12]
		if len(c.Names) > 0 {
			entryName = strings.TrimPrefix(c.Names[0], "/")
		}
		groups[cleanImage] = append(groups[cleanImage], groupEntry{
			cid:      c.ID[:12],
			name:     entryName,
			imgID:    c.ImageID,
			imageTag: c.Image,
			labels:   c.Labels,
		})
	}

	maxWorkers := 5
	if s := os.Getenv("CHECK_CONCURRENCY"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			maxWorkers = n
		}
	}

	sem := make(chan struct{}, maxWorkers)
	resultsCh := make(chan ImageGroup, len(groups))
	sawRateLimit := false
	var rateLimitMu sync.Mutex
	var wg sync.WaitGroup

	for imageName, entries := range groups {
		imageName, entries := imageName, entries
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			remoteDigest, remoteErr := getRemoteDigest(imageName)
			remoteStr := "unknown"
			if remoteErr == nil {
				remoteStr = shortenDigest(remoteDigest)
			} else if IsRateLimitError(remoteErr) {
				rateLimitMu.Lock()
				sawRateLimit = true
				rateLimitMu.Unlock()
			}

			var containers []ContainerItem
			gUpToDate := 0
			gOutdated := 0
			for _, e := range entries {
				item := ContainerItem{
					ContainerID:   e.cid,
					ContainerName: e.name,
				}
				localDigest, localErr := getImageDigest(e.imgID)
				if localErr != nil {
					localDigest, localErr = getLocalDigest(e.imageTag)
				}
				if localErr != nil {
					item.LocalDigest = "unknown"
					item.Status = "unknown"
				} else if remoteErr != nil {
					item.LocalDigest = shortenDigest(localDigest)
					item.Status = "unknown"
				} else if !localDigestMatches(e.imgID, remoteDigest) {
					item.LocalDigest = shortenDigest(localDigest)
					item.Status = "outdated"
					gOutdated++
				} else {
					item.LocalDigest = shortenDigest(localDigest)
					item.Status = "uptodate"
					gUpToDate++
				}

				if e.labels["image-watch.auto-update"] == "true" {
					item.AutoUpdate = true
				}
				app.mu.RLock()
				item.Error = app.containerErrors[e.cid]
				app.mu.RUnlock()
				containers = append(containers, item)
			}

			gStatus := "uptodate"
			if gOutdated > 0 {
				gStatus = "outdated"
			}
			if gOutdated > 0 && gUpToDate > 0 {
				gStatus = "partial"
			}
			if gOutdated+gUpToDate == 0 {
				gStatus = "unknown"
			}

			resultsCh <- ImageGroup{
				Image:        imageName,
				RemoteDigest: remoteStr,
				Status:       gStatus,
				Containers:   containers,
			}
		}()
	}

	wg.Wait()
	close(resultsCh)

	var results []ImageGroup
	for result := range resultsCh {
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Image < results[j].Image
	})

	rateLimited := sawRateLimit
	app.mu.RLock()
	for _, err := range app.containerErrors {
		if IsRateLimitError(errors.New(err)) {
			rateLimited = true
			break
		}
	}
	app.mu.RUnlock()

	app.mu.Lock()
	app.images = results
	app.rateLimited = rateLimited
	app.mu.Unlock()

	autoCooldown := getEnvDuration("AUTO_COOLDOWN", 5*time.Minute)
	for _, g := range results {
		for _, c := range g.Containers {
			if c.AutoUpdate && c.Status == "outdated" {
				app.mu.Lock()
				last, ok := app.cooldowns[c.ContainerID]
				cooldown := !ok || time.Since(last) > autoCooldown
				if cooldown {
					app.cooldowns[c.ContainerID] = time.Now()
				}
				app.mu.Unlock()
				if cooldown {
					log.Printf("auto-updating %s (%s)", c.ContainerName, g.Image)
					app.updateContainer(c.ContainerID)
				}
			}
		}
	}
}

func (app *App) updateContainer(cid string) {
	if cid == app.selfCID {
		log.Printf("update %s: skipping self-update", cid)
		return
	}
	if _, loaded := app.updating.LoadOrStore(cid, true); loaded {
		log.Printf("update %s: already in progress, skipping", cid)
		return
	}
	defer app.updating.Delete(cid)

	app.mu.RLock()
	image, containerName := app.findImageByCID(cid)
	app.mu.RUnlock()

	if image == "" {
		return
	}

	app.progress.Store(cid, PullProgress{Status: "connecting", Layer: "..."})
	if err := pullImageStream(image, func(p PullProgress) {
		app.progress.Store(cid, p)
	}); err != nil {
		app.mu.Lock()
		app.containerErrors[cid] = err.Error()
		if IsRateLimitError(err) {
			app.rateLimited = true
		}
		app.mu.Unlock()
		app.progress.Store(cid, PullProgress{Status: "error: " + err.Error()})
		log.Printf("pull %s: %v", image, err)
		app.checkAll()
		return
	}
	app.mu.Lock()
	delete(app.containerErrors, cid)
	app.mu.Unlock()
	app.progress.Store(cid, PullProgress{Status: "recreating", Percent: 100})
	if err := recreateContainer(cid, image); err != nil {
		app.mu.Lock()
		app.containerErrors[cid] = err.Error()
		app.mu.Unlock()
		app.progress.Store(cid, PullProgress{Status: "error: " + err.Error()})
		log.Printf("recreate %s: %v", cid, err)
		app.checkAll()
		return
	}
	var newDigest string
	if d, err := getLocalDigest(image); err == nil {
		newDigest = shortenDigest(d)
	}
	app.mu.Lock()
	delete(app.containerErrors, cid)
	if c := app.findContainer(cid); c != nil {
		c.Status = "uptodate"
		c.Error = ""
		if newDigest != "" {
			c.LocalDigest = newDigest
		}
	}
	app.mu.Unlock()
	app.progress.Delete(cid)
	log.Printf("updated %s -> %s", containerName, image)
	app.checkAll()
}

func shortenDigest(d string) string {
	if len(d) > 17 {
		return d[:17]
	}
	return d
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	s := os.Getenv(key)
	if s == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Printf("invalid %s=%q, using default %v", key, s, defaultVal)
		return defaultVal
	}
	return d
}
