$ErrorActionPreference = "Stop"

$answer = Read-Host "Install in the current folder? (y/n)"
if ($answer -ne "y" -and $answer -ne "Y") {
    Write-Host "Installation cancelled."
    exit
}

Write-Host "Fetching latest release info..."
$releaseInfo = Invoke-RestMethod -Uri "https://api.github.com/repos/a-le/db-portal/releases/latest"
$latestTag = $releaseInfo.tag_name

Write-Host "Latest version detected: $latestTag"

Write-Host "Downloading db-portal.exe..."
Invoke-WebRequest -Uri "https://github.com/a-le/db-portal/releases/download/$latestTag/db-portal.exe" -OutFile "db-portal.exe"

Write-Host "Downloading release ZIP..."
Invoke-WebRequest -Uri "https://github.com/a-le/db-portal/archive/refs/tags/$latestTag.zip" -OutFile "$latestTag.zip"

Write-Host "Extracting conf/ and web/ folders..."
Expand-Archive -Path "$latestTag.zip" -DestinationPath "db-portal-extracted"

Copy-Item -Recurse -Path "db-portal-extracted\db-portal-$($latestTag.TrimStart('v'))\conf" -Destination ".\conf"
Copy-Item -Recurse -Path "db-portal-extracted\db-portal-$($latestTag.TrimStart('v'))\web" -Destination ".\web"

Write-Host "Cleaning up..."
Remove-Item "$latestTag.zip"
Remove-Item -Recurse -Force "db-portal-extracted"

Write-Host "Installation complete."
Write-Host 'Run the app with: .\db-portal.exe (add --set-master-password="your password" argument on the first run)'
