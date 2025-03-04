# 检查管理员权限
if (-NOT ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Error "请以管理员身份运行此脚本"
    exit 1
}

# 设置变量
$InstallDir = "$env:ProgramFiles\aicommit"
$ConfigDir = "$env:USERPROFILE\.config\aicommit"
$PathEnv = [Environment]::GetEnvironmentVariable("Path", "Machine")

# 删除二进制文件和安装目录
if (Test-Path $InstallDir) {
    Remove-Item -Path $InstallDir -Recurse -Force
    Write-Host "✓ 已删除程序文件"
} else {
    Write-Host "警告：找不到程序目录"
}

# 从系统PATH中移除
if ($PathEnv -like "*$InstallDir*") {
    $NewPath = ($PathEnv.Split(';') | Where-Object { $_ -ne $InstallDir }) -join ';'
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "Machine")
    Write-Host "✓ 已从系统PATH中移除"
}

# 询问是否删除配置文件
$Response = Read-Host "是否删除配置文件？(y/N)"
if ($Response -eq 'y' -or $Response -eq 'Y') {
    if (Test-Path $ConfigDir) {
        Remove-Item -Path $ConfigDir -Recurse -Force
        Write-Host "✓ 已删除配置文件"
    } else {
        Write-Host "警告：找不到配置目录"
    }
}

Write-Host "✓ 卸载完成！" 