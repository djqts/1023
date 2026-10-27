param(
    [string]$Version = $(if ($env:VERSION) { $env:VERSION } else { "dev" }),
    [switch]$SkipClient
)

$ErrorActionPreference = "Continue"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

function Test-Image([string]$Image) {
    cmd /c "docker image inspect $Image >NUL 2>NUL"
    return $LASTEXITCODE -eq 0
}

function Resolve-Image([string]$Official, [string]$Mirror) {
    if (Test-Image $Official) { return $Official }
    if (Test-Image $Mirror) { return $Mirror }
    return $Official
}

$goImage = $env:GO_IMAGE
if (-not $goImage) {
    $goImage = Resolve-Image "golang:1.25-alpine" "docker.m.daocloud.io/library/golang:1.25-alpine"
}
$runtimeImage = $env:RUNTIME_IMAGE
if (-not $runtimeImage) {
    $runtimeImage = Resolve-Image "gcr.io/distroless/static-debian12:nonroot" "gcr.m.daocloud.io/distroless/static-debian12:nonroot"
}
$commit = $env:COMMIT
if (-not $commit) { $commit = "unknown" }

Write-Host "building server with GO_IMAGE=$goImage RUNTIME_IMAGE=$runtimeImage"
docker build --target server `
    --build-arg "GO_IMAGE=$goImage" `
    --build-arg "RUNTIME_IMAGE=$runtimeImage" `
    --build-arg "VERSION=$Version" `
    --build-arg "COMMIT=$commit" `
    -t "banner-fingerprint/server:$Version" `
    -f deploy/Dockerfile .
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "building client"
docker build --target client `
    --build-arg "GO_IMAGE=$goImage" `
    --build-arg "RUNTIME_IMAGE=$runtimeImage" `
    --build-arg "VERSION=$Version" `
    --build-arg "COMMIT=$commit" `
    -t "banner-fingerprint/client:$Version" `
    -f deploy/Dockerfile .
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

$env:VERSION = $Version
docker compose -f docker-compose.yml up -d --no-build --wait mysql redis server
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

if (-not $SkipClient) {
    docker compose -f docker-compose.yml run --rm client
    exit $LASTEXITCODE
}
exit 0
