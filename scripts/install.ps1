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

# 检查二进制文件是否存在
if (-NOT (Test-Path $BinaryName)) {
    Write-Error "错误：找不到 $BinaryName 二进制文件"
    Write-Host "请先运行 'go build -o aicommit.exe cmd/aicommit/main.go'"
    exit 1
}

# 创建安装目录
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
New-Item -ItemType Directory -Force -Path $ConfigDir | Out-Null

# 复制二进制文件
Copy-Item $BinaryName -Destination $InstallDir

# 添加到系统PATH
if ($PathEnv -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$PathEnv;$InstallDir", "Machine")
}

Write-Host "✓ 安装完成！"
Write-Host "现在你可以使用 'aicommit' 命令了"
Write-Host "注意：你可能需要重新打开PowerShell才能使用该命令" 