$package = "github.com/mpotthoff/git-lfs-webdav"
$projectname = "git-lfs-webdav"
$outputdir = "build"

# Check dirty repo
git diff --no-patch --exit-code
if ($LASTEXITCODE -ne 0) {
    Write-Output "Working copy is not clean (unstaged changes)"
    Exit $LASTEXITCODE
}
git diff --no-patch --cached --exit-code
if ($LASTEXITCODE -ne 0) {
    Write-Output "Working copy is not clean (staged changes)"
    Exit $LASTEXITCODE
}

# Check that the latest tag is present directly on HEAD
$version = (git describe --exact-match | Out-String).Trim()
if ($LASTEXITCODE -ne 0) {
    Write-Output "No version tag on HEAD"
    Exit $LASTEXITCODE
}

# Get the Git-Hash
$hash = (git rev-parse --short HEAD | Out-String).Trim()
if ($LASTEXITCODE -ne 0) {
    Write-Output "Failed to get hash of HEAD"
    Exit $LASTEXITCODE
}

Write-Output "Building version: $version"

$env:GOOS="windows"
$env:GOARCH="amd64"
go build -ldflags "-s -w -X $package/internal.Version=$version-$hash" -o "$outputdir/$projectname-windows-amd64.exe"

$env:GOOS="linux"
$env:GOARCH="amd64"
go build -ldflags "-s -w -X $package/internal.Version=$version-$hash" -o "$outputdir/$projectname-linux-amd64"

$env:GOOS="darwin"
$env:GOARCH="amd64"
go build -ldflags "-s -w -X $package/internal.Version=$version-$hash" -o "$outputdir/$projectname-darwin-amd64"
