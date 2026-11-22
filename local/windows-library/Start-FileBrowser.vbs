Option Explicit
Dim shell, fso, appDir, exePath, configPath, svc, processes, proc, exitCode
Set shell = CreateObject("WScript.Shell")
Set fso = CreateObject("Scripting.FileSystemObject")
appDir = fso.GetParentFolderName(WScript.ScriptFullName)
exePath = fso.BuildPath(appDir, "filebrowser.exe")
configPath = fso.BuildPath(appDir, "config.yaml")
Set svc = GetObject("winmgmts:\\.\root\cimv2")
Set processes = svc.ExecQuery("SELECT ExecutablePath FROM Win32_Process WHERE Name='filebrowser.exe'")
For Each proc In processes
    If Not IsNull(proc.ExecutablePath) Then
        If LCase(proc.ExecutablePath) = LCase(exePath) Then WScript.Quit 0
    End If
Next
shell.CurrentDirectory = appDir
exitCode = shell.Run(Chr(34) & exePath & Chr(34) & " -c " & Chr(34) & configPath & Chr(34), 0, True)
WScript.Quit exitCode
