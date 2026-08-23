package navigation

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestNextAgent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		agents []Agent
		states []string
		order  SelectionOrder
		want   string
		ok     bool
	}{
		{
			name: "no blocked Agents",
			agents: []Agent{
				{PaneID: "w1:p1", Status: "working", Focused: true},
				{PaneID: "w2:p1", Status: "done"},
			},
		},
		{
			name: "first displayed blocked Agent when focus is elsewhere",
			agents: []Agent{
				{PaneID: "w1:p1", Status: "working", Focused: true},
				{PaneID: "w2:p1", Status: "blocked"},
				{PaneID: "w3:p1", Status: "blocked"},
			},
			want: "w2:p1",
			ok:   true,
		},
		{
			name: "longest-waiting blocked Agent when focus is elsewhere",
			agents: []Agent{
				{PaneID: "w1:p1", Status: "working", Focused: true, StateChangeSeq: 1},
				{PaneID: "w2:p1", Status: "blocked", StateChangeSeq: 30},
				{PaneID: "w3:p1", Status: "blocked", StateChangeSeq: 10},
			},
			order: WaitingOrder,
			want:  "w3:p1",
			ok:    true,
		},
		{
			name: "next displayed blocked Agent",
			agents: []Agent{
				{PaneID: "w3:p1", Status: "blocked"},
				{PaneID: "w1:p1", Status: "blocked", Focused: true},
				{PaneID: "w4:p1", Status: "working"},
				{PaneID: "w2:p1", Status: "blocked"},
			},
			want: "w2:p1",
			ok:   true,
		},
		{
			name: "focused non-blocked Agent remains the display cursor",
			agents: []Agent{
				{PaneID: "w1:p1", Status: "blocked"},
				{PaneID: "w2:p1", Status: "idle", Focused: true},
				{PaneID: "w3:p1", Status: "working"},
				{PaneID: "w4:p1", Status: "blocked"},
			},
			want: "w4:p1",
			ok:   true,
		},
		{
			name: "wrap after last displayed blocked Agent",
			agents: []Agent{
				{PaneID: "w1:p1", Status: "blocked"},
				{PaneID: "w2:p1", Status: "blocked", Focused: true},
			},
			want: "w1:p1",
			ok:   true,
		},
		{
			name: "Herdr display order is preserved",
			agents: []Agent{
				{PaneID: "w2:p1", Status: "blocked"},
				{PaneID: "w1:p1", Status: "blocked"},
			},
			want: "w2:p1",
			ok:   true,
		},
		{
			name: "single focused blocked Agent wraps to itself",
			agents: []Agent{
				{PaneID: "w1:p1", Status: "working"},
				{PaneID: "w2:p1", Status: "blocked", Focused: true},
			},
			want: "w2:p1",
			ok:   true,
		},
		{
			name: "display order breaks waiting sequence ties",
			agents: []Agent{
				{PaneID: "w2:p1", Status: "blocked", StateChangeSeq: 10},
				{PaneID: "w1:p1", Status: "blocked", StateChangeSeq: 10},
			},
			order: WaitingOrder,
			want:  "w2:p1",
			ok:    true,
		},
		{
			name: "configured states select more than blocked Agents",
			agents: []Agent{
				{PaneID: "w1:p1", Status: "blocked", Focused: true},
				{PaneID: "w2:p1", Status: "working"},
				{PaneID: "w3:p1", Status: "done"},
			},
			states: []string{"blocked", "done"},
			want:   "w3:p1",
			ok:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			states := test.states
			if states == nil {
				states = []string{"blocked"}
			}
			got, ok := NextAgent(test.agents, states, test.order)
			if ok != test.ok || got.PaneID != test.want {
				t.Fatalf("NextAgent() = (%q, %t), want (%q, %t)", got.PaneID, ok, test.want, test.ok)
			}
		})
	}
}

func TestPreviousAgent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		agents []Agent
		states []string
		order  SelectionOrder
		want   string
		ok     bool
	}{
		{
			name: "previous displayed matching Agent",
			agents: []Agent{
				{PaneID: "w1:p1", Status: "blocked"},
				{PaneID: "w2:p1", Status: "working"},
				{PaneID: "w3:p1", Status: "blocked", Focused: true},
			},
			states: []string{"blocked"},
			want:   "w1:p1",
			ok:     true,
		},
		{
			name: "wrap before first displayed matching Agent",
			agents: []Agent{
				{PaneID: "w1:p1", Status: "blocked", Focused: true},
				{PaneID: "w2:p1", Status: "blocked"},
			},
			states: []string{"blocked"},
			want:   "w2:p1",
			ok:     true,
		},
		{
			name: "last displayed match when focus is elsewhere",
			agents: []Agent{
				{PaneID: "w1:p1", Status: "blocked"},
				{PaneID: "w2:p1", Status: "working"},
				{PaneID: "w3:p1", Status: "done"},
			},
			states: []string{"blocked", "done"},
			want:   "w3:p1",
			ok:     true,
		},
		{
			name: "previous in state-change order",
			agents: []Agent{
				{PaneID: "w1:p1", Status: "blocked", StateChangeSeq: 10},
				{PaneID: "w2:p1", Status: "blocked", StateChangeSeq: 30, Focused: true},
				{PaneID: "w3:p1", Status: "blocked", StateChangeSeq: 20},
			},
			states: []string{"blocked"},
			order:  WaitingOrder,
			want:   "w3:p1",
			ok:     true,
		},
		{
			name: "newest state change when focus is elsewhere",
			agents: []Agent{
				{PaneID: "w1:p1", Status: "blocked", StateChangeSeq: 10},
				{PaneID: "w2:p1", Status: "working", Focused: true, StateChangeSeq: 40},
				{PaneID: "w3:p1", Status: "blocked", StateChangeSeq: 30},
			},
			states: []string{"blocked"},
			order:  WaitingOrder,
			want:   "w3:p1",
			ok:     true,
		},
		{
			name:   "no matching Agents",
			agents: []Agent{{PaneID: "w1:p1", Status: "working", Focused: true}},
			states: []string{"blocked"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, ok := PreviousAgent(test.agents, test.states, test.order)
			if ok != test.ok || got.PaneID != test.want {
				t.Fatalf("PreviousAgent() = (%q, %t), want (%q, %t)", got.PaneID, ok, test.want, test.ok)
			}
		})
	}
}

type fakeClient struct {
	agents      []Agent
	listErr     error
	focusErr    error
	notifyErr   error
	focusedPane string
	notified    bool
}

func (client *fakeClient) ListAgents(context.Context) ([]Agent, error) {
	return client.agents, client.listErr
}

func (client *fakeClient) FocusAgent(_ context.Context, paneID string) error {
	client.focusedPane = paneID
	return client.focusErr
}

func (client *fakeClient) NotifyNoMatchingAgents(context.Context) error {
	client.notified = true
	return client.notifyErr
}

func TestRunFocusesSelectedAgent(t *testing.T) {
	t.Parallel()

	client := &fakeClient{agents: []Agent{
		{PaneID: "w1:p1", Status: "working", Focused: true},
		{PaneID: "w2:p1", Status: "blocked"},
	}}
	if err := Run(context.Background(), client, DefaultConfig(), NextDirection); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if client.focusedPane != "w2:p1" || client.notified {
		t.Fatalf("client = %#v, want w2:p1 focused without notification", client)
	}
}

func TestRunFocusesPreviousSelectedAgent(t *testing.T) {
	t.Parallel()

	client := &fakeClient{agents: []Agent{
		{PaneID: "w1:p1", Status: "blocked"},
		{PaneID: "w2:p1", Status: "blocked", Focused: true},
	}}
	if err := Run(context.Background(), client, DefaultConfig(), PreviousDirection); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if client.focusedPane != "w1:p1" || client.notified {
		t.Fatalf("client = %#v, want w1:p1 focused without notification", client)
	}
}

func TestRunNotifiesWhenNoAgentMatches(t *testing.T) {
	t.Parallel()

	client := &fakeClient{agents: []Agent{{PaneID: "w1:p1", Status: "idle", Focused: true}}}
	if err := Run(context.Background(), client, DefaultConfig(), NextDirection); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !client.notified || client.focusedPane != "" {
		t.Fatalf("client = %#v, want notification without focus", client)
	}
}

func TestRunStopsWhenListingAgentsFails(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("list failed")
	client := &fakeClient{listErr: wantErr}
	if err := Run(context.Background(), client, DefaultConfig(), NextDirection); !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
	if client.notified || client.focusedPane != "" {
		t.Fatalf("client = %#v, want no side effects", client)
	}
}

func TestRunReturnsActionError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		client *fakeClient
	}{
		{
			name: "focus",
			client: &fakeClient{
				agents:   []Agent{{PaneID: "w1:p1", Status: "blocked"}},
				focusErr: errors.New("focus failed"),
			},
		},
		{
			name:   "notification",
			client: &fakeClient{notifyErr: errors.New("notification failed")},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if err := Run(context.Background(), test.client, DefaultConfig(), NextDirection); err == nil {
				t.Fatal("Run() error = nil, want action error")
			}
		})
	}
}

type recordingRunner struct {
	responses []CommandResult
	calls     [][]string
}

func (runner *recordingRunner) Run(_ context.Context, argv []string) CommandResult {
	runner.calls = append(runner.calls, slices.Clone(argv))
	response := runner.responses[0]
	runner.responses = runner.responses[1:]
	return response
}

func TestHerdrClientUsesArgvAndParsesAgentList(t *testing.T) {
	t.Parallel()

	stdout, err := json.Marshal(map[string]any{
		"result": map[string]any{
			"agents": []Agent{{
				PaneID:         "w1:p1",
				Status:         "blocked",
				Focused:        true,
				StateChangeSeq: 42,
			}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	runner := &recordingRunner{responses: []CommandResult{
		{Stdout: stdout, ExitCode: 0},
		{Stdout: []byte("{}"), ExitCode: 0},
		{Stdout: []byte("{}"), ExitCode: 0},
	}}
	client := NewHerdrClient("herdr-test", runner)
	agents, err := client.ListAgents(context.Background())
	if err != nil {
		t.Fatalf("ListAgents() error = %v", err)
	}
	if len(agents) != 1 || agents[0].PaneID != "w1:p1" || agents[0].Status != "blocked" {
		t.Fatalf("agents = %#v, want one blocked Agent", agents)
	}
	if err := client.FocusAgent(context.Background(), agents[0].PaneID); err != nil {
		t.Fatalf("FocusAgent() error = %v", err)
	}
	if err := client.NotifyNoMatchingAgents(context.Background()); err != nil {
		t.Fatalf("NotifyNoMatchingAgents() error = %v", err)
	}
	wantCalls := [][]string{
		{"herdr-test", "agent", "list"},
		{"herdr-test", "agent", "focus", "w1:p1"},
		{"herdr-test", "notification", "show", "No matching Agents"},
	}
	if !slices.EqualFunc(runner.calls, wantCalls, func(left, right []string) bool {
		return slices.Equal(left, right)
	}) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, wantCalls)
	}
}

func TestHerdrClientRejectsInvalidAgentList(t *testing.T) {
	t.Parallel()

	for _, response := range []string{
		`not JSON`,
		`{"result":null}`,
		`{"result":{}}`,
		`{"result":{"agents":null}}`,
	} {
		runner := &recordingRunner{responses: []CommandResult{{Stdout: []byte(response)}}}
		client := NewHerdrClient("herdr-test", runner)
		_, err := client.ListAgents(context.Background())
		var protocolErr *ProtocolError
		if !errors.As(err, &protocolErr) {
			t.Errorf("ListAgents() error = %v, want ProtocolError for %q", err, response)
		}
	}
}

func TestHerdrClientReportsCommandFailure(t *testing.T) {
	t.Parallel()

	runner := &recordingRunner{responses: []CommandResult{{
		Stderr:   []byte("socket unavailable\n"),
		ExitCode: 1,
	}}}
	client := NewHerdrClient("herdr-test", runner)
	_, err := client.ListAgents(context.Background())
	var operationErr *OperationError
	if !errors.As(err, &operationErr) || operationErr.Detail != "socket unavailable" {
		t.Fatalf("ListAgents() error = %#v, want socket failure OperationError", err)
	}
}

func TestManifestBuildsAndRunsNativeCommand(t *testing.T) {
	t.Parallel()

	manifest, err := os.ReadFile(filepath.Join("..", "..", "herdr-plugin.toml"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	contents := string(manifest)
	if !strings.Contains(contents, `command = ["go", "build", "-trimpath", "-o", "herdr-next-agent", "./cmd/herdr-next-agent"]`) {
		t.Error("manifest does not build the Go command directly")
	}
	if !strings.Contains(contents, `command = ["./herdr-next-agent", "next"]`) {
		t.Error("manifest does not run the next action directly")
	}
	if !strings.Contains(contents, `command = ["./herdr-next-agent", "previous"]`) {
		t.Error("manifest does not run the previous action directly")
	}
	if strings.Contains(contents, "python") || strings.Contains(contents, `command = ["sh"`) {
		t.Error("manifest unexpectedly invokes Python or a shell")
	}
}
