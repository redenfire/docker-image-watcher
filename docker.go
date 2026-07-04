package main

import (
	"bytes"
	"context"
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
	ID      string   `json:"Id"`
	Names   []string `json:"Names"`
	Image   string   `json:"Image"`
	ImageID string   `json:"ImageID"`
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
	resp, err := dockerAPI("GET", "/containers/json?all=false", nil)
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

func pullImageStream(image string, progressFn func(PullProgress)) error {
	v := url.Values{}
	v.Set("fromImage", image)
	resp, err := dockerAPI("POST", "/images/create?"+v.Encode(), nil)
	if err != nil {
		return fmt.Errorf("pull request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
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
		if _, after, ok := strings.Cut(d, "@"); ok {
			return after, nil
		}
		return d, nil
	}
	return "", fmt.Errorf("no digest found")
}

func recreateContainer(id, image string) error {
	inspect, err := inspectContainer(id)
	if err != nil {
		return fmt.Errorf("inspect: %w", err)
	}

	createBody := make(map[string]interface{})
	json.Unmarshal(inspect.Config, &createBody)
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
	json.Unmarshal(inspect.HostConfig, &hc)
	createBody["HostConfig"] = hc

	if hc.NetworkMode != "" && hc.NetworkMode != "default" && hc.NetworkMode != "bridge" {
		netConfig := map[string]interface{}{
			"EndpointsConfig": map[string]interface{}{
				hc.NetworkMode: map[string]interface{}{},
			},
		}
		createBody["NetworkingConfig"] = netConfig
	}

	body, _ := json.Marshal(createBody)

	// stop with 10s grace period (best effort — container may already be stopped)
	resp, _ := dockerAPI("POST", "/containers/"+id+"/stop?t=10", nil)
	if resp != nil {
		resp.Body.Close()
	}
	time.Sleep(1 * time.Second)
	// inspect to confirm stopped, then remove
	stopped, _ := inspectContainer(id)
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
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create failed: %s %s", resp.Status, string(b))
	}
	var created struct {
		ID string `json:"Id"`
	}
	json.NewDecoder(resp.Body).Decode(&created)
	// start
	resp, _ = dockerAPI("POST", "/containers/"+created.ID+"/start", nil)
	if resp != nil {
		resp.Body.Close()
	}
	return nil
}

func init() {
	if s := os.Getenv("DOCKER_SOCK"); s != "" {
		dockerSocket = s
	}
}
