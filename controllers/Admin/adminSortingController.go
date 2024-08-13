package controllers

import (
	"mountgear/helpers"
	"mountgear/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Sorting(c *gin.Context) {
	sortBy := c.Query("sortBy")

	if sortBy == "" {
		helpers.SendResponse(c, http.StatusBadRequest, "Sort parameter is empty", nil)
		return
	}

	if sortBy != "popular_product" && sortBy != "popular_category" {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid sort option selected", nil)
		return
	}

	var result interface{}
	var err error

	switch sortBy {
	case "popular_product":
		result, err = getPopularProducts(models.DB)
	case "popular_category":
		result, err = getPopularCategories(models.DB)
	}

	if err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Error fetching sorted data", nil)
		return
	}

	helpers.SendResponse(c, http.StatusOK, "Data sorted successfully", nil, gin.H{"result": result})
}

func getPopularProducts(db *gorm.DB) ([]struct {
	ProductID   uint
	Name        string
	TotalSold   int
	TotalAmount float64
}, error) {
	var results []struct {
		ProductID   uint
		Name        string
		TotalSold   int
		TotalAmount float64
	}

	err := db.Table("products").
		Select("products.id as product_id, products.name, SUM(order_items.quantity) as total_sold, SUM(order_items.discounted_price * order_items.quantity) as total_amount").
		Joins("JOIN order_items ON products.id = order_items.product_id").
		Joins("JOIN orders ON order_items.order_id = orders.id").
		Where("orders.status NOT IN ?", []string{"Canceled", "Return"}).
		Where("order_items.is_canceled = ?", false).
		Group("products.id").
		Order("total_sold DESC").
		Limit(5).
		Scan(&results).Error

	return results, err
}

func getPopularCategories(db *gorm.DB) ([]struct {
	CategoryID uint
	Name       string
	TotalSold  int
}, error) {
	var results []struct {
		CategoryID uint
		Name       string
		TotalSold  int
	}

	err := db.Table("categories").
		Select("categories.id as category_id, categories.name, SUM(order_items.quantity) as total_sold").
		Joins("JOIN products ON categories.id = products.category_id").
		Joins("JOIN order_items ON products.id = order_items.product_id").
		Joins("JOIN orders ON order_items.order_id = orders.id").
		Where("orders.status NOT IN ?", []string{"Canceled", "Return"}).
		Where("order_items.is_canceled = ?", false).
		Group("categories.id").
		Order("total_sold DESC").
		Limit(2).
		Scan(&results).Error

	return results, err
}

type SalesData struct {
	Date  string  `json:"date"`
	Total float64 `json:"total"`
}

func Charts(c *gin.Context) {
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

	var salesData []SalesData
	err = models.DB.Model(&models.Order{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Where("status = ?", "Delivered").
		Select("DATE(created_at) as date, SUM(final_amount) as total").
		Group("DATE(created_at)").
		Order("date").
		Find(&salesData).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching sales data: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sales_data": salesData})
	c.HTML(http.StatusOK, "reports,html", gin.H{})

}
