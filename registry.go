package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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

	digest, body, code, headers, err := fetchManifest(baseURL, "", registry)
	if err != nil {
		return "", err
	}
	if code == 200 {
		return resolveManifest(digest, body, headers, registry, repo, "")
	}
	if code != 401 {
		return "", fmt.Errorf("registry returned %d", code)
	}

	authHeader := headers.Get("Www-Authenticate")
	token, err := getToken(authHeader, registry, repo)
	if err != nil {
		return "", fmt.Errorf("auth: %w", err)
	}

	digest, body, code, headers, err = fetchManifest(baseURL, token, registry)
	if err != nil {
		return "", err
	}
	if code != 200 {
		return "", fmt.Errorf("registry returned %d after auth", code)
	}
	return resolveManifest(digest, body, headers, registry, repo, token)
}

func resolveManifest(digest string, body []byte, headers http.Header, registry, repo, token string) (string, error) {
	return digest, nil
}

func fetchManifest(manifestURL, token, registry string) (digest string, body []byte, code int, headers http.Header, err error) {
	req, err := http.NewRequest("GET", manifestURL, nil)
	if err != nil {
		return "", nil, 0, nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else if err := setRegistryBasicAuth(req, registry); err != nil {
		return "", nil, 0, nil, err
	}
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json,application/vnd.oci.image.manifest.v1+json,application/vnd.oci.image.index.v1+json,application/vnd.docker.distribution.manifest.list.v2+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", nil, 0, nil, err
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, 0, nil, err
	}
	return resp.Header.Get("Docker-Content-Digest"), body, resp.StatusCode, resp.Header, nil
}

func getRegistryBasicAuth(registry string) (username, password string, ok bool, err error) {
	auth := strings.TrimSpace(os.Getenv("DOCKER_REGISTRY_AUTH"))
	if auth == "" {
		return "", "", false, nil
	}

	if strings.HasPrefix(auth, "{") {
		var authByRegistry map[string]string
		if err := json.Unmarshal([]byte(auth), &authByRegistry); err != nil {
			return "", "", false, fmt.Errorf("invalid DOCKER_REGISTRY_AUTH JSON: %w", err)
		}
		for _, alias := range registryAliases(registry) {
			creds := authByRegistry[alias]
			if creds == "" {
				continue
			}
			username, password, ok := strings.Cut(creds, ":")
			if !ok || username == "" || password == "" {
				return "", "", false, fmt.Errorf("invalid DOCKER_REGISTRY_AUTH credentials for %s", alias)
			}
			return username, password, true, nil
		}
		return "", "", false, nil
	}

	isDockerHub := false
	for _, alias := range registryAliases(registry) {
		if alias == "https://index.docker.io/v1/" {
			isDockerHub = true
			break
		}
	}
	if !isDockerHub {
		return "", "", false, nil
	}

	username, password, ok = strings.Cut(auth, ":")
	if !ok || username == "" || password == "" {
		return "", "", false, fmt.Errorf("invalid DOCKER_REGISTRY_AUTH shorthand credentials")
	}
	return username, password, true, nil
}

func setRegistryBasicAuth(req *http.Request, registry string) error {
	username, password, ok, err := getRegistryBasicAuth(registry)
	if err != nil || !ok {
		return err
	}
	req.SetBasicAuth(username, password)
	return nil
}

func buildTokenURL(realm, service, scope string) (string, error) {
	u, err := url.Parse(realm)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("scope", scope)
	if service != "" {
		q.Set("service", service)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func getToken(authHeader, registry, repo string) (string, error) {
	matches := authRe.FindStringSubmatch(authHeader)
	if len(matches) >= 2 && matches[1] != "" {
		realm := matches[1]
		service := matches[2]
		scope := matches[3]
		if scope == "" {
			scope = "repository:" + repo + ":pull"
		}
		tokenURL, err := buildTokenURL(realm, service, scope)
		if err != nil {
			return "", err
		}
		req, err := http.NewRequest("GET", tokenURL, nil)
		if err != nil {
			return "", err
		}
		if err := setRegistryBasicAuth(req, registry); err != nil {
			return "", err
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return "", fmt.Errorf("token endpoint returned %d", resp.StatusCode)
		}
		var data struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return "", err
		}
		if data.Token == "" {
			return "", fmt.Errorf("token endpoint returned empty token")
		}
		return data.Token, nil
	}

	tokenURL, err := buildTokenURL("https://auth.docker.io/token", "registry.docker.io", "repository:"+repo+":pull")
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("GET", tokenURL, nil)
	if err != nil {
		return "", err
	}
	if err := setRegistryBasicAuth(req, registry); err != nil {
		return "", err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("token endpoint returned %d", resp.StatusCode)
	}
	var data struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if data.Token == "" {
		return "", fmt.Errorf("token endpoint returned empty token")
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
