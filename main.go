package main

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

//go:embed web
var webFS embed.FS

type ImageStatus struct {
	ContainerID   string `json:"container_id"`
	ContainerName string `json:"container_name"`
	Image         string `json:"image"`
	LocalDigest   string `json:"local_digest"`
	RemoteDigest  string `json:"remote_digest"`
	Status        string `json:"status"`
	AutoUpdate    bool   `json:"auto_update"`
}

type App struct {
	mu         sync.RWMutex
	images     []ImageStatus
	autoFile   string
	autoSaved  map[string]bool
	cooldowns  map[string]time.Time
	progress   sync.Map
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
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{Addr: ":" + port, Handler: mux}
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

func (app *App) handleImageAction(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/images/"), "/")
	if len(parts) != 2 {
		http.Error(w, "bad path: /api/images/{id}/{action}", 400)
		return
	}
	cid, action := parts[0], parts[1]

	// progress is read-only
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
	var img *ImageStatus
	for i := range app.images {
		if app.images[i].ContainerID == cid {
			img = &app.images[i]
			break
		}
	}
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
		var on bool
		for i := range app.images {
			if app.images[i].ContainerID == cid {
				on = app.images[i].AutoUpdate
				break
			}
		}
		app.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"auto_update": on})
	default:
		http.Error(w, "unknown action", 400)
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

	type containerResult struct {
		status ImageStatus
	}

	sem := make(chan struct{}, 5)
	resultsCh := make(chan containerResult, len(containers))
	var wg sync.WaitGroup

	for _, c := range containers {
		if strings.HasPrefix(c.Image, "sha256:") {
			continue
		}
		wg.Add(1)
		go func(c dockerContainer) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			status := ImageStatus{
				ContainerID:   c.ID[:12],
				ContainerName: strings.TrimPrefix(c.Names[0], "/"),
				Image:         c.Image,
			}
			cleanImage := c.Image
			if i := strings.Index(cleanImage, "@"); i >= 0 {
				cleanImage = cleanImage[:i]
			}
			if _, _, ok := strings.Cut(cleanImage, ":"); ok {
				status.Image = cleanImage
			} else {
				status.Image = cleanImage + ":latest"
			}

			localDigest, localErr := getLocalDigest(c.Image)
			if localErr != nil {
				status.LocalDigest = "unknown"
			} else {
				status.LocalDigest = shortenDigest(localDigest)
			}

			remoteDigest, remoteErr := getRemoteDigest(status.Image)
			if remoteErr != nil {
				status.RemoteDigest = "unknown"
			} else {
				status.RemoteDigest = shortenDigest(remoteDigest)
			}

			if localErr != nil || remoteErr != nil {
				status.Status = "unknown"
			} else if localDigest != remoteDigest {
				status.Status = "outdated"
			} else {
				status.Status = "uptodate"
			}

			if auto, ok := autoMap[c.ID[:12]]; ok {
				status.AutoUpdate = auto
			}

			resultsCh <- containerResult{status: status}
		}(c)
	}

	wg.Wait()
	close(resultsCh)

	var results []ImageStatus
	for r := range resultsCh {
		results = append(results, r.status)
	}

	app.mu.Lock()
	app.images = results
	app.mu.Unlock()

	for _, img := range results {
		if img.AutoUpdate && img.Status == "outdated" {
			app.mu.Lock()
			last, ok := app.cooldowns[img.ContainerID]
			cooldown := !ok || time.Since(last) > 5*time.Minute
			if cooldown {
				app.cooldowns[img.ContainerID] = time.Now()
			}
			app.mu.Unlock()
			if cooldown {
				log.Printf("auto-updating %s (%s)", img.ContainerName, img.Image)
				app.updateContainer(img.ContainerID)
			}
		}
	}
}

func (app *App) updateContainer(cid string) {
	app.mu.RLock()
	var img *ImageStatus
	for i := range app.images {
		if app.images[i].ContainerID == cid {
			img = &app.images[i]
			break
		}
	}
	app.mu.RUnlock()
	if img == nil {
		return
	}

	app.progress.Store(cid, PullProgress{Status: "connecting", Layer: "..."})
	if err := pullImageStream(img.Image, func(p PullProgress) {
		app.progress.Store(cid, p)
	}); err != nil {
		app.progress.Store(cid, PullProgress{Status: "error: " + err.Error()})
		log.Printf("pull %s: %v", img.Image, err)
		return
	}
	app.progress.Store(cid, PullProgress{Status: "recreating", Percent: 100})
	if err := recreateContainer(cid, img.Image); err != nil {
		app.progress.Store(cid, PullProgress{Status: "error: " + err.Error()})
		log.Printf("recreate %s: %v", cid, err)
		return
	}
	app.progress.Delete(cid)
	log.Printf("updated %s -> %s", img.ContainerName, img.Image)
	app.checkAll()
}

func (app *App) toggleAuto(cid string) {
	app.mu.Lock()
	defer app.mu.Unlock()
	for i := range app.images {
		if app.images[i].ContainerID == cid {
			app.images[i].AutoUpdate = !app.images[i].AutoUpdate
			app.autoSaved[cid] = app.images[i].AutoUpdate
			break
		}
	}
	saveAuto(app.autoSaved, app.autoFile)
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
