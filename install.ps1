# These skills now live in github.com/diegoclair/harness.
#
# Kept here so existing one-liners and `<skill> update` keep working; it only
# forwards to the new repo's installer.

$ErrorActionPreference = "Stop"
Write-Host "note: diegoclair/skills has moved to diegoclair/harness - forwarding there."

$harness = "https://raw.githubusercontent.com/diegoclair/harness/main/install.ps1"
$script  = (Invoke-WebRequest -Uri $harness -UseBasicParsing).Content
$block   = [ScriptBlock]::Create($script)
& $block @args
exit $LASTEXITCODE
