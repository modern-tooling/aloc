# VHS tape for generating example output screenshot
# Usage: cd to a sample project, then run: vhs /path/to/screenshot.tape
#
# Prerequisites:
# 1. Install aloc: go install github.com/modern-tooling/aloc/cmd/aloc@latest
# 2. cd to a project with git history
# 3. Run: vhs screenshot.tape

Output example-output.png
Set Shell bash
Set FontSize 11
Set Width 850
Set Height 1300
Set Theme "Catppuccin Mocha"
Set Padding 15
Set TypingSpeed 0

Hide
Type "export PS1='$ '" Enter
Sleep 100ms
Type "clear" Enter
Sleep 100ms
Show
Type "aloc . --git" Enter
Sleep 5s
