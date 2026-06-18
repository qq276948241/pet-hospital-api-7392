# 测试预约提醒功能
$baseUrl = "http://localhost:8080"

Write-Host "=== 步骤1: 注册用户 ===" -ForegroundColor Green
$regBody = @{username="reminder_test";password="123456";phone="13900139000";email="test@pet.com"} | ConvertTo-Json
$regResp = Invoke-RestMethod -Uri "$baseUrl/api/auth/register" -Method POST -Body $regBody -ContentType "application/json"
$regResp | ConvertTo-Json -Depth 10
$token = $regResp.data.token
Write-Host "Token: $token" -ForegroundColor Yellow

$headers = @{Authorization="Bearer $token"}

Write-Host "`n=== 步骤2: 创建宠物档案 ===" -ForegroundColor Green
$petBody = @{name="旺财";species="狗";breed="金毛";gender="公";age=3;weight=25.5;color="金色"} | ConvertTo-Json
$petResp = Invoke-RestMethod -Uri "$baseUrl/api/pets" -Method POST -Body $petBody -Headers $headers -ContentType "application/json"
$petResp | ConvertTo-Json -Depth 10
$petId = $petResp.data.id
Write-Host "宠物ID: $petId" -ForegroundColor Yellow

Write-Host "`n=== 步骤3: 获取医生列表 ===" -ForegroundColor Green
$doctorsResp = Invoke-RestMethod -Uri "$baseUrl/api/doctors" -Method GET
$doctorId = $doctorsResp.data[0].id
Write-Host "选择医生ID: $doctorId - $($doctorsResp.data[0].name)" -ForegroundColor Yellow

Write-Host "`n=== 步骤4: 创建预约（明天上午）===" -ForegroundColor Green
$tomorrow = (Get-Date).AddDays(1).ToString("yyyy-MM-dd")
Write-Host "预约日期: $tomorrow" -ForegroundColor Yellow
$apptBody = @{
    pet_id = $petId
    doctor_id = $doctorId
    appointment_date = $tomorrow
    shift = "上午"
    symptoms = "食欲不振，精神萎靡"
} | ConvertTo-Json
$apptResp = Invoke-RestMethod -Uri "$baseUrl/api/appointments" -Method POST -Body $apptBody -Headers $headers -ContentType "application/json"
$apptResp | ConvertTo-Json -Depth 10
$apptId = $apptResp.data.appointment.id

Write-Host "`n=== 步骤5: 获取通知列表（查看提醒是否已创建）===" -ForegroundColor Green
Start-Sleep -Seconds 2
$notifResp = Invoke-RestMethod -Uri "$baseUrl/api/notifications" -Method GET -Headers $headers
$notifResp | ConvertTo-Json -Depth 10
Write-Host "未读通知数量: $($notifResp.data.unread_count)" -ForegroundColor Yellow

Write-Host "`n=== 步骤6: 创建测试提醒（2分钟后发送）===" -ForegroundColor Green
$testReminderBody = @{appointment_id=$apptId; minutes_before=2} | ConvertTo-Json
$testResp = Invoke-RestMethod -Uri "$baseUrl/api/notifications/test-reminder" -Method POST -Body $testReminderBody -Headers $headers -ContentType "application/json"
$testResp | ConvertTo-Json -Depth 10

Write-Host "`n=== 步骤7: 获取未读通知数量 ===" -ForegroundColor Green
Start-Sleep -Seconds 3
$countResp = Invoke-RestMethod -Uri "$baseUrl/api/notifications/unread-count" -Method GET -Headers $headers
$countResp | ConvertTo-Json -Depth 10

Write-Host "`n=== 等待35秒，让调度器触发提醒发送... ===" -ForegroundColor Yellow
for($i=35; $i -ge 1; $i--) {
    Write-Host "`r倒计时: $i 秒" -NoNewline -ForegroundColor Cyan
    Start-Sleep -Seconds 1
}
Write-Host ""

Write-Host "`n=== 步骤8: 再次获取通知（查看提醒是否已发送）===" -ForegroundColor Green
$notifResp2 = Invoke-RestMethod -Uri "$baseUrl/api/notifications" -Method GET -Headers $headers
foreach($notif in $notifResp2.data.list) {
    Write-Host "通知ID: $($notif.id)" -ForegroundColor Cyan
    Write-Host "  标题: $($notif.title)"
    Write-Host "  内容: $($notif.content)"
    Write-Host "  状态: 已发送=$($notif.is_sent), 已读=$($notif.is_read)"
    Write-Host "  计划时间: $($notif.scheduled_time)"
    if($notif.sent_time) { Write-Host "  发送时间: $($notif.sent_time)" }
    Write-Host ""
}

Write-Host "`n=== 步骤9: 标记所有通知为已读 ===" -ForegroundColor Green
$readResp = Invoke-RestMethod -Uri "$baseUrl/api/notifications/read-all" -Method POST -Headers $headers
$readResp | ConvertTo-Json -Depth 10

Write-Host "`n=== 步骤10: 取消预约 ===" -ForegroundColor Green
$cancelBody = @{cancel_reason="临时有事"} | ConvertTo-Json
$cancelResp = Invoke-RestMethod -Uri "$baseUrl/api/appointments/$apptId/cancel" -Method POST -Body $cancelBody -Headers $headers -ContentType "application/json"
$cancelResp | ConvertTo-Json -Depth 10

Write-Host "`n=== 测试完成 ===" -ForegroundColor Green
Write-Host "请查看服务器日志，确认 [通知] 输出" -ForegroundColor Yellow
