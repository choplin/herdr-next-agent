Require herdr
Require jq

Output "../next-agent-demo.gif"

Set Shell bash
Set FontFamily "HackGen Console NF"
Set FontSize 20
Set Width 1868
Set Height 800
Set Theme "Catppuccin Mocha"
Set Padding 20
Set TypingSpeed 35ms
Set CursorBlink false

Hide
Type "/tmp/herdr-next-agent-showcase-config/bin/launch-herdr"
Enter
Wait+Screen@180s /__NEXT_AGENT_CAPTURE_READY__/
Type "clear"
Enter
Sleep 300ms
Show
Sleep 800ms
Type@25ms "# Ctrl+Shift+. / Ctrl+Shift+, moves between blocked Agents."
Enter
Sleep 600ms
Hide
Ctrl+N
Sleep 300ms
Show
Sleep 1400ms
Hide
Ctrl+N
Sleep 300ms
Show
Sleep 1400ms
Hide
Ctrl+N
Sleep 300ms
Show
Sleep 1400ms
Hide
Ctrl+P
Sleep 300ms
Show
Sleep 1400ms
Hide
Ctrl+P
Sleep 300ms
Show
Sleep 1400ms
