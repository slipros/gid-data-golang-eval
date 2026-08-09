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
		// A module nested deeper than one segment (resource-registry groups its
		// integrations: pkg/integration/push/firebase) is still a module — its
		// layers are anchored where the layer folders begin.
		{"nested pkg module domain/service", "mod/pkg/integration/push/firebase/domain/service", []string{"domain", "service"}, true},
		{"nested pkg module dal/entity", "mod/pkg/integration/push/firebase/dal/entity", []string{"dal", "entity"}, true},
		{"nested pkg module root has no layer", "mod/pkg/integration/push/firebase", []string{"domain"}, false},
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
		{"mod/pkg/integration/push/firebase/domain/service", []string{"domain", "service"}},
		// No layer pair in the path: the module root stays one segment deep, so
		// the module's own subdirectories read as the layer path.
		{"mod/pkg/integration/push/firebase", []string{"push", "firebase"}},
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
		// Nested module: the root ends where a canonical layer pair begins.
		{"mod/pkg/integration/push/firebase/domain/service", "mod/pkg/integration/push/firebase", true},
		{"mod/pkg/integration/push/firebase/dal/repository/build", "mod/pkg/integration/push/firebase", true},
		// Without a pair there is nothing in the path to prove the nesting is a
		// module, so the one-segment root stands.
		{"mod/pkg/integration/push/firebase", "mod/pkg/integration", true},
		{"mod/pkg/billing/connect/client/x", "mod/pkg/billing", true},
		// pkg/ itself holding the layers: pkg is the root.
		{"mod/pkg/domain/model", "mod/pkg", true},
	}
	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		root, ok := PkgModuleRoot(tt.path)
		if root != tt.wantRoot || ok != tt.wantOK {
			t.Errorf("PkgModuleRoot(%q) = (%q, %v), want (%q, %v)", tt.path, root, ok, tt.wantRoot, tt.wantOK)
		}
	}
}

func TestIndex(t *testing.T) {
	tests := []struct {
		name string
		path string
		seq  []string
		want int
	}{
		{"single segment", "mod/internal/dal/entity", []string{"dal"}, 2},
		{"pair of segments", "mod/internal/dal/entity", []string{"dal", "entity"}, 2},
		{"first occurrence wins", "mod/dal/x/dal/y", []string{"dal"}, 1},
		{"segment matched whole, not by prefix", "mod/internal/events/dto", []string{"event"}, -1},
		{"pair broken by order", "mod/internal/entity/dal", []string{"dal", "entity"}, -1},
		{"absent", "mod/internal/domain/model", []string{"dal"}, -1},
		{"empty seq", "mod/internal/dal", nil, -1},
		{"seq longer than the path", "dal", []string{"dal", "entity"}, -1},
	}

	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		t.Run(tt.name, func(t *testing.T) {
			if got := Index(tt.path, tt.seq...); got != tt.want {
				t.Errorf("Index(%q, %q) = %d, want %d", tt.path, tt.seq, got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name string
		path string
		seq  []string
		want bool
	}{
		{"segment in the middle", "mod/internal/dal/entity", []string{"dal"}, true},
		{"pair of segments", "mod/internal/dal/entity/filter", []string{"dal", "entity"}, true},
		{"nested layer still counts for Contains", "mod/internal/connect/client/x", []string{"client"}, true},
		{"not a prefix match", "mod/internal/dalstats", []string{"dal"}, false},
		{"absent", "mod/internal/domain", []string{"dal"}, false},
		{"empty seq", "mod/internal/dal", nil, false},
	}

	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.path, tt.seq...); got != tt.want {
				t.Errorf("Contains(%q, %q) = %v, want %v", tt.path, tt.seq, got, tt.want)
			}
		})
	}
}

func TestEndsWith(t *testing.T) {
	tests := []struct {
		name string
		path string
		seq  []string
		want bool
	}{
		{"layer root", "mod/internal/dal/repository", []string{"dal", "repository"}, true},
		{"last segment", "mod/internal/domain/model", []string{"model"}, true},
		{"subpackage of the layer is not its root", "mod/internal/dal/repository/build", []string{"dal", "repository"}, false},
		{"segment matched whole", "mod/internal/repositories", []string{"repository"}, false},
		{"seq longer than the path", "dal", []string{"internal", "dal"}, false},
		{"empty seq ends everything", "mod/internal/dal", nil, true},
	}

	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		t.Run(tt.name, func(t *testing.T) {
			if got := EndsWith(tt.path, tt.seq...); got != tt.want {
				t.Errorf("EndsWith(%q, %q) = %v, want %v", tt.path, tt.seq, got, tt.want)
			}
		})
	}
}

func TestSameLibrary(t *testing.T) {
	const uuidLib = "github.com/gofrs/uuid"

	tests := []struct {
		name       string
		importPath string
		library    string
		want       bool
	}{
		{"exact match", uuidLib, uuidLib, true},
		{"major version suffix", uuidLib + "/v5", uuidLib, true},
		{"two-digit major version", uuidLib + "/v11", uuidLib, true},
		{"subpackage is not the library", uuidLib + "/namespace", uuidLib, false},
		{"subpackage of a versioned module", uuidLib + "/v5/namespace", uuidLib, false},
		{"empty version suffix", uuidLib + "/v", uuidLib, false},
		{"another library", "github.com/google/uuid", uuidLib, false},
		{"prefix but not a segment boundary", uuidLib + "x", uuidLib, false},
	}

	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		t.Run(tt.name, func(t *testing.T) {
			if got := SameLibrary(tt.importPath, tt.library); got != tt.want {
				t.Errorf("SameLibrary(%q, %q) = %v, want %v", tt.importPath, tt.library, got, tt.want)
			}
		})
	}
}
