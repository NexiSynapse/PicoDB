$ErrorActionPreference = "Stop"

$DB = "demo.wal"

if (Test-Path $DB) {
    Remove-Item $DB -Force
}

Write-Host "== Seed durable data =="
.\picodb.exe put $DB alpha one
.\picodb.exe put $DB beta two

Write-Host "== Trigger deterministic crash =="
$env:PICODB_CRASH_AFTER_PREFIX = "1"

$status = 0
try {
    # It will crash with exit code 137 (or similar non-zero code)
    $proc = Start-Process -FilePath ".\picodb.exe" -ArgumentList "put", $DB, "gamma", "THREE" -Wait -NoNewWindow -PassThru
    $status = $proc.ExitCode
} catch {
    # Ignored
}

Write-Host "crashed process status: $status"
Remove-Item Env:\PICODB_CRASH_AFTER_PREFIX

Write-Host "== WAL tail after crash =="
# Emulate `xxd demo.wal | tail -5` using Format-Hex
Format-Hex $DB | Select-Object -Last 5

Write-Host "== Recovery =="

Write-Host "alpha:"
.\picodb.exe get $DB alpha

Write-Host "beta:"
.\picodb.exe get $DB beta

Write-Host "gamma should be missing:"
$proc = Start-Process -FilePath ".\picodb.exe" -ArgumentList "get", $DB, "gamma" -Wait -NoNewWindow -PassThru
$gamma_status = $proc.ExitCode

if ($gamma_status -eq 3) {
    Write-Host "(Exit code 3 verified)"
} else {
    Write-Error "Expected exit code 3, got $gamma_status"
}

Write-Host "== Append after recovery =="
.\picodb.exe put $DB delta four
.\picodb.exe get $DB delta

Write-Host "== Crash recovery demo passed =="
