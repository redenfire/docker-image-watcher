package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var dockerSocket = "/var/run/docker.sock"

type dockerContainer struct {
	ID      string            `json:"Id"`
	Names   []string          `json:"Names"`
	Image   string            `json:"Image"`
	ImageID string            `json:"ImageID"`
	Labels  map[string]string `json:"Labels"`
}

type dockerInspect struct {
	ID         string          `json:"Id"`
	Name       string          `json:"Name"`
	Config     json.RawMessage `json:"Config"`
	HostConfig json.RawMessage `json:"HostConfig"`
}

func dockerClient() *http.Client {
	return &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", dockerSocket)
			},
		},
	}
}

func pullStreamDockerClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				dialer := &net.Dialer{Timeout: 30 * time.Second}
				return dialer.DialContext(ctx, "unix", dockerSocket)
			},
			ResponseHeaderTimeout: 30 * time.Second,
		},
	}
}

func dockerAPI(method, path string, body io.Reader) (*http.Response, error) {
	client := dockerClient()
	url := "http://localhost" + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}

func listContainers() ([]dockerContainer, error) {
	resp, err := dockerAPI("GET", "/containers/json?all=false&digests=1", nil)
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}
	defer resp.Body.Close()
	var containers []dockerContainer
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return containers, nil
}

func inspectContainer(id string) (*dockerInspect, error) {
	resp, err := dockerAPI("GET", "/containers/"+id+"/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var inspect dockerInspect
	if err := json.NewDecoder(resp.Body).Decode(&inspect); err != nil {
		return nil, err
	}
	return &inspect, nil
}

type PullProgress struct {
	Layer   string `json:"layer"`
	Current int64  `json:"current"`
	Total   int64  `json:"total"`
	Percent int    `json:"percent"`
	Status  string `json:"status"`
}

func IsRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "429") ||
		strings.Contains(msg, "toomanyrequests") ||
		strings.Contains(msg, "pull rate limit") ||
		strings.Contains(msg, "too many requests")
}

func registryAliases(hostname string) []string {
	switch hostname {
	case "registry-1.docker.io", "docker.io", "index.docker.io", "https://index.docker.io/v1/":
		return []string{"https://index.docker.io/v1/", "registry-1.docker.io", "docker.io", "index.docker.io"}
	default:
		return []string{hostname}
	}
}

func buildRegistryAuthHeader(serverAddress, auth string) (string, error) {
	username, password, ok := strings.Cut(auth, ":")
	if !ok || username == "" || password == "" {
		return "", fmt.Errorf("invalid registry auth, expected username:password")
	}
	payload, err := json.Marshal(struct {
		Username      string `json:"username"`
		Password      string `json:"password"`
		ServerAddress string `json:"serveraddress,omitempty"`
	}{
		Username:      username,
		Password:      password,
		ServerAddress: serverAddress,
	})
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(payload), nil
}

func pullImageStream(image string, progressFn func(PullProgress)) error {
	v := url.Values{}
	v.Set("fromImage", image)
	client := pullStreamDockerClient()
	req, err := http.NewRequest("POST", "http://localhost/images/create?"+v.Encode(), nil)
	if err != nil {
		return fmt.Errorf("pull request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	auth := strings.TrimSpace(os.Getenv("DOCKER_REGISTRY_AUTH"))
	if auth != "" {
		registry := ""
		if host, _, _, ok := parseImage(image); ok {
			registry = host
		}
		aliases := registryAliases(registry)
		header := ""
		var authByRegistry map[string]string
		if strings.HasPrefix(auth, "{") {
			if err := json.Unmarshal([]byte(auth), &authByRegistry); err != nil {
				return fmt.Errorf("registry auth: invalid JSON auth map: %w", err)
			}
			for _, alias := range aliases {
				if creds, ok := authByRegistry[alias]; ok && strings.TrimSpace(creds) != "" {
					header, err = buildRegistryAuthHeader(alias, creds)
					if err != nil {
						return fmt.Errorf("registry auth: %w", err)
					}
					break
				}
			}
		} else {
			for _, alias := range aliases {
				if alias == "https://index.docker.io/v1/" {
					header, err = buildRegistryAuthHeader(alias, auth)
					if err != nil {
						return fmt.Errorf("registry auth: %w", err)
					}
					break
				}
			}
		}
		if header != "" {
			req.Header.Set("X-Registry-Auth", header)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("pull request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read error body: %w", err)
		}
		return fmt.Errorf("pull failed: %s %s", resp.Status, strings.TrimSpace(string(body)))
	}

	dec := json.NewDecoder(resp.Body)
	type layerProg struct {
		current, total int64
	}
	layers := make(map[string]*layerProg)
	for {
		var evt struct {
			Status         string `json:"status"`
			ID             string `json:"id"`
			ProgressDetail *struct {
				Current int64 `json:"current"`
				Total   int64 `json:"total"`
			} `json:"progressDetail"`
			Error string `json:"error"`
		}
		if err := dec.Decode(&evt); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("pull stream decode: %w", err)
		}
		if evt.Error != "" {
			return fmt.Errorf("pull error: %s", evt.Error)
		}
		if evt.ID == "" || evt.ProgressDetail == nil {
			progressFn(PullProgress{Status: evt.Status})
			continue
		}
		cur, tot := evt.ProgressDetail.Current, evt.ProgressDetail.Total
		if tot > 0 {
			layers[evt.ID] = &layerProg{current: cur, total: tot}
		} else if lp, ok := layers[evt.ID]; ok {
			lp.current = cur
		}
		var sumCur, sumTot int64
		for _, lp := range layers {
			sumCur += lp.current
			sumTot += lp.total
		}
		pct := 0
		if sumTot > 0 {
			pct = int(sumCur * 100 / sumTot)
		}
		progressFn(PullProgress{
			Layer:   evt.ID,
			Current: sumCur,
			Total:   sumTot,
			Percent: pct,
			Status:  evt.Status,
		})
	}
}

func getLocalDigest(image string) (string, error) {
	resp, err := dockerAPI("GET", "/images/"+image+"/json", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var data struct {
		RepoDigests []string `json:"RepoDigests"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	for _, d := range data.RepoDigests {
		_, after, ok := strings.Cut(d, "@")
		if ok {
			return after, nil
		}
	}
	return "", fmt.Errorf("no valid digest found")
}

func getImageDigest(imageID string) (string, error) {
	resp, err := dockerAPI("GET", "/images/"+imageID+"/json", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var data struct {
		RepoDigests []string `json:"RepoDigests"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	for _, d := range data.RepoDigests {
		_, after, ok := strings.Cut(d, "@")
		if ok {
			return after, nil
		}
	}
	return "", fmt.Errorf("no digest found")
}

func localDigestMatches(imageID, imageTag, remoteDigest string) bool {
	for _, ref := range []string{imageID, imageTag} {
		if ref == "" {
			continue
		}
		resp, err := dockerAPI("GET", "/images/"+ref+"/json", nil)
		if err != nil {
			continue
		}
		var data struct {
			RepoDigests []string `json:"RepoDigests"`
		}
		json.NewDecoder(resp.Body).Decode(&data)
		resp.Body.Close()
		for _, d := range data.RepoDigests {
			_, after, ok := strings.Cut(d, "@")
			if ok && after == remoteDigest {
				return true
			}
		}
	}
	return false
}

func resolveImageName(imageID string) string {
	resp, err := dockerAPI("GET", "/images/"+imageID+"/json", nil)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var data struct {
		RepoTags []string `json:"RepoTags"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return ""
	}
	for _, tag := range data.RepoTags {
		if tag != "<none>:<none>" {
			return tag
		}
	}
	return ""
}

func recreateContainer(id, image string) error {
	inspect, err := inspectContainer(id)
	if err != nil {
		return fmt.Errorf("inspect: %w", err)
	}

	createBody := make(map[string]interface{})
	if err := json.Unmarshal(inspect.Config, &createBody); err != nil {
		return fmt.Errorf("unmarshal container config: %w", err)
	}
	createBody["Image"] = image
	delete(createBody, "Hostname")

	type hostCfg struct {
		Binds         []string               `json:"Binds"`
		PortBindings  map[string]interface{} `json:"PortBindings"`
		RestartPolicy map[string]interface{} `json:"RestartPolicy"`
		NetworkMode   string                 `json:"NetworkMode"`
		Privileged    bool                   `json:"Privileged"`
		ExtraHosts    []string               `json:"ExtraHosts"`
		DNS           []string               `json:"Dns"`
		CapAdd        []string               `json:"CapAdd"`
		CapDrop       []string               `json:"CapDrop"`
		Devices       []interface{}          `json:"Devices"`
		ShmSize       int64                  `json:"ShmSize"`
		Sysctls       map[string]string      `json:"Sysctls"`
		Runtime       string                 `json:"Runtime"`
		GroupAdd      []string               `json:"GroupAdd"`
		IpcMode       string                 `json:"IpcMode"`
		PidMode       string                 `json:"PidMode"`
		UsernsMode    string                 `json:"UsernsMode"`
		UTSMode       string                 `json:"UTSMode"`
	}

	var hc hostCfg
	if err := json.Unmarshal(inspect.HostConfig, &hc); err != nil {
		return fmt.Errorf("unmarshal host config: %w", err)
	}
	createBody["HostConfig"] = hc

	if hc.NetworkMode != "" && hc.NetworkMode != "default" && hc.NetworkMode != "bridge" {
		netConfig := map[string]interface{}{
			"EndpointsConfig": map[string]interface{}{
				hc.NetworkMode: map[string]interface{}{},
			},
		}
		createBody["NetworkingConfig"] = netConfig
	}

	body, err := json.Marshal(createBody)
	if err != nil {
		return fmt.Errorf("marshal create body: %w", err)
	}

	// stop with 10s grace period (best effort — container may already be stopped)
	resp, err := dockerAPI("POST", "/containers/"+id+"/stop?t=10", nil)
	if resp != nil {
		resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("stop container %s: %w", id, err)
	}
	time.Sleep(1 * time.Second)
	// inspect to confirm stopped, then remove
	stopped, err := inspectContainer(id)
	if err != nil {
		return fmt.Errorf("inspect after stop %s: %w", id, err)
	}
	if stopped != nil {
		resp, err = dockerAPI("DELETE", "/containers/"+id, nil)
		if resp != nil {
			resp.Body.Close()
		}
		if err != nil {
			return fmt.Errorf("remove container %s: %w", id, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
	// create
	resp, err = dockerAPI("POST", "/containers/create?name="+strings.TrimPrefix(inspect.Name, "/"), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read create error body: %w", err)
		}
		return fmt.Errorf("create failed: %s %s", resp.Status, string(b))
	}
	var created struct {
		ID string `json:"Id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return fmt.Errorf("decode create response: %w", err)
	}
	// start
	resp, err = dockerAPI("POST", "/containers/"+created.ID+"/start", nil)
	if resp != nil {
		resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("start container %s: %w", created.ID, err)
	}
	return nil
}

func init() {
	if s := os.Getenv("DOCKER_SOCK"); s != "" {
		dockerSocket = s
	}
}
