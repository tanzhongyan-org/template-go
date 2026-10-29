package main

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/tanzhongyan-org/template-go/internal/config"
	"github.com/tanzhongyan-org/template-go/internal/db"
	"gopkg.in/yaml.v3"
)

// openapiSpec is the subset of the spec the coverage check reads.
type openapiSpec struct {
	Paths map[string]map[string]any `yaml:"paths"`
}

// normalizeRoute drops a trailing slash so chi mount patterns and spec paths
// compare in one shape.
func normalizeRoute(p string) string {
	if len(p) > 1 {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}

func specRoutes(t *testing.T) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "api", "openapi.yaml"))
	if err != nil {
		t.Fatalf("read OpenAPI spec: %v", err)
	}
	var spec openapiSpec
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		t.Fatalf("parse OpenAPI spec: %v", err)
	}
	routes := make(map[string]bool, len(spec.Paths)*2)
	for path, ops := range spec.Paths {
		for method := range ops {
			routes[strings.ToUpper(method)+" "+normalizeRoute(path)] = true
		}
	}
	return routes
}

func routerRoutes(t *testing.T) map[string]bool {
	t.Helper()
	cfg := &config.Config{AppName: "template-go", AppEnv: "test", Port: 8080}
	router := SetupRouter(cfg, db.New(nil))
	routes := make(map[string]bool)
	if err := chi.Walk(router, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		routes[method+" "+normalizeRoute(route)] = true
		return nil
	}); err != nil {
		t.Fatalf("walk router: %v", err)
	}
	return routes
}

func missingKeys(want, have map[string]bool) []string {
	var missing []string
	for k := range want {
		if !have[k] {
			missing = append(missing, k)
		}
	}
	sort.Strings(missing)
	return missing
}

// TestRoutesMatchOpenAPISpec keeps the committed spec and the router in lock
// step: a route missing from api/openapi.yaml fails, and a spec path with no
// route fails. Update the spec, then run `make openapi-client`.
func TestRoutesMatchOpenAPISpec(t *testing.T) {
	fromSpec := specRoutes(t)
	fromRouter := routerRoutes(t)
	if missing := missingKeys(fromRouter, fromSpec); len(missing) > 0 {
		t.Errorf("routes missing from api/openapi.yaml: %v", missing)
	}
	if missing := missingKeys(fromSpec, fromRouter); len(missing) > 0 {
		t.Errorf("spec paths with no route: %v", missing)
	}
}
