$ErrorActionPreference = 'Stop'
$appDir = $PSScriptRoot
$exePath = Join-Path $appDir 'filebrowser.exe'
$existing = Get-Process -Name filebrowser -ErrorAction SilentlyContinue | Where-Object { $_.Path -eq $exePath }
if ($existing) { exit 0 }
$process = Start-Process -FilePath $exePath -ArgumentList @('-c', ('"' + (Join-Path $appDir 'config.yaml') + '"')) -WorkingDirectory $appDir -WindowStyle Hidden -RedirectStandardOutput (Join-Path $appDir 'startup.log') -RedirectStandardError (Join-Path $appDir 'startup-error.log') -PassThru
$process.WaitForExit()
exit $process.ExitCode
