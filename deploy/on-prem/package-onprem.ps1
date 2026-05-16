param(
  [string]$PackageName = "UnityAid-onprem",
  [string]$OutputRoot = "./dist",
  [string]$AppVersion = "local"
)

$ErrorActionPreference = "Stop"

$scriptRoot = $PSScriptRoot
$repoRoot = Resolve-Path (Join-Path $scriptRoot "../..")
$stamp = Get-Date -Format "yyyyMMdd_HHmmss"
$packagePath = Join-Path (Join-Path $scriptRoot $OutputRoot) "$PackageName-$AppVersion-$stamp"

New-Item -ItemType Directory -Force -Path $packagePath | Out-Null
Copy-Item -LiteralPath (Join-Path $repoRoot "deploy/client") -Destination (Join-Path $packagePath "client") -Recurse -Force

$docsPath = Join-Path $packagePath "docs"
New-Item -ItemType Directory -Force -Path $docsPath | Out-Null
Copy-Item -LiteralPath (Join-Path $repoRoot "README.md") -Destination (Join-Path $docsPath "README.md") -Force
Copy-Item -LiteralPath (Join-Path $repoRoot "docs/roadmaps/saas_product_roadmap.md") -Destination (Join-Path $docsPath "saas_product_roadmap.md") -Force

$manifest = [ordered]@{
  packageName = $PackageName
  appVersion = $AppVersion
  createdAt = (Get-Date).ToUniversalTime().ToString("o")
  includes = @("client compose stack", "backup restore scripts", "export import scripts", "on-prem guide")
}
$manifest | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath (Join-Path $packagePath "manifest.json") -Encoding UTF8

Write-Output "On-prem package created: $packagePath"
