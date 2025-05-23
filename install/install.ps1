$ErrorActionPreference = "Stop"

$answer = Read-Host "Install in the current folder? (y/n)"
if ($answer -ne "y" -and $answer -ne "Y") {
    Write-Host "Installation cancelled."
    exit
}

Write-Host "Downloading db-portal.exe..."
Invoke-WebRequest -Uri "https://github.com/a-le/db-portal/releases/download/v0.2.0/db-portal.exe" -OutFile "db-portal.exe"

Write-Host "Downloading release ZIP..."
Invoke-WebRequest -Uri "https://github.com/a-le/db-portal/archive/refs/tags/v0.2.0.zip" -OutFile "v0.2.0.zip"

Write-Host "Extracting conf/ and web/ folders..."
Expand-Archive -Path "v0.2.0.zip" -DestinationPath "db-portal-extracted"

Copy-Item -Recurse -Path "db-portal-extracted\db-portal-0.2.0\conf" -Destination ".\conf"
Copy-Item -Recurse -Path "db-portal-extracted\db-portal-0.2.0\web" -Destination ".\web"

Write-Host "Cleaning up..."
Remove-Item "v0.2.0.zip"
Remove-Item -Recurse -Force "db-portal-extracted"

Write-Host "Installation complete."
Write-Host "Run the app with: .\db-portal.exe"
