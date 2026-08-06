// Eval of GID-258: in /domain/**, a function that checks an error must be able
// to return one. The fixtures mirror the 2026-08-06 incident (advertising-api
// internal/domain/service/ad_cabinet_resolver.go): a registry outage logged as
// a warning and dropped, leaving the caller unable to tell it from an empty
// answer.
package service

import "fmt"

// Cabinet is the value being resolved.
type Cabinet struct {
	ID    string
	Title string
}

// Registry is the injected dependency that can fail.
type Registry interface {
	ByIDs(ids []string) ([]Cabinet, error)
	One(id string) (Cabinet, error)
}

// Resolver owns the methods below.
type Resolver struct {
	registry Registry
}

// --- Positive: the incident shape — the failure is logged and the function
// returns nothing at all. ---

func (r *Resolver) resolveChunk(chunk []string) { // want `GID-258: the function checks an error but cannot return one`
	items, err := r.registry.ByIDs(chunk)
	if err != nil {
		fmt.Println("registry unavailable, degrading to id")

		return
	}
	fmt.Println(len(items))
}

// --- Positive: results without an error — (Cabinet, bool) cannot carry the
// failure either, and the caller reads "not found" for an outage. ---

func (r *Resolver) Cabinet(id string) (Cabinet, bool) { // want `GID-258: the function checks an error but cannot return one`
	cabinet, err := r.registry.One(id)
	if err != nil {
		fmt.Println("registry unavailable, degrading to id")

		return Cabinet{}, false
	}

	return cabinet, true
}

// --- Boundary: the comparison is written the other way round (err == nil) —
// still a function that knows the call can fail. ---

func (r *Resolver) countKnown(ids []string) int { // want `GID-258: the function checks an error but cannot return one`
	items, err := r.registry.ByIDs(ids)
	if err == nil {
		return len(items)
	}

	return 0
}

// --- Negative: the error is returned — the shape the rule asks for. ---

func (r *Resolver) resolveChunkClean(chunk []string) (map[string]Cabinet, error) {
	items, err := r.registry.ByIDs(chunk)
	if err != nil {
		return nil, err
	}

	result := make(map[string]Cabinet, len(items))
	for _, item := range items {
		result[item.ID] = item
	}

	return result, nil
}

// --- Negative: a named error result counts just as much as an unnamed one. ---

func (r *Resolver) resolveNamed(chunk []string) (err error) {
	_, err = r.registry.ByIDs(chunk)
	if err != nil {
		return err
	}

	return nil
}

// --- Negative: nothing failable is checked here at all. ---

func (r *Resolver) titles(cabinets []Cabinet) []string {
	titles := make([]string, 0, len(cabinets))
	for _, cabinet := range cabinets {
		titles = append(titles, cabinet.Title)
	}

	return titles
}

// --- Negative: the error is explicitly discarded, never compared to nil —
// there is no handling to speak of (that is errcheck's business). ---

func (r *Resolver) fireAndForget(ids []string) {
	_, _ = r.registry.ByIDs(ids)
}

// --- Boundary: the check lives inside a function LITERAL — a goroutine body
// whose signature belongs to whoever consumes it, so the fix this rule asks
// for does not exist there. Only the enclosing declaration is judged. ---

func (r *Resolver) resolveAsync(ids []string) {
	go func() {
		if _, err := r.registry.ByIDs(ids); err != nil {
			fmt.Println("async resolve failed")
		}
	}()
}
