$ErrorActionPreference = "Stop"

$answer = Read-Host "Install in the current folder? (y/n)"
if ($answer -ne "y" -and $answer -ne "Y") {
    Write-Host "Installation cancelled."
    exit
}

Write-Host "Fetching latest release info..."
$releaseInfo = Invoke-RestMethod -Uri "https://api.github.com/repos/a-le/db-portal/releases/latest"
$latestTag = $releaseInfo.tag_name

# The top-level directory inside the archive follows this naming convention: {repository-name}-{tag-name-without-v-prefix}
$releaseTempFolder = "db-portal-$($latestTag -replace '^v', '')"

Write-Host "Latest version detected: $latestTag"

Write-Host "Downloading db-portal.exe..."
Start-BitsTransfer -Source "https://github.com/a-le/db-portal/releases/download/$latestTag/db-portal.exe" -Destination "db-portal.exe"

Write-Host "Downloading release ZIP..."
Start-BitsTransfer -Source "https://github.com/a-le/db-portal/archive/refs/tags/$latestTag.zip" -Destination "$latestTag.zip"

Write-Host "Extracting release ZIP..."
Expand-Archive -Path "$latestTag.zip" -DestinationPath "."

Write-Host "Write config files, keeping existing files..."
Copy-Item  -Recurse -Path "$releaseTempFolder\config" -Destination ".\config"

Write-Host "Write web files, overwriting existing files..."
Copy-Item -Force -Recurse -Path "$releaseTempFolder\web" -Destination ".\web"

Write-Host "Cleaning up..."
Remove-Item "$latestTag.zip"
Remove-Item -Recurse -Force "$releaseTempFolder"

Write-Host "Installation complete."
Write-Host 'Run the app and set master password with: .\db-portal.exe --set-master-password="your password"'
Write-Host 'the --set-master-password argument is only needed on the first run, or to reset the master password'
