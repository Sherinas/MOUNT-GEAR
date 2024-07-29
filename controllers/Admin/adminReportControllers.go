package controllers

import (
	"fmt"
	"mountgear/models"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

type SalesReportItem struct {
	OrderID        uint    `json:"order_id"`
	UserID         uint    `json:"user_id"`
	CustomerName   string  `json:"customer_name"`
	FinalAmount    float64 `json:"final_amount"`
	PaymentMethod  string  `json:"payment_method"`
	CouponDiscount float64 `json:"coupon_discount"`
	Status         string  `json:"status"`
	Date           string  `json:"date"`
	TotalQuantity  int     `json:"total_quantity"`
}

func SalesReport(c *gin.Context) {
	filterData := c.Query("filter")

	var startTime, endTime time.Time
	var err error

	// Set time range based on filter
	switch filterData {
	case "daily":
		startTime = time.Now().AddDate(0, 0, -1)
		endTime = time.Now()
	case "weekly":
		startTime = time.Now().AddDate(0, 0, -7)
		endTime = time.Now()
	case "monthly":
		startTime = time.Now().AddDate(0, -1, 0)
		endTime = time.Now()
	case "yearly":
		startTime = time.Now().AddDate(-1, 0, 0)
		endTime = time.Now()
	case "custom":
		startStr := c.Query("start_date")
		endStr := c.Query("end_date")
		startTime, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use YYYY-MM-DD"})
			return
		}
		endTime, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use YYYY-MM-DD"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter parameter"})
		return
	}

	var orders []models.Order
	err = models.DB.Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Select("id", "user_id", "final_amount", "payment_method", "coupon_discount", "status", "created_at").
		Find(&orders).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching orders: " + err.Error()})
		return
	}

	var report []SalesReportItem
	for _, order := range orders {
		var user models.User
		err := models.DB.Select("name").First(&user, order.UserID).Error
		if err != nil {

			user.Name = "Unknown User"
		}

		var items []models.OrderItem
		if err := models.DB.Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching order items: " + err.Error()})
			return
		}

		totalQuantity := 0
		for _, item := range items {
			totalQuantity += item.Quantity
		}

		reportItem := SalesReportItem{
			OrderID:        order.ID,
			UserID:         order.UserID,
			CustomerName:   user.Name,
			FinalAmount:    order.FinalAmount,
			PaymentMethod:  order.PaymentMethod,
			CouponDiscount: order.CouponDiscount,
			Status:         order.Status,
			Date:           order.CreatedAt.Format("2006-01-02"),
			TotalQuantity:  totalQuantity,
		}
		report = append(report, reportItem)
	}

	pdfPath, err := generatePDF(report)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating PDF: " + err.Error()})
		return
	}

	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(pdfPath)))
	c.Writer.Header().Set("Content-Type", "application/pdf")

	c.File(pdfPath)

	c.JSON(http.StatusOK, gin.H{
		"message": "PDF generated successfully",
	})

}

func generatePDF(report []SalesReportItem) (string, error) {

	pdfPath := filepath.Join("temp", "sales_report_"+time.Now().Format("20060102150405")+".pdf")

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Sales Report")
	pdf.Ln(15)

	pdf.SetFont("Arial", "B", 10)
	headers := []string{"Order ID", "Customer Name", "Payment Method", "Final Amount", "Status", "Order Date"}
	for _, header := range headers {
		pdf.Cell(32, 10, header)
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	var totalAmount float64
	for _, item := range report {
		pdf.Cell(32, 8, strconv.Itoa(int(item.OrderID)))
		pdf.Cell(32, 8, item.CustomerName)
		pdf.Cell(32, 8, item.PaymentMethod)
		pdf.Cell(32, 8, strconv.FormatFloat(item.FinalAmount, 'f', 2, 64))
		pdf.Cell(32, 8, item.Status)
		pdf.Cell(32, 8, item.Date)
		pdf.Ln(-1)
		totalAmount += item.FinalAmount
	}

	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 10, fmt.Sprintf("Total Sales Count: %d", len(report)))
	pdf.Ln(-1)
	pdf.Cell(0, 10, fmt.Sprintf("Total Amount: %.2f", totalAmount))

	tempFilePath := "C:/Users/Sherinas/Downloads/sales_report.pdf"
	err := pdf.OutputFileAndClose(tempFilePath)
	if err != nil {
		return "", err
	}

	return pdfPath, nil
}
