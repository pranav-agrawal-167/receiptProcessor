package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ReceiptsforTesting() []Receipt {
	var testReceipts [3]Receipt
	testReceipts[0] = Receipt{
		Retailer:     "ABC Supermarket",
		PurchaseDate: ReceiptPurchaseDate(time.Date(2023, time.January, 1, 15, 0, 0, 0, time.UTC)),
		PurchaseTime: ReceiptPurchaseTime(time.Date(2023, time.January, 1, 15, 0, 0, 0, time.UTC)),
		Items: []Item{
			{ShortDescription: "Item 1", Price: 9.99},
			{ShortDescription: "Item 2", Price: 5.99},
			{ShortDescription: "", Price: 3.49},
		},
		Total: 19.47,
	}
	testReceipts[1] = Receipt{
		Retailer:     "",
		PurchaseDate: ReceiptPurchaseDate(time.Date(2023, time.January, 2, 14, 0, 0, 0, time.UTC)),
		PurchaseTime: ReceiptPurchaseTime(time.Date(2023, time.January, 2, 14, 0, 0, 0, time.UTC)),
		Items: []Item{
			{ShortDescription: "Apples   ", Price: 11.65},
			{ShortDescription: "Bananas", Price: 3.35},
		},
		Total: 15,
	}
	testReceipts[2] = Receipt{
		Retailer:     "Barnes & Noble 60131",
		PurchaseDate: ReceiptPurchaseDate(time.Date(2023, time.January, 10, 14, 10, 0, 0, time.UTC)),
		PurchaseTime: ReceiptPurchaseTime(time.Date(2023, time.January, 10, 14, 10, 0, 0, time.UTC)),
		Items:        nil,
		Total:        12.25,
	}
	return testReceipts[:]
}

func TestPostReceiptSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/receipts/process", PostReceipt)
	receiptData := map[string]interface{}{
		"retailer":     "ABC",
		"purchaseDate": "2022-01-01",
		"purchaseTime": "17:52",
		"items": []map[string]interface{}{
			{
				"shortDescription": "Testing data",
				"price":            "11.11",
			},
		},
		"total": "35.35",
	}
	jsonData, _ := json.Marshal(receiptData)
	req, err := http.NewRequest("POST", "/receipts/process", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Couldn't create request: %v\n", err)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
	var actualResponse map[string]string
	error := json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
	if error != nil {
		t.Error("Error in parsing the JSON: ", error)
	}

	receivedUUID := actualResponse["id"]
	decodedUUID, uuidError := uuid.Parse(receivedUUID)
	fmt.Print(decodedUUID)

	if uuidError != nil {
		t.Error("Generated response is not a valid UUID: ", uuidError)
	}
}

func TestPostReceiptFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/receipts/process", PostReceipt)
	receiptData := map[string]interface{}{
		"retailer":     "ABC",
		"purchaseDate": "2022-01-0",
		"purchaseTime": "17:52",
		"items": []map[string]interface{}{
			{
				"shortDescription": "Testing data",
				"price":            "11.11",
			},
		},
		"total": "35.35s",
	}
	jsonData, _ := json.Marshal(receiptData)

	req, err := http.NewRequest("POST", "/receipts/process", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Couldn't create request: %v\n", err)
	}

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

}

func TestGetPoints_withValidId(t *testing.T) {
	// create a mock receipt
	gin.SetMode(gin.TestMode)
	testReceipt := ReceiptsforTesting()[0]
	// add mock receipt to map
	receiptMap["validID"] = testReceipt
	// call GetPoints with valid ID
	router := gin.Default()
	router.GET("/:id/points", GetPoints)
	reqValidID, err := http.NewRequest("GET", "/validID/points", nil)
	if err != nil {
		t.Fatalf("Couldn't create request: %v\n", err)
	}
	// check if status is 200
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, reqValidID)
	assert.Equal(t, http.StatusOK, recorder.Code)
	// remove validID entry from map
	delete(receiptMap, "validID")
}

func TestGetPoints_withInvalidId(t *testing.T) {
	// call GetPoints with valid ID
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/:id/points", GetPoints)
	reqInvalidID, err := http.NewRequest("GET", "/invalidID/points", nil)
	// check if status is 200
	if err != nil {
		t.Fatalf("Couldn't create request: %v\n", err)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, reqInvalidID)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestCalculateRetailerNamePoints(t *testing.T) {
	testReceipts := ReceiptsforTesting()
	expectedResponse := [3]int{14, 0, 16}
	var actualResponse [3]int
	for i := 0; i < 3; i++ {
		actualResponse[i] = CalculateRetailerNamePoints(&testReceipts[i])
	}
	assert.Equal(t, expectedResponse, actualResponse)
}

func TestAddRoundDollarPoints(t *testing.T) {
	testReceipts := ReceiptsforTesting()
	expectedResponse := [3]int{0, 50, 0}
	var actualResponse [3]int
	for i := 0; i < 3; i++ {
		actualResponse[i] = AddRoundDollarPoints(&testReceipts[i])
	}
	assert.Equal(t, expectedResponse, actualResponse)
}

func TestAddQuarterDollarPoints(t *testing.T) {
	testReceipts := ReceiptsforTesting()
	expectedResponse := [3]int{0, 25, 25}
	var actualResponse [3]int
	for i := 0; i < 3; i++ {
		actualResponse[i] = AddQuarterDollarPoints(&testReceipts[i])
	}
	assert.Equal(t, expectedResponse, actualResponse)
}

func TestCalculatePurchaseItemsPoints(t *testing.T) {
	testReceipts := ReceiptsforTesting()
	expectedResponse := [3]int{5, 5, 0}
	var actualResponse [3]int
	for i := 0; i < 3; i++ {
		actualResponse[i] = CalculatePurchaseItemsPoints(&testReceipts[i])
	}
	assert.Equal(t, expectedResponse, actualResponse)
}

func TestCalculateItemDescriptionPoints(t *testing.T) {
	testReceipts := ReceiptsforTesting()
	expectedResponse := [3]int{4, 3, 0}
	var actualResponse [3]int
	for i := 0; i < 3; i++ {
		actualResponse[i] = CalculateItemDescriptionPoints(&testReceipts[i])
	}
	assert.Equal(t, expectedResponse, actualResponse)
}

func TestAddOddPurchaseDayPoints(t *testing.T) {
	testReceipts := ReceiptsforTesting()
	expectedResponse := [3]int{6, 0, 0}
	var actualResponse [3]int
	for i := 0; i < 3; i++ {
		actualResponse[i] = AddOddPurchaseDayPoints(&testReceipts[i])
	}
	assert.Equal(t, expectedResponse, actualResponse)
}

func TestAddPurchaseTimePoints(t *testing.T) {
	testReceipts := ReceiptsforTesting()
	expectedResponse := [3]int{10, 0, 10}
	var actualResponse [3]int
	for i := 0; i < 3; i++ {
		actualResponse[i] = AddPurchaseTimePoints(&testReceipts[i])
	}
	assert.Equal(t, expectedResponse, actualResponse)
}
