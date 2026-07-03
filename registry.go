package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

var authRe = regexp.MustCompile(`Bearer realm="([^"]+)"(?:,service="([^"]*)")?(?:,scope="([^"]*)")?`)

func getRemoteDigest(image string) (string, error) {
	registry, repo, tag, ok := parseImage(image)
	if !ok {
		return "", fmt.Errorf("cannot parse image")
	}

	baseURL := fmt.Sprintf("https://%s/v2/%s/manifests/%s", registry, repo, tag)

	digest, code, headers, err := fetchManifest(baseURL, "")
	if err != nil {
		return "", err
	}
	if code == 200 {
		return digest, nil
	}
	if code != 401 {
		return "", fmt.Errorf("registry returned %d", code)
	}

	authHeader := headers.Get("Www-Authenticate")
	token, err := getToken(authHeader, registry, repo)
	if err != nil {
		return "", fmt.Errorf("auth: %w", err)
	}

	digest, code, _, err = fetchManifest(baseURL, token)
	if err != nil {
		return "", err
	}
	if code != 200 {
		return "", fmt.Errorf("registry returned %d after auth", code)
	}
	return digest, nil
}

func fetchManifest(url, token string) (digest string, code int, headers http.Header, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", 0, nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json,application/vnd.oci.image.manifest.v1+json,application/vnd.oci.image.index.v1+json,application/vnd.docker.distribution.manifest.list.v2+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", 0, nil, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return resp.Header.Get("Docker-Content-Digest"), resp.StatusCode, resp.Header, nil
}

func getToken(authHeader, registry, repo string) (string, error) {
	// parse WWW-Authenticate header
	matches := authRe.FindStringSubmatch(authHeader)
	if len(matches) >= 2 && matches[1] != "" {
		realm := matches[1]
		service := matches[2]
		scope := matches[3]
		if scope == "" {
			scope = "repository:" + repo + ":pull"
		}
		url := realm + "?scope=" + scope
		if service != "" {
			url += "&service=" + service
		}
		resp, err := httpClient.Get(url)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		var data struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return "", err
		}
		return data.Token, nil
	}

	// fallback: Docker Hub-style token endpoint
	url := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", repo)
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var data struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	return data.Token, nil
}

func parseImage(image string) (registry, repo, tag string, ok bool) {
	if strings.HasPrefix(image, "sha256:") {
		return "", "", "", false
	}
	// strip @sha256 suffix
	if i := strings.Index(image, "@"); i >= 0 {
		image = image[:i]
	}

	tag = "latest"
	if i := strings.LastIndex(image, ":"); i >= 0 {
		// check that the part after ":" doesn't contain "/"
		// (avoid matching port in registry like localhost:5000)
		if !strings.Contains(image[i+1:], "/") {
			tag = image[i+1:]
			image = image[:i]
		}
	}

	// split registry from repo path
	if strings.Contains(image, "/") {
		parts := strings.SplitN(image, "/", 2)
		if strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":") {
			registry = parts[0]
			repo = parts[1]
		} else {
			registry = "registry-1.docker.io"
			repo = image
		}
	} else {
		registry = "registry-1.docker.io"
		repo = "library/" + image
	}

	// strip docker.io alias
	if registry == "docker.io" {
		registry = "registry-1.docker.io"
	}

	return registry, repo, tag, true
}
