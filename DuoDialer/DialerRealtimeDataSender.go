package main

import (
	"strconv"
	"fmt"
	"encoding/json"
	"github.com/fatih/color"
)


func AddCampaignDataRealtime(campaignData Campaign) {
	color.Cyan(fmt.Sprintf("Adding Campaign Realtime Data"))
	campInfoRealTime := make(map[string]string)

	campInfoRealTime["CampaignId"] = strconv.Itoa(campaignData.CampaignId)
	campInfoRealTime["CampaignName"] = campaignData.CampaignName
	campInfoRealTime["StartTime"] = campaignData.CampConfigurations.StartDate.Format("02 Jan 06 15:04 -0700")
	campInfoRealTime["EndTime"] = campaignData.CampConfigurations.EndDate.Format("02 Jan 06 15:04 -0700")
	campInfoRealTime["CampaignMode"] = campaignData.CampaignMode
	campInfoRealTime["CampaignChannel"] = campaignData.CampaignChannel
	campInfoRealTime["DialoutMechanism"] = campaignData.DialoutMechanism
	campInfoRealTime["Extension"] = campaignData.Extensions
	campInfoRealTime["OperationalStatus"] = "WAITING"

	key := fmt.Sprintf("RealTimeCampaign:%d:%d:%d", campaignData.TenantId, campaignData.CompanyId, campaignData.CampaignId)

	RedisHMSet(key, campInfoRealTime)

	campData, _ := json.Marshal(campInfoRealTime)
	campDataStr := string(campData)

	go SendNotificationToRoom("DIALER:RealTimeCampaignEvents", "DIALER", "STATELESS", campDataStr, "NEW_CAMPAIGN", campaignData.CompanyId, campaignData.TenantId)
	
}

func AddCampaignCallsRealtime(PhoneNumber, TryCount, DialState, TenantId, CompanyId, CampaignId, SessionId string) {
	color.Cyan(fmt.Sprintf("Adding Campaign Call Realtime Data"))
	campCallRealTime := make(map[string]string)

	campCallRealTime["PhoneNumber"] = PhoneNumber
	campCallRealTime["TryCount"] = TryCount
	campCallRealTime["DialState"] = DialState
	campCallRealTime["TenantId"] = TenantId
	campCallRealTime["CompanyId"] = CompanyId
	campCallRealTime["CampaignId"] = CampaignId
	campCallRealTime["SessionId"] = SessionId

	key := fmt.Sprintf("RealTimeCampaignCalls:%s:%s:%s:%s", TenantId, CompanyId, CampaignId, SessionId)

	RedisHMSet(key, campCallRealTime)

	campData, _ := json.Marshal(campCallRealTime)
	campDataStr := string(campData)

	companyIdInt, _ := strconv.Atoi(CompanyId)
	tenantIdInt, _ := strconv.Atoi(TenantId)

	go SendNotificationToRoom("DIALER:RealTimeCampaignEvents", "DIALER", "STATELESS", campDataStr, "NEW_CAMPAIGN_CALL", companyIdInt, tenantIdInt)
	
}

func UpdateCampaignRealtimeField(fieldName, val string, tenantId, companyId, campaignId int) {
	color.Cyan(fmt.Sprintf("Updating Campaign Realtime Field"))

	key := fmt.Sprintf("RealTimeCampaign:%d:%d:%d", tenantId, companyId, campaignId)

	RedisHashSetField(key, fieldName, val)

	campInfoRealTime := make(map[string]string)

	campInfoRealTime[fieldName] = val
	campInfoRealTime["CampaignId"] = strconv.Itoa(campaignId)

	campData, _ := json.Marshal(campInfoRealTime)
	campDataStr := string(campData)

	go SendNotificationToRoom("DIALER:RealTimeCampaignEvents", "DIALER", "STATELESS", campDataStr, "UPDATE_CAMPAIGN", companyId, tenantId)
}

func UpdateCampaignCallRealtimeField(fieldName, val, tenantId, companyId, campaignId, sessionId string) {
	color.Cyan(fmt.Sprintf("Updating Campaign Realtime Field"))

	key := fmt.Sprintf("RealTimeCampaignCalls:%s:%s:%s:%s", tenantId, companyId, campaignId, sessionId)

	RedisHashSetField(key, fieldName, val)

	campCallInfoRealTime := make(map[string]string)

	campCallInfoRealTime[fieldName] = val
	campCallInfoRealTime["SessionId"] = sessionId

	campData, _ := json.Marshal(campCallInfoRealTime)
	campDataStr := string(campData)

	companyIdInt, _ := strconv.Atoi(companyId)
	tenantIdInt, _ := strconv.Atoi(tenantId)

	go SendNotificationToRoom("DIALER:RealTimeCampaignEvents", "DIALER", "STATELESS", campDataStr, "UPDATE_CAMPAIGN_CALL", companyIdInt, tenantIdInt)

	
}

func RemoveCampaignRealtime(tenantId, companyId, campaignId int) {
	color.Cyan(fmt.Sprintf("Removing Campaign Realtime"))

	key := fmt.Sprintf("RealTimeCampaign:%d:%d:%d", tenantId, companyId, campaignId)

	RedisRemove(key)

	campInfoRealTime := make(map[string]string)

	campInfoRealTime["CampaignId"] = strconv.Itoa(campaignId)

	campData, _ := json.Marshal(campInfoRealTime)
	campDataStr := string(campData)

	go SendNotificationToRoom("DIALER:RealTimeCampaignEvents", "DIALER", "STATELESS", campDataStr, "REMOVE_CAMPAIGN", companyId, tenantId)
	
}

func RemoveCampaignCallRealtime(tenantId, companyId, campaignId, sessionId string) {
	color.Cyan(fmt.Sprintf("Removing Campaign Realtime"))

	key := fmt.Sprintf("RealTimeCampaignCalls:%s:%s:%s:%s", tenantId, companyId, campaignId, sessionId)

	RedisRemove(key)

	campCallInfoRealTime := make(map[string]string)

	campCallInfoRealTime["SessionId"] = sessionId

	campData, _ := json.Marshal(campCallInfoRealTime)
	campDataStr := string(campData)

	companyIdInt, _ := strconv.Atoi(companyId)
	tenantIdInt, _ := strconv.Atoi(tenantId)

	go SendNotificationToRoom("DIALER:RealTimeCampaignEvents", "DIALER", "STATELESS", campDataStr, "REMOVE_CAMPAIGN", companyIdInt, tenantIdInt)
	
}