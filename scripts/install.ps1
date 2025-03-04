# 检查管理员权限
if (-NOT ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Error "请以管理员身份运行此脚本"
    exit 1
}

# 设置变量
$InstallDir = "$env:ProgramFiles\aicommit"
$BinaryName = "aicommit.exe"
$ConfigDir = "$env:USERPROFILE\.config\aicommit"
$PathEnv = [Environment]::GetEnvironmentVariable("Path", "Machine")
$Repo = "SimonGino/aicommit"

# 检测系统架构
$Arch = if ([Environment]::Is64BitOperatingSystem) {
    if ([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -eq [System.Runtime.InteropServices.Architecture]::Arm64) {
        "arm64"
    } else {
        "x86_64"
    }
} else {
    "i386"
}

# 获取最新版本
Write-Host "正在获取最新版本..."
$LatestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
$Version = $LatestRelease.tag_name

if (-NOT $Version) {
    Write-Error "获取版本信息失败"
    exit 1
}

Write-Host "最新版本: $Version"

# 下载二进制文件
$DownloadUrl = "https://github.com/$Repo/releases/download/$Version/aicommit_windows_$Arch.zip"
Write-Host "正在下载: $DownloadUrl"

$TempDir = New-Item -ItemType Directory -Path "$env:TEMP\aicommit_install" -Force
$ZipFile = Join-Path $TempDir "aicommit.zip"

try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipFile
    Expand-Archive -Path $ZipFile -DestinationPath $TempDir -Force

    # 创建安装目录和配置目录
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    New-Item -ItemType Directory -Force -Path $ConfigDir | Out-Null

    # 复制二进制文件
    Copy-Item -Path (Join-Path $TempDir $BinaryName) -Destination $InstallDir -Force

    # 添加到系统PATH
    if ($PathEnv -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$PathEnv;$InstallDir", "Machine")
    }

    Write-Host "✓ 安装完成！"
    Write-Host "现在你可以使用 'aicommit' 命令了"
    Write-Host "配置目录: $ConfigDir"
    Write-Host "注意：你可能需要重新打开PowerShell才能使用该命令"
}
finally {
    # 清理临时文件
    Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue
} 