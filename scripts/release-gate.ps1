param(
  [switch]$Http
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $repoRoot

function Run-Step {
  param(
    [string]$Name,
    [scriptblock]$Script
  )
  Write-Host "==> $Name"
  & $Script
  Write-Host "OK: $Name"
}

function Invoke-Json {
  param(
    [string]$Method = "GET",
    [string]$Uri,
    [string]$Body = "{}"
  )
  if ($Method -eq "GET") {
    return Invoke-RestMethod -Method Get -Uri $Uri
  }
  return Invoke-RestMethod -Method $Method -Uri $Uri -ContentType "application/json" -Body $Body
}

function Wait-Json {
  param(
    [string]$Uri,
    [int]$Attempts = 30,
    [int]$DelaySeconds = 1
  )
  for ($i = 1; $i -le $Attempts; $i++) {
    try {
      return Invoke-Json -Uri $Uri
    } catch {
      if ($i -eq $Attempts) {
        throw
      }
      Start-Sleep -Seconds $DelaySeconds
    }
  }
}

Run-Step -Name "Go 服务单测" -Script {
  go test ./services/control-plane/... ./services/ad-integration/...
}

Run-Step -Name "lead-lifecycle 单测" -Script {
  $env:UV_CACHE_DIR = ".cache\uv"
  $env:PYTHONPATH = "services\lead-lifecycle\src"
  uv run python -m unittest services/lead-lifecycle/src/lead_lifecycle/service_test.py
}

Run-Step -Name "data-insight 单测" -Script {
  $env:UV_CACHE_DIR = ".cache\uv"
  $env:PYTHONPATH = "services\data-insight\src"
  uv run python -m unittest `
    services/data-insight/src/data_insight/etl_test.py `
    services/data-insight/src/data_insight/bi_test.py `
    services/data-insight/src/data_insight/bi_config_test.py `
    services/data-insight/src/data_insight/reports_test.py
}

Run-Step -Name "virtual-account 单测" -Script {
  $env:UV_CACHE_DIR = ".cache\uv"
  $env:PYTHONPATH = "services\virtual-account\src"
  uv run python -m unittest services/virtual-account/src/virtual_account/service_test.py
}

Run-Step -Name "前端生产构建" -Script {
  npm.cmd --prefix apps/frontend-web run build
}

if ($Http) {
  Run-Step -Name "服务健康检查" -Script {
    Wait-Json -Uri "http://127.0.0.1:8080/healthz" | Out-Null
    Wait-Json -Uri "http://127.0.0.1:8081/healthz" | Out-Null
    Wait-Json -Uri "http://127.0.0.1:8090/healthz" | Out-Null
    Wait-Json -Uri "http://127.0.0.1:8091/healthz" | Out-Null
    Wait-Json -Uri "http://127.0.0.1:8092/healthz" | Out-Null
  }

  Run-Step -Name "广告同步探针" -Script {
    $sync = Invoke-Json -Method "POST" -Uri "http://127.0.0.1:8081/api/ad-sync/run" -Body '{"platform":"douyin","date_from":"2026-05-20","date_to":"2026-05-20"}'
    if (-not $sync.success -or $sync.data.entity_count -lt 1) { throw "ad sync probe failed" }
  }

  Run-Step -Name "线索冲突探针" -Script {
    $batch = Invoke-Json -Method "POST" -Uri "http://127.0.0.1:8090/api/leads/imports" -Body '{"tenant_id":"gate-tenant","organization_id":"gate-org","team_id":"gate-team","channel_id":"gate-channel"}'
    if (-not $batch.success) { throw "lead import batch probe failed" }
  }

  Run-Step -Name "BI 与报表探针" -Script {
    $overview = Invoke-Json -Uri "http://127.0.0.1:8091/api/analytics/group-overview"
    $report = Invoke-Json -Method "POST" -Uri "http://127.0.0.1:8091/api/reports" -Body '{"tenant_id":"demo-tenant","report_type":"standard_summary","format":"xlsx","created_by":"release-gate"}'
    if (-not $overview.success -or $report.data.consistency.status -ne "passed") { throw "bi/report probe failed" }
  }

  Run-Step -Name "虚拟账户清理探针" -Script {
    $account = Invoke-Json -Method "POST" -Uri "http://127.0.0.1:8092/api/virtual-accounts" -Body '{"tenant_id":"gate-tenant","owner_user_id":"gate-user","display_name":"release gate","purpose":"release validation"}'
    $destroy = Invoke-Json -Method "POST" -Uri ("http://127.0.0.1:8092/api/virtual-accounts/" + $account.data.id + "/destroy") -Body '{"actor":"release-gate","reason":"manual_destroy"}'
    if (-not $destroy.data.account.cleanup_verified) { throw "virtual account cleanup probe failed" }
  }
}

Write-Host "Release gate passed."
