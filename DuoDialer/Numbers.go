package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/fatih/color"
)

func GetNumbersFromNumberBase(company, tenant, numberLimit int, campaignId, camScheduleId string) []CampaignDialNumber {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in GetNumbersFromNumberBase", r)
		}
	}()
	numbers := make([]CampaignDialNumber, 0)
	pageKey := fmt.Sprintf("PhoneNumberPage:%d:%d:%s:%s", company, tenant, campaignId, camScheduleId)

	numberOffsetToRequest := "0"

	/*if numberLimit == 500 {
		numberOffsetToRequest = RedisGet(pageKey)
	}*/
	DialerLog(fmt.Sprintf("numberOffsetToRequest: %s", numberOffsetToRequest))

	// Get phone number from campign service and append
	jwtToken := fmt.Sprintf("Bearer %s", accessToken)
	internalAuthToken := fmt.Sprintf("%d:%d", tenant, company)
	DialerLog(fmt.Sprintf("Start GetPhoneNumbers Auth: %s  CampaignId: %s  camScheduleId: %s", internalAuthToken, campaignId, camScheduleId))
	client := &http.Client{}

	request := fmt.Sprintf("http://%s/DVP/API/1.0.0.0/CampaignManager/Campaign/%s/NumbersByOffset/%s/%d/%s", CreateHost(campaignServiceHost, campaignServicePort), campaignId, camScheduleId, numberLimit, numberOffsetToRequest)
	DialerLog(fmt.Sprintf("Start GetPhoneNumbers request: ", request))
	req, _ := http.NewRequest("GET", request, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", jwtToken)
	req.Header.Set("companyinfo", internalAuthToken)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return numbers
	}
	defer resp.Body.Close()

	response, _ := ioutil.ReadAll(resp.Body)

	var phoneNumberResult PhoneNumberResult
	json.Unmarshal(response, &phoneNumberResult)
	if phoneNumberResult.IsSuccess == true {
		for _, numRes := range phoneNumberResult.Result {
			numberInfo := CampaignDialNumber{}

			numberInfo.PhoneNumber = numRes.CampContactInfo.ContactId
			numberInfo.BusinessUnit = numRes.CampContactInfo.BusinessUnit
			numberInfo.PreviewData = numRes.ExtraData
			numberInfo.TryCount = "1"

			numbers = append(numbers, numberInfo)

			/* if numRes.ExtraData != "" {
				numPrevSplit := strings.Split(numRes.CampContactInfo.ContactId, ":")
				numberWithExtraD := ""
				if len(numPrevSplit) > 1 {
					numberWithExtraD = fmt.Sprintf("%s:%s:%s", numPrevSplit[0], "1", strings.Join(numPrevSplit[1:], ":"))
				} else {
					numberWithExtraD = fmt.Sprintf("%s:%s:%s", numRes.CampContactInfo.ContactId, "1", numRes.ExtraData)
				}

				fmt.Println("======NUMBER 1 : " + numberWithExtraD)
				numbers = append(numbers, numberWithExtraD)
			} else {
				numberWithData := strings.Split(numRes.CampContactInfo.ContactId, ":")
				if len(numberWithData) > 1 {
					exData := strings.Join(numberWithData[1:], ":")
					numberAndExtraD := fmt.Sprintf("%s:%s:%s", numberWithData[0], "1", exData)
					fmt.Println("======NUMBER 2 : " + numberAndExtraD)
					numbers = append(numbers, numberAndExtraD)
				} else {
					numberWithoutExtraData := fmt.Sprintf("%s:%s:", numRes.CampContactInfo.ContactId, "1")
					fmt.Println("======NUMBER 3 : " + numberWithoutExtraData)
					numbers = append(numbers, numberWithoutExtraData)
				}
			} */
		}

		RedisIncrBy(pageKey, len(phoneNumberResult.Result))
	}
	return numbers
}

func GetContactsFromNumberBase(company, tenant, numberLimit int, campaignId, camScheduleId string) []ContactsDetails {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in GetNumbersFromNumberBase", r)
		}
	}()
	numbers := make([]ContactsDetails, 0)
	//pageKey := fmt.Sprintf("PhoneNumberPage:%d:%d:%s:%s", company, tenant, campaignId, camScheduleId)

	numberOffsetToRequest := "0"

	/* if numberLimit == 500 {
		numberOffsetToRequest = RedisGet(pageKey)
	} */
	//DialerLog(fmt.Sprintf("numberOffsetToRequest: %s", numberOffsetToRequest))

	// Get phone number from campign service and append
	jwtToken := fmt.Sprintf("Bearer %s", accessToken)
	internalAuthToken := fmt.Sprintf("%d:%d", tenant, company)
	DialerLog(fmt.Sprintf("Start GetContacts Auth: %s  CampaignId: %s  camScheduleId: %s", internalAuthToken, campaignId, camScheduleId))
	client := &http.Client{}

	request := fmt.Sprintf("http://%s/DVP/API/1.0.0.0/Campaign/%s/Contacts/%d/%s?scheduleId=%s", CreateHost(contactServiceHost, contactServicePort), campaignId, numberLimit, numberOffsetToRequest, camScheduleId)
	color.Cyan(fmt.Sprintf("Start GetContacts request: %s", request))
	req, _ := http.NewRequest("GET", request, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", jwtToken)
	req.Header.Set("companyinfo", internalAuthToken)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return numbers
	}
	defer resp.Body.Close()

	response, _ := ioutil.ReadAll(resp.Body)

	var contactNumberResult ContactsResult
	json.Unmarshal(response, &contactNumberResult)
	if contactNumberResult.IsSuccess == true {
		numbers = contactNumberResult.Result

		//RedisIncrBy(pageKey, len(contactNumberResult.Result))
	}
	return numbers
}

func SetDncNumbersFromNumberBase(company, tenant int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in SetDncNumbersFromNumberBase", r)
		}
	}()

	jwtToken := fmt.Sprintf("Bearer %s", accessToken)
	internalAuthToken := fmt.Sprintf("%d:%d", tenant, company)
	DialerLog(fmt.Sprintf("Start SetDncNumbersFromNumberBase Auth: %s", internalAuthToken))
	client := &http.Client{}

	request := fmt.Sprintf("http://%s/DVP/API/1.0.0.0/CampaignManager/Dnc", CreateHost(campaignServiceHost, campaignServicePort))
	DialerLog(fmt.Sprintf("Start GetDncPhoneNumbers request: %s", request))
	req, _ := http.NewRequest("GET", request, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", jwtToken)
	req.Header.Set("companyinfo", internalAuthToken)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()

	response, _ := ioutil.ReadAll(resp.Body)

	var dncNumberResult DncNumberResult
	json.Unmarshal(response, &dncNumberResult)
	if dncNumberResult.IsSuccess == true {
		dncNumberKey := fmt.Sprintf("DncNumber:%d:%d", tenant, company)
		RedisSetAdd(dncNumberKey, dncNumberResult.Result)
	}
}

func LoadNumbers(company, tenant, numberLimit int, campaignId, camScheduleId string) {
	listId := fmt.Sprintf("CampaignNumbers:%d:%d:%s:%s", company, tenant, campaignId, camScheduleId)
	numbers := GetNumbersFromNumberBase(company, tenant, numberLimit, campaignId, camScheduleId)

	DialerLog(fmt.Sprintf("Number count = %d", len(numbers)))
	if len(numbers) == 0 {
		numLoadingStatusKey := fmt.Sprintf("PhoneNumberLoading:%d:%d:%s:%s", company, tenant, campaignId, camScheduleId)
		RedisSet(numLoadingStatusKey, "waiting")
	} else {
		numLoadingStatusKey := fmt.Sprintf("PhoneNumberLoading:%d:%d:%s:%s", company, tenant, campaignId, camScheduleId)
		dncNumberKey := fmt.Sprintf("DncNumber:%d:%d", tenant, company)
		RedisSet(numLoadingStatusKey, "waiting")
		for _, number := range numbers {
			if !RedisSetIsMember(dncNumberKey, number.PhoneNumber) {
				numInf, err := json.Marshal(number)

				if err == nil {
					fmt.Println("Adding number to campaign: ", number.PhoneNumber)
					RedisListRpush(listId, string(numInf))

				}

			}
		}
	}
}

func LoadContacts(company, tenant, numberLimit int, campaignId, camScheduleId string) {
	listId := fmt.Sprintf("CampaignContacts:%d:%d:%s", company, tenant, campaignId)
	numbers := GetContactsFromNumberBase(company, tenant, numberLimit, campaignId, camScheduleId)

	color.Green(fmt.Sprintf("=========== LOADING CONTACTS : %s ==========", campaignId))
	DialerLog(fmt.Sprintf("Number count = %d", len(numbers)))
	if len(numbers) == 0 {
		color.Green(fmt.Sprintf("=========== NO CONTACTS FOUND : %s ==========", campaignId))
		numLoadingStatusKey := fmt.Sprintf("PhoneNumberLoading:%d:%d:%s:%s", company, tenant, campaignId, camScheduleId)
		RedisSet(numLoadingStatusKey, "waiting")
	} else {
		color.Green(fmt.Sprintf("=========== CONTACTS FOUND : %s : No : %d ==========", campaignId, len(numbers)))
		numLoadingStatusKey := fmt.Sprintf("PhoneNumberLoading:%d:%d:%s:%s", company, tenant, campaignId, camScheduleId)
		dncNumberKey := fmt.Sprintf("DncNumber:%d:%d", tenant, company)
		RedisSet(numLoadingStatusKey, "waiting")
		for _, number := range numbers {
			if !RedisSetIsMember(dncNumberKey, number.Phone) {
				fmt.Println("Adding number to campaign: ", number)
				num_detail, _ := json.Marshal(number)
				color.Green(string(num_detail))
				RedisListRpush(listId, string(num_detail))
			}
		}
	}
}

func LoadInitialNumberSet(company, tenant int, campaignId, camScheduleId string, numLoadingMethod string) {
	numLoadingStatusKey := fmt.Sprintf("PhoneNumberLoading:%d:%d:%s:%s", company, tenant, campaignId, camScheduleId)

	if numLoadingMethod == "CONTACT" {
		LoadContacts(company, tenant, 1000, campaignId, camScheduleId)
	} else {
		LoadNumbers(company, tenant, 1000, campaignId, camScheduleId)
	}

	RedisSet(numLoadingStatusKey, "waiting")
}

func CheckDuplicates(company, tenant int, campaignId, camScheduleId, number string, timeout, tryCount int) bool {
	color.Green(fmt.Sprintf("========= Checking For Duplicates - Number : %s - Timeout : %d - Trycount : %d", number, timeout, tryCount))

	if timeout > 0 && tryCount < 2 {
		duplicateNumKey := fmt.Sprintf("NumberDuplicateCheck:%d:%d:%s:%s", company, tenant, campaignId, number)

		incrVal := RedisIncr(duplicateNumKey)
		RedisExpire(duplicateNumKey, timeout)

		color.Green(fmt.Sprintf("========= Number Dial Duplicate Count - Key : %s - Count : %d", duplicateNumKey, incrVal))

		if incrVal > 1 {
			color.Red(fmt.Sprintf("========= DUPLICATE NUMBER DETECTED - Key : %s - Timeout : %d", duplicateNumKey, timeout))
			return false
		} else {
			return true
		}

	} else {
		return true
	}
}

func GetNumberToDial(company, tenant int, campaignId, camScheduleId, numLoadingMethod string) (string, string, string, string, string, []Contact, []string) {
	listId := fmt.Sprintf("CampaignNumbers:%d:%d:%s:%s", company, tenant, campaignId, camScheduleId)

	if numLoadingMethod == "CONTACT" {
		listId = fmt.Sprintf("CampaignContacts:%d:%d:%s", company, tenant, campaignId)
	}
	numLoadingStatusKey := fmt.Sprintf("PhoneNumberLoading:%d:%d:%s:%s", company, tenant, campaignId, camScheduleId)
	numberCount := RedisListLlen(listId)
	numLoadingStatus := RedisGet(numLoadingStatusKey)

	if numLoadingStatus == "waiting" {
		if numberCount < 500 {
			if numLoadingMethod == "CONTACT" {
				LoadContacts(company, tenant, 500, campaignId, camScheduleId)
			} else {
				LoadNumbers(company, tenant, 500, campaignId, camScheduleId)
			}
		}
	} else if numLoadingStatus == "" {
		LoadInitialNumberSet(company, tenant, campaignId, camScheduleId, numLoadingMethod)
	}

	color.Red(fmt.Sprintf("======= NUMBER LOADING STATUS : %s", numLoadingStatus))
	numberWithTryCount := RedisListLpop(listId)
	if numLoadingMethod == "CONTACT" {
		//Add contacts to redis here
		contactInf := ContactsDetails{}
		_ = json.Unmarshal([]byte(numberWithTryCount), &contactInf)

		color.Green("NUMBER POPPED OUT TO DIAL : " + numberWithTryCount)

		strTryCount := strconv.Itoa(contactInf.TryCount)

		return contactInf.Phone, strTryCount, contactInf.PreviewData, "", contactInf.Thirdpartyreference, contactInf.Api_Contacts, contactInf.Skills
	} else {
		//18705056550:{"A":18705056550}:1:{}
		color.Green("NUMBER POPPED OUT TO DIAL : " + numberWithTryCount)

		numInf := CampaignDialNumber{}
		_ = json.Unmarshal([]byte(numberWithTryCount), &numInf)

		return numInf.PhoneNumber, numInf.TryCount, numInf.PreviewData, numInf.BusinessUnit, "", make([]Contact, 0), make([]string, 0)

		/* numberInfos := strings.Split(numberWithTryCount, ":")
		if len(numberInfos) > 3 {
			return numberInfos[0], numberInfos[1], strings.Join(numberInfos[2:], ":"), "", make([]Contact, 0)
			//return numberInfos[0], numberInfos[2], numberInfos[1], "", make([]Contact, 0)
		} else if len(numberInfos) == 3 {
			color.Green("NUMBER POPPED OUT TO DIAL : 3")
			return numberInfos[0], numberInfos[1], numberInfos[2], "", make([]Contact, 0)
		} else if len(numberInfos) == 2 {
			color.Green("NUMBER POPPED OUT TO DIAL : 2")
			return numberInfos[0], numberInfos[1], "", "", make([]Contact, 0)
		} else {
			return "", "", "", "", make([]Contact, 0)
		} */

	}

}

func GetNumberCount(company, tenant int, campaignId, camScheduleId string) int {
	DialerLog("Start GetNumberCount")
	listId := fmt.Sprintf("CampaignNumbers:%d:%d:%s:%s", company, tenant, campaignId, camScheduleId)
	return RedisListLlen(listId)
}

func RemoveNumbers(company, tenant int, campaignId string) {
	searchKey := fmt.Sprintf("CampaignNumbers:%d:%d:%s:*", company, tenant, campaignId)
	relatedNumberList := RedisSearchKeys(searchKey)
	for _, key := range relatedNumberList {
		RedisRemove(key)
	}
}

func RemoveNumberStatusKey(company, tenant int, campaignId string) {

	searchKeyPhoneNumberPage := fmt.Sprintf("PhoneNumberPage:%d:%d:%s:*", company, tenant, campaignId)
	searchKeyPhoneNumberLoading := fmt.Sprintf("PhoneNumberLoading:%d:%d:%s:*", company, tenant, campaignId)
	relatedNumberStatusKey := RedisSearchKeys(searchKeyPhoneNumberPage)
	relatedNumberStatusKeyNumberLoading := RedisSearchKeys(searchKeyPhoneNumberLoading)

	for _, key := range relatedNumberStatusKey {
		RedisRemove(key)
	}

	for _, key := range relatedNumberStatusKeyNumberLoading {
		RedisRemove(key)
	}
}

func AddNumberToFront(company, tenant int, campaignId, camScheduleId string, numberInfo CampaignDialNumber) bool {
	listId := fmt.Sprintf("CampaignNumbers:%d:%d:%s:%s", company, tenant, campaignId, camScheduleId)

	numInf, _ := json.Marshal(numberInfo)
	return RedisListLpush(listId, string(numInf))
}

func AddContactToFront(company, tenant int, campaignId string, contact ContactsDetails) bool {
	listId := fmt.Sprintf("CampaignContacts:%d:%d:%s", company, tenant, campaignId)
	num_detail, _ := json.Marshal(contact)
	RedisListRpush(listId, string(num_detail))
	return true
}
