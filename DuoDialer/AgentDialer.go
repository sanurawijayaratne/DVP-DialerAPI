package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/fatih/color"
)

//Add preview dial request to dialer
func AddAgentDialRequest(company, tenant int, resourceServer ResourceServerInfo, campaignId, scheduleId, campaignName, dialoutMec, uuid, fromNumber, trunkCode, phoneNumber, xGateway, numExtraData, tryCount, extention string, integrationData *IntegrationConfig, contacts *[]Contact, thirdpartyreference, businessUnit string, numberWiseSkills []string) {
	fmt.Println("Start AddPreviewDialRequest: ", uuid, ": ", fromNumber, ": ", trunkCode, ": ", phoneNumber, ": ", extention, ": ", xGateway)

	strTenant := strconv.Itoa(tenant)
	strCompany := strconv.Itoa(company)
	AddCampaignCallsRealtime(phoneNumber, tryCount, "WAITING", strTenant, strCompany, campaignId, uuid)

	IncrConcurrentChannelCount(resourceServer.ResourceServerId, campaignId)

	IncrCampaignDialCount(company, tenant, campaignId)
	InitiateSessionInfo(company, tenant, 240, "Campaign", "Dialer", "AgentDial", tryCount, campaignId, scheduleId, campaignName, uuid, phoneNumber, "ards added", "dial_start", time.Now().UTC().Format(layout4), resourceServer.ResourceServerId, integrationData, contacts, "", thirdpartyreference, businessUnit)
	InitiateAgentSessionInfo(company, tenant, 240, campaignId, campaignName, uuid, phoneNumber, integrationData, thirdpartyreference)
	SetSessionInfo(campaignId, uuid, "FromNumber", fromNumber)
	SetSessionInfo(campaignId, uuid, "TrunkCode", trunkCode)
	SetSessionInfo(campaignId, uuid, "Extention", extention)
	SetSessionInfo(campaignId, uuid, "XGateway", xGateway)

	//get attribute info from redis ** after put data stucture to cam service
	attributeInfo := make([]string, 0)

	if numberWiseSkills != nil && len(numberWiseSkills) > 0 {
		attributeInfo = numberWiseSkills
	} else {
		attributeInfo = RequestCampaignAttributeInfo(company, tenant, campaignId)
	}

	reqOtherData := RequestOtherData{}
	reqOtherData.CampaignId = campaignId
	reqOtherData.StrData = numExtraData
	reqOtherData.DialoutMec = dialoutMec
	tmpReqOtherData, _ := json.Marshal(reqOtherData)
	tmpAttributeInfo, _ := json.Marshal(attributeInfo)
	SetSessionInfo(campaignId, uuid, "Skills", string(tmpAttributeInfo))

	resp, err := AddRequest(company, tenant, uuid, string(tmpReqOtherData), attributeInfo)
	if err != nil {
		SetSessionInfo(campaignId, uuid, "Reason", "ards_failed")
		SetSessionInfo(campaignId, uuid, "DialerStatus", "failed")
		go UploadSessionInfo(campaignId, uuid)
		fmt.Println(err.Error())
	}

	if resp != "" {
		var ardsRes = ArdsResult{}
		json.Unmarshal([]byte(resp), &ardsRes)
		if ardsRes.IsSuccess == false {
			SetSessionInfo(campaignId, uuid, "Reason", ardsRes.CustomMessage)
			SetSessionInfo(campaignId, uuid, "DialerStatus", "failed")
			go UploadSessionInfo(campaignId, uuid)
		}
	}
}

//Once agent accept, send dial data to server
func DialAgent(contactName, domain, contactType, resourceId, company, tenant, campaignId, ardsServerType, ardsReqType, sessionId string) {
	sessionInfoKey := fmt.Sprintf("sessionInfo:%s:%s", campaignId, sessionId)

	agentPrepTimeInt, _ := strconv.ParseInt(agentPrepareTime, 10, 64)
	if agentPrepTimeInt > 0 {
		color.Yellow("===========SLEEEPING============")
		time.Sleep(time.Duration(agentPrepTimeInt) * time.Second)
		color.Yellow("===========SLEEEP ENDED============")
	}
	//

	if RedisCheckKeyExist(sessionInfoKey) {
		sessionInfo := RedisHashGetAll(sessionInfoKey)
		fromNumber := sessionInfo["FromNumber"]
		trunkCode := sessionInfo["TrunkCode"]
		phoneNumber := sessionInfo["Number"]
		extention := sessionInfo["Extention"]
		callServerId := sessionInfo["ServerId"]
		xGateway := sessionInfo["XGateway"]
		campaignName := sessionInfo["CampaignName"]
		ardsQueueName := sessionInfo["ArdsQueueName"]

		color.Yellow(fmt.Sprintf("SESSION DATA : %v", sessionInfo))

		companyInt, _ := strconv.Atoi(company)
		tenantInt, _ := strconv.Atoi(tenant)
		resourceServer := GetResourceServerInfo(companyInt, tenantInt, callServerId, ardsReqType)

		fmt.Println("Start DialPreviewNumber: ", campaignName, ": ", sessionId, ": ", fromNumber, ": ", trunkCode, ": ", phoneNumber, ": ", extention)
		customCompanyStr := fmt.Sprintf("%s_%s", company, tenant)

		if fromNumber != "" && trunkCode != "" && phoneNumber != "" && extention != "" {
			/*
				param := fmt.Sprintf(" {sip_h_DVP-DESTINATION-TYPE=GATEWAY,DVP_CUSTOM_PUBID=%s,CampaignId=%s,CustomCompanyStr=%s,OperationType=Dialer,return_ring_ready=true,ignore_early_media=false,origination_uuid=%s,origination_caller_id_number=%s,originate_timeout=30}", subChannelName, campaignId, customCompanyStr, sessionId, fromNumber)
				furl := fmt.Sprintf("sofia/gateway/%s/%s ", trunkCode, phoneNumber)
				var data string
				var dial bool
				if contactType == "PRIVATE" {
					dial = true
					data = fmt.Sprintf(" &bridge({sip_h_DVP-DESTINATION-TYPE=PRIVATE_USER,ards_client_uuid=%s,ards_resource_id=%s,tenantid=%s,companyid=%s,ards_servertype=%s,ards_requesttype=%s,DVP_ACTION_CAT=DIALER}user/%s@%s)", sessionId, resourceId, tenant, company, ardsServerType, ardsReqType, contactName, domain)
				} else if contactType == "PUBLIC" {
					dial = true
					data = fmt.Sprintf(" &bridge({sip_h_DVP-DESTINATION-TYPE=PUBLIC_USER,ards_client_uuid=%s,ards_resource_id=%s,tenantid=%s,companyid=%s,ards_servertype=%s,ards_requesttype=%s,DVP_ACTION_CAT=DIALER}sofia/external/%s@%s)", sessionId, resourceId, tenant, company, ardsServerType, ardsReqType, contactName, domain)
				} else if contactType == "TRUNK" {
					dial = true
					data = fmt.Sprintf(" &bridge({sip_h_DVP-DESTINATION-TYPE=GATEWAY,ards_client_uuid=%s,ards_resource_id=%s,tenantid=%s,companyid=%s,ards_servertype=%s,ards_requesttype=%s,DVP_ACTION_CAT=DIALER}sofia/gateway/%s/%s)", sessionId, resourceId, tenant, company, ardsServerType, ardsReqType, domain, contactName)
				} else {
					dial = false
					fmt.Println("Invalied ContactType")
				}
			*/

			//http://159.203.160.47:8080/api/originate?%20{sip_h_DVP-DESTINATION-TYPE=GATEWAY,DVP_CUSTOM_PUBID=dialerDialer2,CampaignId=106,CustomCompanyStr=103_1,OperationType=Dialer,return_ring_ready=true,ignore_early_media=false,origination_uuid=a63cc8f7-56f6-4dee-a907-f67a9392d56c,origination_caller_id_number=94112500500,originate_timeout=30}sofia/gateway/DemoTrunk/18705056540%20&bridge({sip_h_DVP-DESTINATION-TYPE=PRIVATE_USER,ards_client_uuid=a63cc8f7-56f6-4dee-a907-f67a9392d56c,ards_resource_id=111,tenantid=1,companyid=103,ards_servertype=DIALER,ards_requesttype=CALL,DVP_ACTION_CAT=DIALER}user/heshan@duo.media1.veery.cloud)

			//http://159.203.160.47:8080/api/originate?%20{sip_h_DVP-DESTINATION-TYPE=PRIVATE_USER,ards_client_uuid=a63cc8f7-56f6-4dee-a907-f67a9392d56c,origination_uuid=a63cc8f7-56f6-4dee-a907-f67a9392d56c,ards_resource_id=111,tenantid=1,companyid=103,ards_servertype=DIALER,ards_requesttype=CALL,DVP_ACTION_CAT=DIALER,return_ring_ready=false,ignore_early_media=true,origination_caller_id_number=18705056560}user/heshan@duo.media1.veery.cloud%20&bridge({sip_h_DVP-DESTINATION-TYPE=GATEWAY,DVP_CUSTOM_PUBID=dialerDialer2,CampaignId=106,CustomCompanyStr=103_1,OperationType=Dialer,origination_caller_id_number=94112500500,originate_timeout=30}sofia/gateway/DemoTrunk/18705056560)
			//http://159.203.160.47:8080/api/originate?%20{sip_h_DVP-DESTINATION-TYPE=PRIVATE_USER,ards_client_uuid=3ce432e7-9c3a-4c9c-b432-10951883ad60,origination_uuid=3ce432e7-9c3a-4c9c-b432-10951883ad60,ards_resource_id=111,tenantid=1,companyid=103,ards_servertype=DIALER,ards_requesttype=CALL,DVP_ACTION_CAT=DIALER,return_ring_ready=false,ignore_early_media=true,origination_caller_id_number=94777888999}user/heshan@duo.media1.veery.cloud%20&bridge({sip_h_DVP-DESTINATION-TYPE=GATEWAY,DVP_CUSTOM_PUBID=dialerDialer2,CampaignId=106,CustomCompanyStr=103_1,OperationType=Dialer,origination_caller_id_number=94777888999,originate_timeout=30}sofia/gateway/DemoTrunk/18705056540)
			var param string
			var furl string
			var data string

			param = fmt.Sprintf("{sip_h_DVP-DESTINATION-TYPE=GATEWAY,DVP_CALL_DIRECTION=outbound,ards_skill_display=%s,DVP_CUSTOM_PUBID=%s,nolocal:DIALER_AGENT_EVENT=%s,CustomCompanyStr=%s,CampaignId=%s,CampaignName='%s',tenantid=%s,companyid=%s,ards_resource_id=%s,ards_client_uuid=%s,origination_uuid=%s,ards_servertype=%s,ards_requesttype=%s,DVP_ACTION_CAT=DIALER,DVP_OPERATION_CAT=AGENT,return_ring_ready=false,ignore_early_media=true,origination_caller_id_name=%s,origination_caller_id_number=%s,DialerCustomerNumber=%s,DialerAgentName=%s,DialerAgentSipName=%s,sip_h_X-Gateway=%s,CALL_LEG_TYPE=CUSTOMER}", ardsQueueName, subChannelName, subChannelNameAgent, customCompanyStr, campaignId, campaignName, tenant, company, resourceId, sessionId, sessionId, ardsServerType, ardsReqType, fromNumber, fromNumber, phoneNumber, sessionInfo["Agent"], sessionInfo["AgentSipName"], xGateway)
			furl = fmt.Sprintf("sofia/gateway/%s/%s", trunkCode, phoneNumber)

			/* if contactType == "PRIVATE" {
				dial = true
				param = fmt.Sprintf(" {sip_h_DVP-DESTINATION-TYPE=PRIVATE_USER,DVP_CALL_DIRECTION=outbound,ards_skill_display=%s,nolocal:DVP_CUSTOM_PUBID=%s,DIALER_AGENT_EVENT=%s,CustomCompanyStr=%s,CampaignId=%s,CampaignName='%s',tenantid=%s,companyid=%s,ards_resource_id=%s,ards_client_uuid=%s,origination_uuid=%s,ards_servertype=%s,ards_requesttype=%s,DVP_ACTION_CAT=DIALER,DVP_OPERATION_CAT=AGENT,return_ring_ready=false,ignore_early_media=true,origination_caller_id_number=%s}", ardsQueueName, subChannelName, subChannelNameAgent, customCompanyStr, campaignId, campaignName, tenant, company, resourceId, sessionId, sessionId, ardsServerType, ardsReqType, phoneNumber)
				furl = fmt.Sprintf("user/%s@%s", contactName, domain)
			} else if contactType == "PUBLIC" {
				dial = true
				furl = fmt.Sprintf("sofia/external/%s@%s", contactName, domain)
			} else if contactType == "TRUNK" {
				dial = true
				furl = fmt.Sprintf("sofia/gateway/%s/%s", domain, contactName)
			} else {
				dial = false
				fmt.Println("Invalied ContactType")
			} */

			//			if xGateway != "" {

			//				data = fmt.Sprintf(" &bridge({sip_h_DVP-DESTINATION-TYPE=GATEWAY,DVP_CUSTOM_PUBID=%s,tenantid=%s,companyid=%s,ards_client_uuid=%s,CampaignId=%s,CampaignName='%s',CustomCompanyStr=%s,OperationType=Dialer,origination_caller_id_number=%s,DVP_OPERATION_CAT=CUSTOMER,originate_timeout=30,sip_h_X-Gateway=%s,ignore_early_media=false}sofia/gateway/%s/%s)", subChannelName, tenant, company, sessionId, campaignId, campaignName, customCompanyStr, fromNumber, xGateway, trunkCode, phoneNumber)
			//			} else {
			//				data = fmt.Sprintf(" &bridge({sip_h_DVP-DESTINATION-TYPE=GATEWAY,DVP_CUSTOM_PUBID=%s,tenantid=%s,companyid=%s,ards_client_uuid=%s,CampaignId=%s,CampaignName='%s',CustomCompanyStr=%s,OperationType=Dialer,origination_caller_id_number=%s,DVP_OPERATION_CAT=CUSTOMER,originate_timeout=30}sofia/gateway/%s/%s)", subChannelName, tenant, company, sessionId, campaignId, campaignName, customCompanyStr, fromNumber, trunkCode, phoneNumber)
			//			}

			//call recording enable
			data = fmt.Sprintf(" %s xml %d_%d_dialer", contactName, tenantInt, companyInt)
			//data = fmt.Sprintf(" %s xml dialer", phoneNumber)

			SetSessionInfo(campaignId, sessionId, "Reason", "Dial Number")
			redwhite := color.New(color.FgRed).Add(color.BgWhite)
			redwhite.Println(fmt.Sprintf("DIALING OUT CALL - AGENT CAMPAIGN : %s | NUMBER : %s", campaignName, phoneNumber))

			//resp, err := DialNew(resourceServer.Url, param, furl, data)
			RemoveRequest(company, tenant, sessionId)
			//PublishCampaignCallCounts(sessionId, "DIALED", company, tenant, campaignId)
			SetSessionInfo(campaignId, sessionId, "IsDialed", "TRUE")

			dashboardparam2 := "BASIC"
			tryCountInt, _ := strconv.Atoi(sessionInfo["TryCount"])
			if tryCountInt > 1 {
				dashboardparam2 = "CALLBACK"
			}
			PublishCampaignCallCounts(sessionId, "DIALING", company, tenant, campaignId, dashboardparam2)
			UpdateCampaignCallRealtimeField("DialState", "DIALING", tenant, company, campaignId, sessionId)
			resp, err := Dial(resourceServer.Url, param, furl, data)
			HandleDialResponse(resp, err, resourceServer, campaignId, sessionId)
		} else {
			color.Red(fmt.Sprintf("=====Redis Hash From Numbers Not Found - Dial Failed : %s", sessionInfoKey))
			RemoveRequestNoSession(company, tenant, sessionId)
			SetAgentStatusArds(company, tenant, "", resourceId, sessionId, "Completed", ardsServerType, ardsReqType)
			AbortDialing(company, tenant, campaignId, sessionId, "NoSession")
		}

	} else {
		color.Red(fmt.Sprintf("=====Redis Hash Not Found For Key - Dial Failed : %s", sessionInfoKey))
		RemoveRequestNoSession(company, tenant, sessionId)
		SetAgentStatusArds(company, tenant, "", resourceId, sessionId, "Completed", ardsServerType, ardsReqType)
		AbortDialing(company, tenant, campaignId, sessionId, "NoSession")
	}
}

func AgentRejectWithoutCallRemove(company, tenant, campaignId, sessionId, requestType, resourceId, rejectReason string) {
	sessionInfoKey := fmt.Sprintf("sessionInfo:%s:%s", campaignId, sessionId)
	if RedisCheckKeyExist(sessionInfoKey) {
		SetSessionInfo(campaignId, sessionId, "Reason", rejectReason)
		SetSessionInfo(campaignId, sessionId, "DialerStatus", "agent_reject")
		ClearResourceSlotWhenReject(company, tenant, requestType, resourceId, sessionId)
		//RejectRequest(company, tenant, sessionId)
		//go UploadSessionInfo(campaignId, sessionId)
	}
}

func AgentReject(company, tenant, campaignId, sessionId, requestType, resourceId, rejectReason string) {
	sessionInfoKey := fmt.Sprintf("sessionInfo:%s:%s", campaignId, sessionId)
	if RedisCheckKeyExist(sessionInfoKey) {
		SetSessionInfo(campaignId, sessionId, "Reason", rejectReason)
		SetSessionInfo(campaignId, sessionId, "DialerStatus", "agent_reject")
		ClearResourceSlotWhenReject(company, tenant, requestType, resourceId, sessionId)
		//RejectRequest(company, tenant, sessionId)
		go UploadSessionInfo(campaignId, sessionId)
	}
}

func AbortDialing(company, tenant, campaignId, sessionId, rejectReason string) {
	sessionInfoKey := fmt.Sprintf("sessionInfo:%s:%s", campaignId, sessionId)
	if RedisCheckKeyExist(sessionInfoKey) {
		SetSessionInfo(campaignId, sessionId, "ReasonCode", rejectReason)
		SetSessionInfo(campaignId, sessionId, "Reason", rejectReason)
		SetSessionInfo(campaignId, sessionId, "DialerStatus", "abort_dialing")
		//RejectRequest(company, tenant, sessionId)
		go UploadSessionInfo(campaignId, sessionId)
	}
}

//-----------------------------------CampaignManager Service-------------------------------------------------
func RequestCampaignAttributeInfo(company, tenant int, campaignId string) []string {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in RequestCampaignAttributeInfo", r)
		}
	}()
	//Request campaign from Campaign Manager service
	attributeDetails := make([]string, 0)

	response := RequestCampaignAddtionalData(company, tenant, campaignId, "PREVIEW", "ARDS", "ATTRIBUTE")
	if len(response) > 0 && response[0] != "" {
		var attInfo []string
		json.Unmarshal([]byte(response[0]), &attInfo)
		attributeDetails = attInfo
	}
	return attributeDetails
}

func RequestCampaignAddtionalData(company, tenant int, campaignId, class, ctype, category string) []string {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in RequestCampaignAddtionalData", r)
		}
	}()

	additionalData := make([]string, 0)
	//Request campaign from Campaign Manager service
	jwtToken := fmt.Sprintf("Bearer %s", accessToken)
	internalAuthToken := fmt.Sprintf("%d:%d", tenant, company)

	client := &http.Client{}

	request := fmt.Sprintf("http://%s/DVP/API/1.0.0.0/CampaignManager/Campaign/%s/AdditionalData/%s/%s/%s", CreateHost(campaignServiceHost, campaignServicePort), campaignId, class, ctype, category)
	fmt.Println("Start RequestCampaign request: ", request)
	req, _ := http.NewRequest("GET", request, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", jwtToken)
	req.Header.Set("companyinfo", internalAuthToken)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return additionalData
	}
	defer resp.Body.Close()

	response, _ := ioutil.ReadAll(resp.Body)

	if response != nil {
		var campaignAdditionalDataResult CampaignAdditionalDataResult
		json.Unmarshal(response, &campaignAdditionalDataResult)
		if campaignAdditionalDataResult.IsSuccess == true {

			for _, data := range campaignAdditionalDataResult.Result {
				additionalData = append(additionalData, data.AdditionalData)
			}
			return additionalData
		} else {
			return additionalData
		}
	} else {
		return additionalData
	}
}
