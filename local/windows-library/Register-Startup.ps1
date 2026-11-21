$ErrorActionPreference = 'Stop'
$appDir = $PSScriptRoot
$taskUser = [Security.Principal.WindowsIdentity]::GetCurrent().Name
$taskAction = New-ScheduledTaskAction -Execute (Get-Process -Id $PID).Path -Argument ('-NoProfile -WindowStyle Hidden -File "' + (Join-Path $appDir 'Start-FileBrowser.ps1') + '"') -WorkingDirectory $appDir
$taskTrigger = New-ScheduledTaskTrigger -AtLogOn -User $taskUser
$taskTrigger.Delay = 'PT10S'
$taskPrincipal = New-ScheduledTaskPrincipal -UserId $taskUser -LogonType Interactive -RunLevel Limited
$taskSettings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable -MultipleInstances IgnoreNew -ExecutionTimeLimit ([TimeSpan]::Zero) -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1)
Register-ScheduledTask -TaskName 'FileBrowser Quantum' -Action $taskAction -Trigger $taskTrigger -Principal $taskPrincipal -Settings $taskSettings -Description 'Start FileBrowser Quantum at sign-in.' -Force
