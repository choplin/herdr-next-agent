// Package navigation selects and focuses Herdr Agents in configured states.
package navigation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// Agent contains the Herdr Agent fields needed for state-based navigation.
type Agent struct {
	PaneID         string `json:"pane_id"`
	Status         string `json:"agent_status"`
	Focused        bool   `json:"focused"`
	StateChangeSeq uint64 `json:"state_change_seq"`
}

// Direction controls which neighboring matching Agent is selected.
type Direction int

const (
	// NextDirection moves forward through the configured order.
	NextDirection Direction = 1
	// PreviousDirection moves backward through the configured order.
	PreviousDirection Direction = -1
)

// CommandResult contains the observable result of an argv command.
type CommandResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	StartErr error
}

// Runner executes argv commands without invoking a shell.
type Runner interface {
	Run(ctx context.Context, argv []string) CommandResult
}

// OSRunner executes commands on the host operating system.
type OSRunner struct{}

// Run implements Runner.
func (OSRunner) Run(ctx context.Context, argv []string) CommandResult {
	if len(argv) == 0 {
		return CommandResult{ExitCode: -1, StartErr: &CommandStartError{}}
	}

	command := exec.CommandContext(ctx, argv[0], argv[1:]...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	result := CommandResult{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: 0}
	if err == nil {
		return result
	}

	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		result.ExitCode = exitError.ExitCode()
		return result
	}
	result.ExitCode = -1
	result.StartErr = err
	return result
}

// CommandStartError indicates that an empty command was passed to a runner.
type CommandStartError struct{}

func (*CommandStartError) Error() string {
	return "cannot start an empty command"
}

// OperationError reports a failed Herdr operation.
type OperationError struct {
	Operation string
	Detail    string
}

func (err *OperationError) Error() string {
	return err.Operation + " failed: " + err.Detail
}

// ProtocolError reports malformed output from the Herdr CLI.
type ProtocolError struct {
	Operation string
}

func (err *ProtocolError) Error() string {
	return err.Operation + " returned an invalid response"
}

// Client is the Herdr API surface used by the action.
type Client interface {
	ListAgents(ctx context.Context) ([]Agent, error)
	FocusAgent(ctx context.Context, paneID string) error
	NotifyNoMatchingAgents(ctx context.Context) error
}

// HerdrClient calls the Herdr CLI through argv commands.
type HerdrClient struct {
	binary string
	runner Runner
}

// NewHerdrClient constructs a CLI-backed Herdr client.
func NewHerdrClient(binary string, runner Runner) *HerdrClient {
	return &HerdrClient{binary: binary, runner: runner}
}

type agentListEnvelope struct {
	Result *agentListResult `json:"result"`
}

type agentListResult struct {
	Agents *[]Agent `json:"agents"`
}

// ListAgents returns all currently detected Agents.
func (client *HerdrClient) ListAgents(ctx context.Context) ([]Agent, error) {
	const operation = "herdr agent list"
	result := client.runner.Run(ctx, []string{client.binary, "agent", "list"})
	if err := commandError(operation, result); err != nil {
		return nil, err
	}

	var response agentListEnvelope
	if err := json.Unmarshal(result.Stdout, &response); err != nil ||
		response.Result == nil || response.Result.Agents == nil {
		return nil, &ProtocolError{Operation: operation}
	}
	return *response.Result.Agents, nil
}

// FocusAgent focuses the pane containing the selected Agent.
func (client *HerdrClient) FocusAgent(ctx context.Context, paneID string) error {
	result := client.runner.Run(ctx, []string{client.binary, "agent", "focus", paneID})
	return commandError("herdr agent focus", result)
}

// NotifyNoMatchingAgents reports that the action has no target.
func (client *HerdrClient) NotifyNoMatchingAgents(ctx context.Context) error {
	result := client.runner.Run(ctx, []string{
		client.binary,
		"notification",
		"show",
		"No matching Agents",
	})
	return commandError("herdr notification show", result)
}

// NextAgent selects the next Agent whose state matches the configuration.
func NextAgent(agents []Agent, states []string, order SelectionOrder) (Agent, bool) {
	return selectAgent(agents, states, order, NextDirection)
}

// PreviousAgent selects the previous Agent whose state matches the configuration.
func PreviousAgent(agents []Agent, states []string, order SelectionOrder) (Agent, bool) {
	return selectAgent(agents, states, order, PreviousDirection)
}

func selectAgent(agents []Agent, states []string, order SelectionOrder, direction Direction) (Agent, bool) {
	matches := func(agent Agent) bool {
		return slices.Contains(states, agent.Status)
	}
	if order != WaitingOrder {
		current := slices.IndexFunc(agents, func(agent Agent) bool {
			return agent.Focused
		})
		if current < 0 && direction == PreviousDirection {
			current = 0
		}
		for offset := 1; offset <= len(agents); offset++ {
			index := wrappedIndex(current+int(direction)*offset, len(agents))
			candidate := agents[index]
			if matches(candidate) {
				return candidate, true
			}
		}
		return Agent{}, false
	}

	matching := make([]Agent, 0, len(agents))
	for _, agent := range agents {
		if matches(agent) {
			matching = append(matching, agent)
		}
	}
	if len(matching) == 0 {
		return Agent{}, false
	}
	sort.SliceStable(matching, func(left, right int) bool {
		return matching[left].StateChangeSeq < matching[right].StateChangeSeq
	})

	current := slices.IndexFunc(matching, func(agent Agent) bool {
		return agent.Focused
	})
	if current < 0 && direction == PreviousDirection {
		current = 0
	}
	return matching[wrappedIndex(current+int(direction), len(matching))], true
}

func wrappedIndex(index, length int) int {
	return (index%length + length) % length
}

// Run focuses the matching Agent in the requested direction or reports that none exist.
func Run(ctx context.Context, client Client, cfg Config, direction Direction) error {
	agents, err := client.ListAgents(ctx)
	if err != nil {
		return err
	}
	var target Agent
	var ok bool
	if direction == PreviousDirection {
		target, ok = PreviousAgent(agents, cfg.States, cfg.Order)
	} else {
		target, ok = NextAgent(agents, cfg.States, cfg.Order)
	}
	if !ok {
		return client.NotifyNoMatchingAgents(ctx)
	}
	return client.FocusAgent(ctx, target.PaneID)
}

func commandError(operation string, result CommandResult) error {
	if result.StartErr != nil {
		return &OperationError{Operation: operation, Detail: result.StartErr.Error()}
	}
	if result.ExitCode == 0 {
		return nil
	}
	detail := strings.TrimSpace(string(result.Stderr))
	if detail == "" {
		detail = "exit status " + strconv.Itoa(result.ExitCode)
	}
	return &OperationError{Operation: operation, Detail: detail}
}
