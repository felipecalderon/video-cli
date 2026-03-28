$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Push-Location $root
try {
    $dist = Join-Path $root "dist"
    New-Item -ItemType Directory -Force $dist | Out-Null

    $targets = @(
        @{ goos = "windows"; goarch = "amd64"; ext = ".exe" },
        @{ goos = "windows"; goarch = "arm64"; ext = ".exe" },
        @{ goos = "linux"; goarch = "amd64"; ext = "" },
        @{ goos = "linux"; goarch = "arm64"; ext = "" },
        @{ goos = "darwin"; goarch = "amd64"; ext = "" },
        @{ goos = "darwin"; goarch = "arm64"; ext = "" }
    )

    foreach ($t in $targets) {
        $env:GOOS = $t.goos
        $env:GOARCH = $t.goarch
        $env:CGO_ENABLED = "0"
        $out = Join-Path $dist ("vterminal_" + $t.goos + "_" + $t.goarch + $t.ext)
        Write-Host "Building $out"
        go build -trimpath -ldflags "-s -w" -o $out ./cmd/vterminal
    }
}
finally {
    Pop-Location
}
