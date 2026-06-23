# ============================================================================
#  loadgen.ps1 - PEMBANGKIT BEBAN untuk Windows (PowerShell)
# ----------------------------------------------------------------------------
#  Versi Windows dari loadgen.sh. Jalankan di PowerShell.
#
#  CARA PAKAI:
#    .\loadgen.ps1                 # skenario campuran (default)
#    .\loadgen.ps1 -Scenario slow  # request lambat
#    .\loadgen.ps1 -Scenario error # request gagal
#    .\loadgen.ps1 -Scenario checkout
#
#  Jika muncul error "running scripts is disabled", jalankan dulu:
#    Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
#
#  Hentikan dengan Ctrl + C.
# ============================================================================

param(
  [string]$Scenario = "mix",
  [string]$BaseUrl  = "http://localhost:8080"
)

Write-Host "Load generator -> $BaseUrl  (skenario: $Scenario)"
Write-Host "Tekan Ctrl+C untuk berhenti.`n"

function Hit($url) {
  try {
    $resp = Invoke-WebRequest -Uri $url -UseBasicParsing -ErrorAction Stop
    Write-Host ("  {0}  <-  {1}" -f $resp.StatusCode, $url)
  } catch {
    # Request yang sengaja gagal (HTTP 500) masuk ke sini.
    $code = $_.Exception.Response.StatusCode.value__
    Write-Host ("  {0}  <-  {1}" -f $code, $url)
  }
}

while ($true) {
  switch ($Scenario) {
    "normal"   { Hit "$BaseUrl/work" }
    "slow"     { Hit "$BaseUrl/work?slow=1" }
    "error"    { Hit "$BaseUrl/work?fail=1" }
    "checkout" { Hit "$BaseUrl/checkout" }
    default {
      Hit "$BaseUrl/work"
      Hit "$BaseUrl/work"
      Hit "$BaseUrl/work?slow=1"
      Hit "$BaseUrl/work?fail=1"
      Hit "$BaseUrl/checkout"
    }
  }
  Start-Sleep -Milliseconds 200
}
