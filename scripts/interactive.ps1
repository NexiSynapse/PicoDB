$ErrorActionPreference = "Continue"

# Resolve the repo root from this script's location (scripts/) so it works
# from any clone, not just a specific absolute path.
$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $repoRoot
$db = "mydata.wal"
$exe = "$repoRoot\picodb.exe"

if (Test-Path $db) { Remove-Item $db }

function ShowMenu {
    Write-Host ""
    Write-Host "=============================================="
    Write-Host "   PicoDB interactive store  (db: $db)"
    Write-Host "=============================================="
    Write-Host "   1) Put / create a key"
    Write-Host "   2) Get / read a key"
    Write-Host "   3) Delete a key"
    Write-Host "   4) List all keys"
    Write-Host "   5) Clear all data"
    Write-Host "   0) Exit (keeps data saved)"
    Write-Host "=============================================="
}

function ListData {
    # dump isn't implemented, so rebuild from the session index cache instead
    Write-Host "  Keys stored so far in this session:"
    if ($script:keys.Count -eq 0) {
        Write-Host "    (none)"
    } else {
        foreach ($k in ($script:keys.Keys | Sort-Object)) {
            Write-Host "    $k = $($script:keys[$k])"
        }
    }
}

$script:keys = @{}

while ($true) {
    ShowMenu
    $choice = Read-Host "  Choose an option"
    switch ($choice) {
        "1" {
            $k = Read-Host "  Key"
            if ([string]::IsNullOrWhiteSpace($k)) { Write-Host "  Key cannot be empty."; continue }
            $v = Read-Host "  Value"
            if ([string]::IsNullOrWhiteSpace($v)) { Write-Host "  Value cannot be empty."; continue }
            $out = & $exe put $db $k $v 2>&1
            if ($LASTEXITCODE -eq 0) {
                $script:keys[$k] = $v
                Write-Host "  Saved: $k = $v"
            } else {
                Write-Host "  Error: $out"
            }
        }
        "2" {
            $k = Read-Host "  Key"
            $out = & $exe get $db $k 2>&1
            if ($LASTEXITCODE -eq 0) { Write-Host "  $k = $out" } else { Write-Host "  $k = <not found>" }
        }
        "3" {
            $k = Read-Host "  Key"
            $out = & $exe del $db $k 2>&1
            if ($LASTEXITCODE -eq 0) {
                $script:keys.Remove($k) | Out-Null
                Write-Host "  Deleted: $k"
            } else {
                Write-Host "  Not deleted: $out"
            }
        }
        "4" { ListData }
        "5" {
            Remove-Item $db -ErrorAction SilentlyContinue
            $script:keys = @{}
            Write-Host "  All data cleared."
        }
        "0" {
            Write-Host ""
            Write-Host "  Data saved to $db in $repoRoot"
            Write-Host "  Goodbye!"
            exit
        }
        default { Write-Host "  Invalid option. Try 1-5 or 0." }
    }
}
