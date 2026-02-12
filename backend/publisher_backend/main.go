package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/google/uuid"
)

const (
	SecretAuthToken = "2790c37723d14f8c9964d368e2203325"
	CampaignAPIURL  = "https://external-service.com/api/campaigns" // Your fallback API
	SDKHash         = "afd8d01eb160a5e534a659b7c30a7cc0"
	InspectToken    = "823acee21e47d7876dd1b5a51caf00ef288e62b1947120b715b3b4640cf58efb"
)

var (
	tokenBalance   = 1000
	mu             sync.Mutex
	externalUserID string
)

type CampaignDetailsResponse struct {
	Campaigns []CampaignDetail `json:"Campaigns"`
}

type CampaignDetail struct {
	UUID                   string            `json:"UUID"`
	Type                   string            `json:"Type"`
	Description            string            `json:"Description"`
	PostInstallRewardCoins int               `json:"PostInstallRewardCoins"`
	App                    AppInfo           `json:"App"`
	ImageURLs              CreativeURLs      `json:"ImageURLs"`
	EventConfigs           EventConfigs      `json:"EventConfigs"`
	CashbackConfig         CashbackSDKConfig `json:"CashbackSDKConfig"`
	Promotion              *Promotion        `json:",omitempty"`
}
type Promotion struct {
	Name        string     `json:",omitempty"`
	Description string     `json:",omitempty"`
	BoostFactor float32    `json:",omitempty"`
	StartAt     *time.Time `json:",omitempty"`
	StopAt      *time.Time `json:",omitempty"`
}

type CashbackSDKConfig struct {
	IsEnabled              bool    `json:"IsEnabled"`
	ExchangeRate           float64 `json:"ExchangeRate"`
	MaxLimitPerCampaignUSD float64 `json:"MaxLimitPerCampaignUSD"`
}
type AppInfo struct {
	ID       string `json:"ID"`
	Name     string `json:"Name"`
	BundleID string `json:"BundleID"`
	Category string `json:"Category"`
}

type CreativeURLs struct {
	Portrait  string `json:"Portrait"`
	Landscape string `json:"Landscape"`
	Icon      string `json:"Icon"`
}

type EventConfigs struct {
	AdvancePlus struct {
		SequentialEvents struct {
			TotalCoinsPossible int `json:"TotalCoinsPossible"`
			Events             []struct {
				Name        string `json:"Name"`
				Description string `json:"Description"`
				Coins       int    `json:"Coins"`
			} `json:"Events"`
		} `json:"SequentialEvents"`
	} `json:"AdvancePlus"`
}
type PayoutRequest struct {
	Tokens int `json:"cost"`
}

type TokenResponse struct {
	Tokens int `json:"tokens"`
}
type InitRequest struct {
	ExternalUserID          string    `json:"ExternalUserID"`
	ProvidedDoB             string    `json:"ProvidedDoB"`
	ProvidedGender          string    `json:"ProvidedGender"`
	UserAgent               string    `json:"UserAgent"`
	ClientIP                string    `json:"ClientIP"`
	DeviceID                string    `json:"DeviceID"`
	Placement               string    `json:"Placement"`
	UANetwork               string    `json:"UANetwork"`
	UAChannel               string    `json:"UAChannel"`
	UASubPublisherEncrypted string    `json:"UASubPublisherEncrypted"`
	TOSAccepted             bool      `json:"TOSAccepted"`
	Extension               Extension `json:"Extension"`
}
type InitResponse struct {
	ExternalUserID          string    `json:"ExternalUserID"`
	AppHash                 string    `json:"AppHash"`
	UserUUID                string    `json:"UserUUID"`
	ProvidedGender          string    `json:"ProvidedGender"`
	ProvidedDoB             string    `json:"ProvidedDoB"`
	UserAgent               string    `json:"UserAgent"`
	ClientIP                string    `json:"ClientIP"`
	DeviceID                string    `json:"DeviceID"`
	Placement               string    `json:"Placement"`
	UANetwork               string    `json:"UANetwork"`
	UAChannel               string    `json:"UAChannel"`
	UASubPublisherEncrypted string    `json:"UASubPublisherEncrypted"`
	TOSAccepted             bool      `json:"TOSAccepted"`
	DeviceName              string    `json:"DeviceName"`
	Extension               Extension `json:"Extension"`
	IsRelayEnabled          bool      `json:"IsRelayEnabled"`
	ATTLimit                int       `json:"ATTLimit"`
	ATTShown                int       `json:"ATTShown"`
}
type Extension struct {
	SubID1 string `json:"SubID1"`
	SubID2 string `json:"SubID2"`
	SubID3 string `json:"SubID3"`
	SubID4 string `json:"SubID4"`
	SubID5 string `json:"SubID5"`
}
type Offer struct {
	AppID          string            `json:"AppID"`
	AppName        string            `json:"AppName"`
	Coins          int               `json:"Coins"`
	Token          string            `json:"Token"`
	IsRecommended  bool              `json:"IsRecommended"`
	ImageURLs      map[string]string `json:"ImageURLs"`
	VideoURLs      map[string]string `json:"VideoURLs"`
	App            AppInfo           `json:"App"`
	Description    string            `json:"Description"`
	EventConfigs   EventConfigs      `json:"EventConfigs"`
	CashbackConfig CashbackSDKConfig `json:"CashbackSDKConfig"`
	Promotion      *Promotion        `json:",omitempty"`
}

type OffersResponse struct {
	Offers []*Offer `json:"Offers"`
}

var (
	initOnce       sync.Once
	globalInitData *InitResponse
	initErr        error
)

func getGlobalUser() (*InitResponse, error) {
	// This code block is guaranteed to run only once
	initOnce.Do(func() {
		fmt.Println("Performing one-time system initialization...")
		globalInitData, initErr = initializeUser()
	})
	return globalInitData, initErr
}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received request for token balance check")
	initData, err := getGlobalUser()
	if err != nil {
		http.Error(w, "System Initialization Failed", 500)
		return
	}
	// 1. Auth Check
	if r.Header.Get("Authorization") != SecretAuthToken {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	cost, err := strconv.Atoi(r.URL.Query().Get("cost"))
	if err != nil {
		http.Error(w, "Wrong format for cost", http.StatusBadRequest)
		return
	}
	if cost == 0 {
		http.Error(w, "Cost parameter is required", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// 3. Check if balance is sufficient
	if tokenBalance >= cost && tokenBalance-cost >= 0 {
		tokenBalance -= cost
		fmt.Printf("Processed request. Cost: %d, Remaining: %d\n", cost, tokenBalance)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenResponse{Tokens: tokenBalance})
		return
	}

	fmt.Println("Insufficient balance.")
	fmt.Println("Fetching campaigns.")
	offers, err := fetchOffers(initData)
	if err != nil {
		fmt.Println("Failed fetching offers" + err.Error())
		return
	}
	processCampaignDetails(offers, initData.AppHash, initData.UserUUID)
	// Example: Just returning the offers found
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(offers)
}

func fetchOffers(initData *InitResponse) (*OffersResponse, error) {
	// 1. Build the base URL
	baseURL := fmt.Sprintf("https://sb2.mainsb2.com:443/v1/studio-sdk/user/%s/offers",
		initData.ExternalUserID,
	)

	// 2. Add the Query Parameters from your curl
	params := url.Values{}
	params.Add("ignore_constraints", "CountryMatchConstraint,SDKAdvancePlusSupportConstraint,PlatformConstraint,StudioSDKNonS2SConstraint,IosFraudConstraints")
	params.Add("inspect_country", "DE")
	params.Add("usage_access_allowed", "false")

	fullURL := baseURL + "?" + params.Encode()
	// 3. Create Request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	// 4. Headers
	req.Header.Set("Adjoe-SDKHash", initData.AppHash)
	req.Header.Set("x-api-key", SecretAuthToken)
	req.Header.Set("Adjoe-IntegrationType", "studio")
	req.Header.Set("X-Inspect-Token", InspectToken)
	req.Header.Set("User-Agent", "curl/8.1.2")
	req.Header.Set("Accept", "*/*")
	// 5. Execute
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// ... after client.Do(req) ...
	bodyBytes, _ := io.ReadAll(resp.Body)

	// Re-open the body for the JSON decoder since we just read it
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	defer resp.Body.Close()

	// 6. Decode
	var result OffersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
func processCampaignDetails(offers *OffersResponse, sdkHash, userUUID string) {
	for _, offer := range offers.Offers {
		// Construct URL with the Token from the offer
		url := fmt.Sprintf("https://sb2.mainsb2.com/v1/studio-sdk/sdk/%s/tokens/%s/language/en/campaign-details",
			sdkHash, offer.Token)

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Adjoe-UserUUID", userUUID)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("Error fetching details for %s: %v\n", offer.AppName, err)
			continue
		}
		defer resp.Body.Close()

		var detailResp CampaignDetailsResponse
		if err := json.NewDecoder(resp.Body).Decode(&detailResp); err != nil {
			fmt.Println("Decode error:", err)
			continue
		}

		// Extract the info you need (assuming 1 campaign per token)
		if len(detailResp.Campaigns) > 0 {
			c := detailResp.Campaigns[0]
			offer.App = c.App
			offer.CashbackConfig = c.CashbackConfig
			offer.Description = c.Description
			offer.EventConfigs = c.EventConfigs
			offer.Promotion = c.Promotion
		}
	}
}
func initializeUser() (*InitResponse, error) {
	initURL := "https://sb2.mainsb2.com:443/v2/user-management/public/app/" + SDKHash + "/init"
	externalUserID = uuid.NewString()
	payload := InitRequest{
		ExternalUserID:          externalUserID,
		ProvidedDoB:             "1992-12-21T18:21:25.000Z",
		ProvidedGender:          "female",
		UserAgent:               "Mozilla/5.0 (iPhone; CPU iPhone OS 16_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.4 Mobile/15E148 Safari/604.1",
		ClientIP:                "62.224.63.244",
		DeviceID:                "f365c07d-c0e0-4cb7-a540-62e467d63d4b",
		Placement:               "home_screen_1",
		UANetwork:               "ironsource1",
		UAChannel:               "video1",
		UASubPublisherEncrypted: "58e468e6f77c2372f0a7891a6254bd4851b7df7b1",
		TOSAccepted:             true,
		Extension: Extension{
			SubID1: "Ident1",
			SubID2: "Ident2",
			SubID3: "Ident3",
			SubID4: "Ident4",
			SubID5: "Ident5",
		},
	}

	// Convert struct to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Create the request
	req, err := http.NewRequest("POST", initURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var initResp InitResponse
	if err := json.NewDecoder(resp.Body).Decode(&initResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	time.Sleep(2 * time.Second)
	fmt.Printf("Successfully initialized User: %s\n", initResp.UserUUID)
	return &initResp, nil
}

func payoutHandler(w http.ResponseWriter, r *http.Request) {
	payoutRequest := &PayoutRequest{}
	err := json.NewDecoder(r.Body).Decode(payoutRequest)
	if err != nil {

	}
}

func main() {
	http.HandleFunc("/check-balance", tokenHandler)
	http.HandleFunc("/s2s-payout", payoutHandler)

	fmt.Println("Server starting on :8080...")
	http.ListenAndServe(":8080", nil)
}
