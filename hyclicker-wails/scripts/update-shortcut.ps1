$ws = New-Object -ComObject WScript.Shell
$menu = [Environment]::GetFolderPath('Programs')
$lnk = $ws.CreateShortcut("$menu\HyClicker.lnk")
$lnk.TargetPath = "$PSScriptRoot\..\build\bin\HyClicker.exe"
$lnk.WorkingDirectory = "$PSScriptRoot\..\build\bin"
$lnk.IconLocation = "$PSScriptRoot\..\..\icon.ico"
$lnk.Save()
