$ws = New-Object -ComObject WScript.Shell
$menu = [Environment]::GetFolderPath('Programs')
$root = Split-Path -Parent $PSScriptRoot
$exe = Join-Path $root 'build\bin\HyClicker.exe'
$lnk = $ws.CreateShortcut("$menu\HyClicker.lnk")
$lnk.TargetPath = $exe
$lnk.WorkingDirectory = (Split-Path -Parent $exe)
$lnk.IconLocation = (Join-Path (Split-Path -Parent $root) 'icon.ico')
$lnk.Save()
