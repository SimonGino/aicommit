# 需要管理员权限运行
if (-NOT ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Warning "Please run this script as Administrator"
    Break
}

# 设置安装目录
$InstallDir = "$env:ProgramFiles\aicommit"

# 从系统路径中移除
$PathEntry = $InstallDir
$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
if ($CurrentPath -like "*$PathEntry*") {
    $NewPath = ($CurrentPath.Split(';') | Where-Object { $_ -ne $PathEntry }) -join ';'
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "Machine")
}

# 删除安装目录
if (Test-Path $InstallDir) {
    Remove-Item -Path $InstallDir -Recurse -Force
    Write-Host "aicommit has been uninstalled."
} else {
    Write-Host "aicommit is not installed in $InstallDir"
}

# 删除配置文件（可选）
$ConfigDir = "$env:USERPROFILE\.config\aicommit"
if (Test-Path $ConfigDir) {
    $Response = Read-Host "Do you want to remove configuration files? [y/N]"
    if ($Response -eq 'y' -or $Response -eq 'Y') {
        Remove-Item -Path $ConfigDir -Recurse -Force
        Write-Host "Configuration files removed."
    }
} 