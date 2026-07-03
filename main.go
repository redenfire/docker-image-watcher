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
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

//go:embed web
var webFS embed.FS

var (
	authUser   string
	authPass   string
	hmacKey    []byte
	loginMu    sync.Mutex
	failCount  = make(map[string]*failEntry)
)

type failEntry struct {
	count       int
	last        time.Time
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
	parts2 := strings.SplitN(string(payload), "@", 2)
	if len(parts2) != 2 {
		return false
	}
	expiry, err := strconv.ParseInt(parts2[1], 10, 64)
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
		if p == "/login.html" || p == "/api/login" || p == "/api/logout" || p == "/health" {
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
}

type ImageGroup struct {
	Image        string          `json:"image"`
	RemoteDigest string          `json:"remote_digest"`
	Status       string          `json:"status"`
	Containers   []ContainerItem `json:"containers"`
}

type App struct {
	mu        sync.RWMutex
	images    []ImageGroup
	autoFile  string
	autoSaved map[string]bool
	cooldowns map[string]time.Time
	progress  sync.Map
	updating  sync.Map
	selfCID   string
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	autoFile := os.Getenv("AUTO_FILE")
	if autoFile == "" {
		autoFile = "/data/auto-update.json"
	}

	app := &App{
		autoFile:  autoFile,
		autoSaved: make(map[string]bool),
		cooldowns: make(map[string]time.Time),
	}
	app.loadAuto()
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
		ticker := time.NewTicker(10 * time.Minute)
		for range ticker.C {
			app.checkAll()
		}
	}()

	sub, _ := fs.Sub(webFS, "web")
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("/api/images", app.handleImages)
	mux.HandleFunc("/api/images/", app.handleImageAction)
	mux.HandleFunc("/api/groups/", app.handleGroupAction)
	mux.HandleFunc("/api/login", handleLogin)
	mux.HandleFunc("/api/logout", handleLogout)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	var handler http.Handler = mux
	if authUser != "" && authPass != "" {
		handler = authMiddleware(mux)
	}
	srv := &http.Server{Addr: ":" + port, Handler: handler}
	go func() {
		log.Printf("Listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
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
		app.toggleAuto(cid)
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
	app.mu.RLock()
	var cids []string
	for _, g := range app.images {
		if g.Image == image {
			for _, c := range g.Containers {
				cids = append(cids, c.ContainerID)
			}
			break
		}
	}
	app.mu.RUnlock()

	log.Printf("updating group %s (%d containers)", image, len(cids))
	for _, cid := range cids {
		app.updateContainer(cid)
	}
}

func (app *App) checkAll() {
	containers, err := listContainers()
	if err != nil {
		log.Printf("list containers: %v", err)
		return
	}

	app.mu.Lock()
	autoMap := make(map[string]bool)
	for k, v := range app.autoSaved {
		autoMap[k] = v
	}
	app.mu.Unlock()

	type groupEntry struct {
		cid      string
		name     string
		imgID    string
		imageTag string
	}
	groups := make(map[string][]groupEntry)
	for _, c := range containers {
		if strings.HasPrefix(c.Image, "sha256:") {
			continue
		}
		if app.selfCID != "" && c.ID[:12] == app.selfCID {
			continue
		}
		cleanImage := c.Image
		if i := strings.Index(cleanImage, "@"); i >= 0 {
			cleanImage = cleanImage[:i]
		}
		if _, _, ok := strings.Cut(cleanImage, ":"); !ok {
			cleanImage = cleanImage + ":latest"
		}
		groups[cleanImage] = append(groups[cleanImage], groupEntry{
			cid:      c.ID[:12],
			name:     strings.TrimPrefix(c.Names[0], "/"),
			imgID:    c.ImageID,
			imageTag: c.Image,
		})
	}

	var results []ImageGroup
	for imageName, entries := range groups {
		remoteDigest, remoteErr := getRemoteDigest(imageName)
		remoteStr := "unknown"
		if remoteErr == nil {
			remoteStr = shortenDigest(remoteDigest)
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
			} else if remoteErr != nil || localDigest != remoteDigest {
				item.LocalDigest = shortenDigest(localDigest)
				item.Status = "outdated"
				gOutdated++
			} else {
				item.LocalDigest = shortenDigest(localDigest)
				item.Status = "uptodate"
				gUpToDate++
			}

			if auto, ok := autoMap[e.cid]; ok {
				item.AutoUpdate = auto
			}
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

		results = append(results, ImageGroup{
			Image:        imageName,
			RemoteDigest: remoteStr,
			Status:       gStatus,
			Containers:   containers,
		})
	}

	app.mu.Lock()
	app.images = results
	app.mu.Unlock()

	for _, g := range results {
		for _, c := range g.Containers {
			if c.AutoUpdate && c.Status == "outdated" {
				app.mu.Lock()
				last, ok := app.cooldowns[c.ContainerID]
				cooldown := !ok || time.Since(last) > 5*time.Minute
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
		app.progress.Store(cid, PullProgress{Status: "error: " + err.Error()})
		log.Printf("pull %s: %v", image, err)
		return
	}
	app.progress.Store(cid, PullProgress{Status: "recreating", Percent: 100})
	if err := recreateContainer(cid, image); err != nil {
		app.progress.Store(cid, PullProgress{Status: "error: " + err.Error()})
		log.Printf("recreate %s: %v", cid, err)
		return
	}
	app.progress.Delete(cid)
	log.Printf("updated %s -> %s", containerName, image)
	app.checkAll()
}

func (app *App) toggleAuto(cid string) {
	if cid == app.selfCID {
		return
	}
	app.mu.Lock()
	defer app.mu.Unlock()
	for i := range app.images {
		for j := range app.images[i].Containers {
			if app.images[i].Containers[j].ContainerID == cid {
				app.images[i].Containers[j].AutoUpdate = !app.images[i].Containers[j].AutoUpdate
				app.autoSaved[cid] = app.images[i].Containers[j].AutoUpdate
				return
			}
		}
	}
}

func (app *App) loadAuto() {
	data, err := os.ReadFile(app.autoFile)
	if err != nil {
		return
	}
	app.mu.Lock()
	defer app.mu.Unlock()
	json.Unmarshal(data, &app.autoSaved)
}

func saveAuto(auto map[string]bool, p string) {
	data, _ := json.Marshal(auto)
	if i := strings.LastIndex(p, "/"); i > 0 {
		os.MkdirAll(p[:i], 0755)
	}
	os.WriteFile(p, data, 0644)
}

func shortenDigest(d string) string {
	if len(d) > 17 {
		return d[:17]
	}
	return d
}
