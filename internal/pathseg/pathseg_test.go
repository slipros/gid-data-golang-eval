package pathseg

import (
	"slices"
	"testing"
)

func TestHasLayer(t *testing.T) {
	tests := []struct {
		name string
		path string
		seq  []string
		want bool
	}{
		// internal/ layout: layer is anchored right after /internal/.
		{"internal client layer", "mod/internal/client/billing", []string{"client"}, true},
		{"internal domain/model layer", "mod/internal/domain/model/filter", []string{"domain", "model"}, true},
		{"internal domain prefix", "mod/internal/domain/service", []string{"domain"}, true},
		// The reported false-positive: a nested client segment below another
		// layer is NOT the client layer.
		{"nested client not a layer", "mod/internal/connect/client/interceptor", []string{"client"}, false},
		{"nested dal not a layer", "mod/internal/server/grpc/dalstats", []string{"dal"}, false},
		// pair scope must not match a single-segment prefix.
		{"dal/repository vs dal/entity", "mod/internal/dal/entity", []string{"dal", "repository"}, false},
		{"dal/repository match", "mod/internal/dal/repository/build", []string{"dal", "repository"}, true},
		// segment must match exactly, not by prefix of the segment string.
		{"events is not event", "mod/internal/events/dto", []string{"event"}, false},
		{"metrics is not metric", "mod/internal/metrics/registry", []string{"metric"}, false},
		// pkg/<module> layout: layer is anchored right after pkg/<module>.
		{"pkg module client layer", "mod/pkg/billing/client/snapshot", []string{"client"}, true},
		{"pkg module nested client", "mod/pkg/billing/connect/client/x", []string{"client"}, false},
		{"pkg module root has no layer", "mod/pkg/billing", []string{"client"}, false},
		// non-standard layout (testdata): first segment is the module root.
		{"testdata client layer", "svc/client/billing", []string{"client"}, true},
		{"testdata nested client", "svc/connect/client/interceptor", []string{"client"}, false},
		{"testdata domain/model", "svc/domain/model", []string{"domain", "model"}, true},
		// empty seq never matches.
		{"empty seq", "svc/client", nil, false},
	}
	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		t.Run(tt.name, func(t *testing.T) {
			if got := HasLayer(tt.path, tt.seq...); got != tt.want {
				t.Errorf("HasLayer(%q, %v) = %v, want %v", tt.path, tt.seq, got, tt.want)
			}
		})
	}
}

func TestLayerSegments(t *testing.T) {
	tests := []struct {
		path string
		want []string
	}{
		{"mod/internal/domain/model/filter", []string{"domain", "model", "filter"}},
		{"mod/pkg/billing/dal/repository", []string{"dal", "repository"}},
		{"mod/pkg/billing", nil},
		{"svc/client/billing", []string{"client", "billing"}},
		{"svc", nil},
	}
	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		if got := LayerSegments(tt.path); !slices.Equal(got, tt.want) {
			t.Errorf("LayerSegments(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestModuleRoot(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"github.com/org/repo/internal/client/x", "github.com/org/repo"},
		{"github.com/org/repo/internal/domain/model", "github.com/org/repo"},
		{"github.com/other/lib/pb", "github.com"},
		{"mod/pkg/billing/dal/entity", "mod/pkg/billing"},
		{"svc/client/snapshot", "svc"},
		{"external/pb", "external"},
	}
	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		if got := ModuleRoot(tt.path); got != tt.want {
			t.Errorf("ModuleRoot(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestPkgModuleRoot(t *testing.T) {
	tests := []struct {
		path     string
		wantRoot string
		wantOK   bool
	}{
		{"mod/pkg/billing/dal/entity", "mod/pkg/billing", true},
		{"mod/pkg/billing", "mod/pkg/billing", true},
		{"mod/internal/client/x", "", false},
		{"mod/pkg/", "", false},
	}
	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		root, ok := PkgModuleRoot(tt.path)
		if root != tt.wantRoot || ok != tt.wantOK {
			t.Errorf("PkgModuleRoot(%q) = (%q, %v), want (%q, %v)", tt.path, root, ok, tt.wantRoot, tt.wantOK)
		}
	}
}
