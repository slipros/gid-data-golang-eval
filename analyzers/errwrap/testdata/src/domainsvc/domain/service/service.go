// Eval of GID-176 (part 2, v2): /domain/** — Wrap of a same-module non-static
// error is forbidden (its stack, if any, was already collected upstream), but
// Wrap of an error from a direct external call (mechanism a) is required —
// the domain may be the first place that reaches out to an external
// dependency, e.g. a DB connection or a stdlib call.
package service

import (
	"strconv"

	"github.com/pkg/errors"

	"domainsvc/domain/model"
	"extgrpc"
)

type Service struct{}

func (s *Service) call() error { return nil }

// --- Positive: Wrap of an incoming same-module non-static error ---

func (s *Service) badWrap() error {
	err := s.call()
	return errors.Wrap(err, "ctx") // want `GID-176: the stack is already collected upstream for a same-module error\. Fix: use errors\.WithMessage instead of errors\.Wrap for an incoming error`
}

func (s *Service) badWrapParam(err error) error {
	return errors.Wrap(err, "ctx") // want `GID-176: the stack is already collected upstream for a same-module error\. Fix: use errors\.WithMessage instead of errors\.Wrap for an incoming error`
}

// --- Negative: WithMessage for a same-module incoming error ---

func (s *Service) goodWithMessage() error {
	err := s.call()
	return errors.WithMessage(err, "ctx")
}

// --- Boundary: Wrap of a static error from model — allowed ---

func (s *Service) goodWrapStatic() error {
	return errors.Wrap(model.ErrSnapshotNotFound, "ctx")
}

// --- Inapplicable: not Wrap (WithStack of an incoming error) — GID-176 part 2 stays silent here ---

func (s *Service) notWrap() error {
	err := s.call()
	return errors.WithStack(err)
}

// --- Positive (v2): a direct external call in /domain/service must be wrapped ---

func (s *Service) badExternalCall() error {
	_, err := strconv.Atoi("x")
	return err // want `GID-176: an error from an external call must be wrapped with errors\.Wrap\. Fix: collect stack and context`
}

// --- Negative (v2): the external call is correctly wrapped — Wrap is required, not forbidden, in domain too ---

func (s *Service) goodExternalWrap() error {
	_, err := strconv.Atoi("x")
	return errors.Wrap(err, "parse")
}

// --- A consumer-side interface over a generated external client ---
//
// The interface is declared here, next to its consumer (GID-134), but its
// variadic option type comes from outside the module: the call goes to an
// external system, so it is a boundary call even in /domain/service.

// RegistryClient — a narrow consumer-side interface over the generated client
// of an external registry.
type RegistryClient interface {
	Integration(id string, opts ...extgrpc.CallOption) (string, error)
}

// LocalRepository — an ordinary injected dependency of this module: its
// variadic option type is local, so the call is not a boundary call.
type LocalRepository interface {
	Snapshot(id string, opts ...model.Option) (string, error)
}

// PlainRepository — the common shape: no variadic parameter at all.
type PlainRepository interface {
	Snapshot(id string) (string, error)
}

type Client struct {
	registry RegistryClient
	local    LocalRepository
	plain    PlainRepository
}

// Positive: the error of an external client call is passed through as is.
func (c *Client) badExternalClient() error {
	_, err := c.registry.Integration("id")
	return err // want `GID-176: an error from an external call must be wrapped with errors\.Wrap\. Fix: collect stack and context`
}

// Positive: WithMessage adds no stack — banned in the service by GID-237 too,
// which is exactly why the boundary has to be visible here.
func (c *Client) badExternalClientWithMessage() error {
	_, err := c.registry.Integration("id")
	return errors.WithMessage(err, "integration") // want `GID-176: an error from an external call must be wrapped with errors\.Wrap \(WithMessage adds no context\)`
}

// Negative: Wrap of an external client error — required, not forbidden, in
// /domain/** (the domain-wrap ban does not apply to a boundary error).
func (c *Client) goodExternalClientWrap() error {
	_, err := c.registry.Integration("id")
	return errors.Wrap(err, "integration")
}

// Non-applicability: the variadic option type belongs to this module — an
// ordinary same-module dependency, so Wrap stays banned here.
func (c *Client) badLocalOptionWrap() error {
	res, err := c.local.Snapshot("id")
	_ = res
	return errors.Wrap(err, "snapshot") // want `GID-176: the stack is already collected upstream for a same-module error\. Fix: use errors\.WithMessage instead of errors\.Wrap for an incoming error`
}

// Non-applicability: no variadic parameter — the usual repository interface.
func (c *Client) goodPlainWithMessage() error {
	_, err := c.plain.Snapshot("id")
	return errors.WithMessage(err, "snapshot")
}
