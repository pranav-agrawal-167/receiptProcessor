package main

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Item struct {
	ShortDescription string  `json:"shortDescription"`
	Price            float64 `json:"price,string"`
}

type ReceiptPurchaseDate time.Time

type ReceiptPurchaseTime time.Time

type Receipt struct {
	Retailer     string              `json:"Retailer"`
	PurchaseDate ReceiptPurchaseDate `json:"purchaseDate"`
	PurchaseTime ReceiptPurchaseTime `json:"purchaseTime"`
	Items        []Item              `json:"items"`
	Total        float64             `json:"total,string"`
}

// map to store receiptID and receipt as a key value pair
var receiptMap = make(map[string]Receipt)

func main() {
	router := gin.Default()
	receiptsGroup := router.Group("/receipts")

	receiptsGroup.POST("/process", PostReceipt)
	receiptsGroup.GET("/:id/points", GetPoints)

	router.Run(":8080")
}

func PostReceipt(c *gin.Context) {
	var receipt Receipt
	// process JSON payload to store as a Receipt struct
	if err := c.BindJSON(&receipt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// generate UUID
	receiptId := uuid.New().String()

	receiptMap[receiptId] = receipt
	c.JSON(http.StatusOK, gin.H{"id": receiptId})
}

func GetPoints(c *gin.Context) {
	receiptId := c.Param("id")
	receipt, ok := receiptMap[receiptId]
	if ok {
		points := CalculatePoints(&receipt)
		c.JSON(http.StatusOK, gin.H{"points": points})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
	}
}

func CalculatePoints(receipt *Receipt) int {
	var pointsRewarded = 0

	pointsRewarded += CalculateRetailerNamePoints(receipt)

	pointsRewarded += AddRoundDollarPoints(receipt)

	pointsRewarded += AddQuarterDollarPoints(receipt)

	pointsRewarded += CalculatePurchaseItemsPoints(receipt)

	pointsRewarded += CalculateItemDescriptionPoints(receipt)

	pointsRewarded += AddOddPurchaseDayPoints(receipt)

	pointsRewarded += AddPurchaseTimePoints(receipt)

	return pointsRewarded
}

func CalculateRetailerNamePoints(receipt *Receipt) int {
	var count = 0
	if receipt.Retailer != "" {
		// Rule 1: One point for each alphanumeric character in retailer name
		for _, char := range receipt.Retailer {
			if unicode.IsLetter(char) || unicode.IsDigit(char) {
				count++
			}
		}
	}
	return count
}

func AddRoundDollarPoints(receipt *Receipt) int {
	// Rule 2: 50 points if the total is a round dollar amount with no cents.
	if math.Mod(receipt.Total, 1) == 0 {
		return 50
	}
	return 0
}

func AddQuarterDollarPoints(receipt *Receipt) int {
	// Rule 3: 25 points if the total is a multiple of 0.25.
	if math.Mod(receipt.Total, 0.25) == 0 {
		return 25
	}
	return 0
}

func CalculatePurchaseItemsPoints(receipt *Receipt) int {
	// Rule 4: 5 points for every two items on the receipt.
	if receipt.Items != nil {
		return len(receipt.Items) / 2 * 5
	}
	return 0
}

func CalculateItemDescriptionPoints(receipt *Receipt) int {
	// Rule 5: If the trimmed length of the item description is a multiple of 3, multiply the price by 0.2 and round up to the nearest integer.
	var earnedPoints = 0
	if receipt.Items != nil {
		for _, item := range receipt.Items {
			if item.ShortDescription != "" {
				if len(strings.TrimSpace(item.ShortDescription))%3 == 0 {
					earnedPoints += int(math.Ceil(item.Price * 0.2))
				}
			}
		}
	}
	return earnedPoints
}

func AddOddPurchaseDayPoints(receipt *Receipt) int {
	// Rule 6: 6 points if the day in the purchase date is odd.
	day := time.Time(receipt.PurchaseDate).Day()
	if day%2 != 0 {
		return 6
	}
	return 0
}

func AddPurchaseTimePoints(receipt *Receipt) int {
	// Rule 7: 10 points if the time of purchase is after 2:00pm and before 4:00pm.
	purTime := time.Time(receipt.PurchaseTime)
	startTime := time.Date(purTime.Year(), purTime.Month(), purTime.Day(), 14, 0, 0, 0, purTime.Location()) // 2:00 PM
	endTime := time.Date(purTime.Year(), purTime.Month(), purTime.Day(), 16, 0, 0, 0, purTime.Location())   // 4:00 PM

	if purTime.After(startTime) && purTime.Before(endTime) {
		return 10
	}
	return 0
}

// Unmarshal JSON to store the PurchaseDate and PurchaseTime in the required format
func (purchaseDate *ReceiptPurchaseDate) UnmarshalJSON(b []byte) error {
	trimmedDateString := strings.Trim(string(b), "\"")
	formattedDate, err := time.Parse("2006-01-02", trimmedDateString)
	if err != nil {
		return fmt.Errorf("failed to parse ReceiptPurchaseDate: %w", err)
	}
	*purchaseDate = ReceiptPurchaseDate(formattedDate)
	return nil
}

func (purchaseTime *ReceiptPurchaseTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	formattedTime, err := time.Parse("15:04", s)
	if err != nil {
		return fmt.Errorf("failed to parse ReceiptPurchaseTime: %w", err)
	}
	*purchaseTime = ReceiptPurchaseTime(formattedTime)
	return nil
}
