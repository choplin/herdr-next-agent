#!/bin/sh

set -eu

session_name=next-agent-showcase
demo_root=/tmp/herdr-next-agent-showcase
config_root=/tmp/herdr-next-agent-showcase-config
script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
plugin_root=$(CDPATH= cd -- "$script_dir/../.." && pwd)

HERDR_CONFIG_PATH="$config_root/config.toml" \
  herdr session stop "$session_name" >/dev/null 2>&1 || true
HERDR_CONFIG_PATH="$config_root/config.toml" \
  herdr session delete "$session_name" >/dev/null 2>&1 || true

rm -rf -- "$demo_root" "$config_root"
mkdir -p \
  "$demo_root/overview" \
  "$demo_root/payments-api" \
  "$demo_root/storefront" \
  "$demo_root/developer-docs" \
  "$demo_root/release-notes" \
  "$config_root/bin" \
  "$config_root/snapshots"

cp "$script_dir/fixtures/claude.ansi" "$config_root/snapshots/claude.ansi"
cp "$script_dir/fixtures/codex.ansi" "$config_root/snapshots/codex.ansi"
cp "$script_dir/fixtures/cursor.ansi" "$config_root/snapshots/cursor.ansi"

cat >"$config_root/config.toml" <<'EOF'
onboarding = false

[theme]
name = "catppuccin"

[theme.custom]
sidebar_bg = "#11111b"
active_row_bg = "#45475a"
accent = "#f9e2af"

[terminal]
default_shell = "/bin/sh"
shell_mode = "non_login"

[keys]
prefix = "ctrl+b"

[[keys.command]]
key = "ctrl+shift+."
type = "plugin_action"
command = "choplin.next-agent.next"
description = "Focus next blocked Agent"

[[keys.command]]
key = "ctrl+shift+,"
type = "plugin_action"
command = "choplin.next-agent.previous"
description = "Focus previous blocked Agent"

# VHS cannot emit the punctuation shortcuts reliably. These capture-only
# bindings invoke the same real plugin actions without entering agent input.
[[keys.command]]
key = "ctrl+n"
type = "plugin_action"
command = "choplin.next-agent.next"
description = "Focus next blocked Agent"

[[keys.command]]
key = "ctrl+p"
type = "plugin_action"
command = "choplin.next-agent.previous"
description = "Focus previous blocked Agent"

[ui]
sidebar_min_width = 24
sidebar_start_collapsed = false
hide_tab_bar_when_single_tab = true
pane_borders = true
pane_gaps = true
status_indicators = "symbols"

[ui.sidebar.agents]
rows = [
  ["state_icon", "state_text"],
  ["workspace", "agent"],
]
EOF

cat >"$config_root/bin/launch-herdr" <<'EOF'
#!/bin/sh

/tmp/herdr-next-agent-showcase-config/bin/seed-session \
  >/tmp/herdr-next-agent-showcase-config/seed.log 2>&1 &
cd /tmp/herdr-next-agent-showcase/overview

exec env \
  -u NO_COLOR \
  -u HERDR_ENV \
  -u HERDR_PANE_ID \
  -u HERDR_SESSION \
  -u HERDR_SOCKET_PATH \
  -u HERDR_TAB_ID \
  -u HERDR_WORKSPACE_ID \
  HOME="$HOME" \
  PATH="/tmp/herdr-next-agent-showcase-config/bin:$PATH" \
  TERM=xterm-256color \
  COLORTERM=truecolor \
  LANG=en_US.UTF-8 \
  LC_ALL=en_US.UTF-8 \
  PS1='\[\e[1;38;2;166;227;161m\]❯\[\e[0m\] ' \
  PS2='· ' \
  HERDR_CONFIG_PATH=/tmp/herdr-next-agent-showcase-config/config.toml \
  herdr --session next-agent-showcase
EOF

cat >"$config_root/bin/render-agent-screen" <<'EOF'
#!/bin/sh

set -eu

# These ANSI fixtures were captured from real blocked Agent sessions with
# `herdr agent read --source visible --format ansi`. Replaying them keeps the
# README demo deterministic while preserving each Agent's actual interface.
case "$PWD" in
  */payments-api)
    printf '\033[2J\033[H'
    # Captured from the real Claude Code startup banner.
    printf '\033[38;2;215;119;87m'
    printf '                       ▐▛███▛█\n'
    printf '                      ▝▜██████▀\n'
    printf '                        ▝▝ ▝▝\033[0m\n\n'
    exec tail -n +18 /tmp/herdr-next-agent-showcase-config/snapshots/claude.ansi
    ;;
  */developer-docs)
    printf '\033[2J\033[H'
    exec awk '
      index($0, "MCP startup interrupted") { next }
      index($0, "Skill descriptions were shortened") { skip_next = 1; next }
      skip_next { skip_next = 0; next }
      { print }
    ' /tmp/herdr-next-agent-showcase-config/snapshots/codex.ansi
    ;;
  */release-notes)
    printf '\033[2J\033[H'
    exec cat /tmp/herdr-next-agent-showcase-config/snapshots/cursor.ansi
    ;;
esac

reset='\033[0m'
bold='\033[1m'
dim='\033[2m'
green='\033[1;38;2;166;227;161m'
cyan='\033[1;38;2;137;220;235m'

frame_row() {
  style=$1
  content=$2
  width=${3:-68}
  printf '%b│%b ' "$border_color" "$reset"
  printf '%b%-*s%b' "$style" "$width" "$content" "$reset"
  printf ' %b│%b\n' "$border_color" "$reset"
}

case "$PWD" in
  */payments-api)
    border_color='\033[1;38;2;245;169;127m'
    printf '\033[2J\033[H'
    printf '%b╭──────────────────────────────────────────────────────────────────────╮%b\n' "$border_color" "$reset"
    frame_row "$bold" "Claude Code v2.1.239"
    frame_row "$dim" "Sonnet 4.6 · /tmp/herdr-next-agent-showcase/payments-api"
    printf '%b╰──────────────────────────────────────────────────────────────────────╯%b\n\n' "$border_color" "$reset"
    printf '\033[1;38;2;216;177;255m❯\033[0m Push the release branch so CI can publish it.\n\n'
    printf '\033[38;2;216;177;255m⏺\033[0m I updated the release workflow. I need to publish the branch to continue.\n\n'
    printf '%b╭──────────────────────────────────────────────────────────────────────╮%b\n' "$border_color" "$reset"
    frame_row "$bold" "Bash command"
    frame_row "$reset" ""
    frame_row "$cyan" "git push origin release"
    frame_row "$dim" "Publish the release branch to the remote repository"
    frame_row "$reset" ""
    frame_row "$reset" "Do you want to proceed?"
    frame_row "$green" "❯ 1. Yes" 70
    frame_row "$reset" "  2. Yes, and don't ask again for git push commands"
    frame_row "$reset" "  3. No"
    frame_row "$reset" ""
    printf '%b╰──────────────────────────────────────────────────────────────────────╯%b\n' "$border_color" "$reset"
    printf '  \033[2mEsc to cancel · Tab to amend\033[0m\n'
    ;;
  */storefront)
    printf '\033[2J\033[H'
    printf '\033[1mCursor\033[0m\n\n'
    printf ' Implementing the checkout flow…\n'
    ;;
  */developer-docs)
    border_color='\033[2;38;2;147;153;178m'
    printf '\033[2J\033[H'
    printf '%b╭──────────────────────────────────────────────────────────────────────╮%b\n' "$border_color" "$reset"
    frame_row "$bold" ">_ OpenAI Codex (v0.147.0)"
    frame_row "$dim" "model:     gpt-5.6-sol medium"
    frame_row "$dim" "directory: /tmp/herdr-next-agent-showcase/developer-docs"
    printf '%b╰──────────────────────────────────────────────────────────────────────╯%b\n\n' "$border_color" "$reset"
    printf '\033[1m›\033[0m Update the documentation examples and verify them.\n\n'
    printf '\033[38;2;137;220;235m•\033[0m I updated the examples. I\047ll run the test suite before finishing.\n\n'
    printf '%b╭──────────────────────────────────────────────────────────────────────╮%b\n' "$border_color" "$reset"
    frame_row "$bold" "Would you like to run the following command?"
    frame_row "$reset" ""
    frame_row "$cyan" '$ npm test'
    frame_row "$reset" ""
    frame_row "$green" "› 1. Yes, proceed" 70
    frame_row "$reset" "  2. Yes, and don't ask again for this command in this session"
    frame_row "$reset" "  3. No, and tell Codex what to do differently"
    printf '%b╰──────────────────────────────────────────────────────────────────────╯%b\n' "$border_color" "$reset"
    printf '  \033[2mPress enter to confirm or esc to cancel\033[0m\n'
    printf '  \033[2mgpt-5.6-sol medium fast · Context 12%% used\033[0m\n'
    ;;
  */release-notes)
    border_color='\033[1;38;2;137;220;235m'
    printf '\033[2J\033[H'
    printf '\033[1mCursor Agent\033[0m\n'
    printf '\033[2mv2026.08.11-e8db854\033[0m\n\n'
    printf '\033[1m→\033[0m Create a draft release from the prepared notes.\n\n'
    printf '\033[38;2;137;220;235m●\033[0m The release notes are ready. I need to create the draft release.\n\n'
    printf '%b╭──────────────────────────────────────────────────────────────────────╮%b\n' "$border_color" "$reset"
    frame_row "$bold" 'gh release create v2.4.0 --draft --notes-file RELEASE.md'
    frame_row "$reset" ""
    frame_row "$dim" "Create a draft GitHub release from the prepared notes"
    frame_row "$reset" ""
    frame_row "$reset" "This command needs approval"
    frame_row "$green" "▶ Allow once" 70
    frame_row "$reset" "  Always allow gh release create"
    frame_row "$reset" "  Reject"
    frame_row "$reset" ""
    printf '%b╰──────────────────────────────────────────────────────────────────────╯%b\n' "$border_color" "$reset"
    printf '  \033[2m↑/↓ to navigate · Enter to select · Esc to cancel\033[0m\n\n'
    printf '  \033[2mCursor Grok 4.5 High Fast                              Auto-review\033[0m\n'
    printf '  \033[2m/tmp/herdr-next-agent-showcase/release-notes\033[0m\n'
    ;;
  *)
    printf '\033[2J\033[H'
    ;;
esac
EOF

cat >"$config_root/bin/seed-session" <<'EOF'
#!/bin/sh

set -eu

herdr_cli() {
  HERDR_CONFIG_PATH=/tmp/herdr-next-agent-showcase-config/config.toml \
    herdr --session next-agent-showcase "$@"
}

overview_workspace=
for _ in $(seq 1 100); do
  workspaces=$(herdr_cli workspace list 2>/dev/null || true)
  overview_workspace=$(printf '%s' "$workspaces" | jq -r '.result.workspaces[0].workspace_id // empty')
  if [ -n "$overview_workspace" ]; then
    break
  fi
  sleep 0.1
done
if [ -z "$overview_workspace" ]; then
  printf 'Herdr showcase session did not become ready.\n' >&2
  exit 1
fi

overview_pane=$(herdr_cli pane list --workspace "$overview_workspace" | jq -r '.result.panes[0].pane_id')
herdr_cli workspace rename "$overview_workspace" "Overview" >/dev/null
herdr_cli pane run "$overview_pane" \
  "printf '\\033[2J\\033[H'" >/dev/null
herdr_cli pane report-agent "$overview_pane" \
  --source showcase --agent codex --state working --message "Coordinating release" --seq 5 >/dev/null

payments_json=$(herdr_cli workspace create \
  --cwd /tmp/herdr-next-agent-showcase/payments-api \
  --label "Payments API" --no-focus)
payments_pane=$(printf '%s' "$payments_json" | jq -r '.result.root_pane.pane_id')
herdr_cli pane run "$payments_pane" \
  "PS1=''; render-agent-screen" >/dev/null
herdr_cli pane report-agent "$payments_pane" \
  --source showcase --agent claude --state blocked --message "Approval required" --seq 10 >/dev/null

storefront_json=$(herdr_cli workspace create \
  --cwd /tmp/herdr-next-agent-showcase/storefront \
  --label "Storefront" --no-focus)
storefront_pane=$(printf '%s' "$storefront_json" | jq -r '.result.root_pane.pane_id')
herdr_cli pane run "$storefront_pane" \
  "PS1=''; render-agent-screen" >/dev/null
herdr_cli pane report-agent "$storefront_pane" \
  --source showcase --agent cursor --state working --message "Implementing checkout" --seq 20 >/dev/null

docs_json=$(herdr_cli workspace create \
  --cwd /tmp/herdr-next-agent-showcase/developer-docs \
  --label "Developer Docs" --no-focus)
docs_pane=$(printf '%s' "$docs_json" | jq -r '.result.root_pane.pane_id')
herdr_cli pane run "$docs_pane" \
  "PS1=''; render-agent-screen" >/dev/null
herdr_cli pane report-agent "$docs_pane" \
  --source showcase --agent codex --state blocked --message "Approval required" --seq 30 >/dev/null

release_json=$(herdr_cli workspace create \
  --cwd /tmp/herdr-next-agent-showcase/release-notes \
  --label "Release Notes" --no-focus)
release_pane=$(printf '%s' "$release_json" | jq -r '.result.root_pane.pane_id')
herdr_cli pane run "$release_pane" \
  "PS1=''; render-agent-screen" >/dev/null
herdr_cli pane report-agent "$release_pane" \
  --source showcase --agent cursor --state blocked --message "Approval required" --seq 40 >/dev/null

herdr_cli workspace focus "$overview_workspace" >/dev/null
herdr_cli pane run "$overview_pane" \
  "printf '\\033[2J\\033[H__NEXT_AGENT_CAPTURE_READY__\\n'" >/dev/null
EOF
chmod +x \
  "$config_root/bin/launch-herdr" \
  "$config_root/bin/render-agent-screen" \
  "$config_root/bin/seed-session"

HERDR_CONFIG_PATH="$config_root/config.toml" herdr config check

(
  cd "$plugin_root"
  TMPDIR="$config_root" GOCACHE="$config_root/go-build-cache" \
    go build -trimpath -o herdr-next-agent ./cmd/herdr-next-agent
)
HERDR_CONFIG_PATH="$config_root/config.toml" \
  herdr plugin link "$plugin_root" --enabled >/dev/null
printf '__NEXT_AGENT_SETUP_READY__\n'
