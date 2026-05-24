param(
    [string]$Owner = "kenny0915",
    [string]$RepoName = "heart-lattigo-hw-bootstrap",
    [ValidateSet("public", "private")]
    [string]$Visibility = "public",
    [string]$CommitMessage = "Initial hardware bootstrapping model"
)

$ErrorActionPreference = "Stop"

function Require-Command {
    param([string]$Name)
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Required command '$Name' was not found in PATH. Install it and rerun this script."
    }
}

function Test-NativeCommand {
    param([scriptblock]$Command)

    $oldPreference = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        & $Command *> $null
        return ($LASTEXITCODE -eq 0)
    } finally {
        $ErrorActionPreference = $oldPreference
    }
}

function Invoke-NativeCommand {
    param(
        [scriptblock]$Command,
        [string]$Description
    )

    & $Command
    if ($LASTEXITCODE -ne 0) {
        throw "$Description failed with exit code $LASTEXITCODE."
    }
}

function Get-GitConfigValue {
    param([string]$Key)

    $oldPreference = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        $value = & git config --get $Key 2>$null
        if ($LASTEXITCODE -ne 0) {
            return ""
        }
        return ($value | Select-Object -First 1)
    } finally {
        $ErrorActionPreference = $oldPreference
    }
}

Require-Command git
Require-Command gh

if (-not (Test-Path ".git")) {
    Invoke-NativeCommand { git init } "git init"
}

Invoke-NativeCommand { git branch -M main } "git branch -M main"

$remoteUrl = "https://github.com/$Owner/$RepoName.git"
$hasOrigin = git remote | Select-String -SimpleMatch "origin"
if ($hasOrigin) {
    Invoke-NativeCommand { git remote set-url origin $remoteUrl } "git remote set-url origin"
} else {
    Invoke-NativeCommand { git remote add origin $remoteUrl } "git remote add origin"
}

$gitUserName = Get-GitConfigValue "user.name"
$gitUserEmail = Get-GitConfigValue "user.email"
if ([string]::IsNullOrWhiteSpace($gitUserName) -or [string]::IsNullOrWhiteSpace($gitUserEmail)) {
    throw @"
Git user identity is not configured, so Git cannot create the first commit.

Run these commands once, then rerun this publish script:

git config --global user.name "$Owner"
git config --global user.email "your-email@example.com"
"@
}

Invoke-NativeCommand { git add . } "git add"

$hasCommit = Test-NativeCommand { git rev-parse --verify --quiet HEAD }
$hasStagedChanges = -not (Test-NativeCommand { git diff --cached --quiet })

if (-not $hasCommit -or $hasStagedChanges) {
    Invoke-NativeCommand { git commit -m $CommitMessage } "git commit"
}

if (-not (Test-NativeCommand { gh auth status })) {
    throw "GitHub CLI is not authenticated. Run 'gh auth login' and rerun this script."
}

$repo = "$Owner/$RepoName"
if (-not (Test-NativeCommand { gh repo view $repo })) {
    if ($Visibility -eq "private") {
        Invoke-NativeCommand { gh repo create $repo --private } "gh repo create"
    } else {
        Invoke-NativeCommand { gh repo create $repo --public } "gh repo create"
    }
}

Invoke-NativeCommand { git push -u origin main } "git push"

Write-Host "Published repository: https://github.com/$Owner/$RepoName"
