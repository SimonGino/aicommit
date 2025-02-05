# 需要管理员权限运行
if (-NOT ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Warning "Please run this script as Administrator"
    Break
}

# 设置安装目录
$InstallDir = "$env:ProgramFiles\aicommit"
$BinaryPath = "$InstallDir\aicommit.exe"

# 创建安装目录
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}

# 获取最新版本
Write-Host "Fetching latest release..."
$LatestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/SimonGino/aicommit/releases/latest"
$DownloadUrl = $LatestRelease.assets | Where-Object { $_.name -eq "aicommit-windows.exe" } | Select-Object -ExpandProperty browser_download_url

if (-not $DownloadUrl) {
    Write-Error "Could not find latest release"
    Exit 1
}

# 下载二进制文件
Write-Host "Downloading aicommit..."
Invoke-WebRequest -Uri $DownloadUrl -OutFile $BinaryPath

# 添加到系统路径
$PathEntry = $InstallDir
$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
if ($CurrentPath -notlike "*$PathEntry*") {
    [Environment]::SetEnvironmentVariable("Path", "$CurrentPath;$PathEntry", "Machine")
}

Write-Host "Installation completed! You can now use 'aicommit' command."
Write-Host "Try 'aicommit --help' to get started." 