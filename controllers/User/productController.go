package controllers

import (
	"mountgear/helpers"
	"mountgear/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ...............................................................Shop page.............................................................
func GetShopPage(ctx *gin.Context) {
	var products []models.Product
	var totalCount int64

	// Parse query parameters
	categoryParam := ctx.Query("category")
	inStock := ctx.Query("in_stock")
	minPrice := ctx.Query("min_price")
	maxPrice := ctx.Query("max_price")
	search := ctx.Query("search")
	sort := ctx.Query("sort")
	// page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	// perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))

	// Build the base query
	query := models.DB.Model(&models.Product{}).
		Joins("JOIN categories ON products.category_id = categories.id").
		Where("products.is_active = ? AND categories.is_active = ?", true, true)

	// Apply category filter
	if categoryParam != "" {
		categoryID, err := strconv.Atoi(categoryParam)
		if err == nil {
			query = query.Where("products.category_id = ?", categoryID)
		} else {
			query = query.Where("categories.name LIKE ?", "%"+categoryParam+"%")
		}
	}

	// Apply other filters
	if inStock == "true" {
		query = query.Where("products.stock > 0")
	}
	if minPrice != "" {
		if minPriceFloat, err := strconv.ParseFloat(minPrice, 64); err == nil {
			query = query.Where("products.price >= ?", minPriceFloat)
		}
	}
	if maxPrice != "" {
		if maxPriceFloat, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			query = query.Where("products.price <= ?", maxPriceFloat)
		}
	}

	// Apply search
	if search != "" {
		search = "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(products.name) LIKE ? OR LOWER(products.description) LIKE ?", search, search)
	}

	// Apply sorting
	switch sort {
	case "popularity":
		query = query.Order("products.popularity DESC")
	case "price_asc":
		query = query.Order("products.price ASC")
	case "price_desc":
		query = query.Order("products.price DESC")
	case "rating":
		query = query.Order("products.average_rating DESC")
	case "featured":
		query = query.Order("products.featured DESC")
	case "new_arrivals":
		query = query.Order("products.created_at DESC")
	case "name_asc":
		query = query.Order("products.name ASC")
	case "name_desc":
		query = query.Order("products.name DESC")
	default:
		query = query.Order("products.id ASC") // Default sorting
	}

	// Count total matching products
	query.Count(&totalCount)

	// Apply pagination
	// offset := (page - 1) * perPage
	// query = query.Offset(offset).Limit(perPage)

	// Execute the query
	if err := query.Preload("Category").Preload("Images", "id IN (SELECT MIN(id) FROM images GROUP BY product_id)").Find(&products).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}
	helpers.SendResponse(ctx, http.StatusOK, "", nil, gin.H{"products": products, "total_count": totalCount})

	// "page":        page,
	// "per_page":    perPage,

}

// .....................................................Product Deatils..........................................................
func GetProductDetails(ctx *gin.Context) {
	id := ctx.Param("id")
	var product models.Product

	if err := models.GetRecordByID(models.DB.Preload("Images"), &product, id); err != nil {
		helpers.SendResponse(ctx, http.StatusInternalServerError, "not fetch images", nil)

	}
	helpers.SendResponse(ctx, http.StatusOK, "", nil, gin.H{"Product": product})

}

func ProductSerch(c *gin.Context) {

	query := c.Query("query")

	var products []models.Product
	if err := models.DB.Where("LOWER(name) LIKE LOWER(?)", "%"+query+"%").Find(&products).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "could not search users", nil)
		return
	}
	helpers.SendResponse(c, http.StatusOK, "", nil, gin.H{"products": products})

}

func AddToCart(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var input struct {
		ProductID uint `json:"product_id" binding:"required"`
		Quantity  int  `json:"quantity" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "struct not bind", nil)

		return
	}
	// Find the product
	var product models.Product
	if err := models.DB.First(&product, input.ProductID).Error; err != nil {
		helpers.SendResponse(c, http.StatusNotFound, "product not found", nil)

		return
	}
	if !product.IsActive {
		helpers.SendResponse(c, http.StatusNotFound, "product not found", nil)

		return
	}
	// Find or create cart for the user
	var cart models.Cart
	if err := models.DB.FirstOrCreate(&cart, models.Cart{UserID: userID.(uint)}).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "could not find or create cart", nil)

		return
	}

	// Check if the product is already in the cart
	var cartItem models.CartItem
	result := models.DB.Where("cart_id = ? AND product_id = ?", cart.ID, input.ProductID).First(&cartItem)

	if result.Error == gorm.ErrRecordNotFound {
		// Product not in cart, add new cart item
		if input.Quantity > 5 {
			helpers.SendResponse(c, http.StatusBadRequest, "Cannot have more than 5 quantities of a product in cart", nil)
			return
		}

		if int32(input.Quantity) > product.Stock {
			helpers.SendResponse(c, http.StatusBadRequest, "Not enough stock", nil)

			return
		}

		cartItem = models.CartItem{
			CartID:    cart.ID,
			ProductID: input.ProductID,
			Quantity:  input.Quantity,
		}

		if err := models.DB.Create(&cartItem).Error; err != nil {
			helpers.SendResponse(c, http.StatusInternalServerError, "could not add product to cart", nil)
			return
		}
	} else if result.Error != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to check cart", nil)

		return
	} else {
		// Product already in cart, update quantity
		newQuantity := cartItem.Quantity + input.Quantity

		if newQuantity > 5 {
			helpers.SendResponse(c, http.StatusBadRequest, "Cannot have more than 5 quantities of a product in cart", nil)

			return
		}

		if int32(newQuantity) > product.Stock {
			helpers.SendResponse(c, http.StatusBadRequest, "Not enough stock", nil)

			return
		}

		cartItem.Quantity = newQuantity

		if err := models.DB.Save(&cartItem).Error; err != nil {
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update cart item", nil)

			return
		}
	}
	helpers.SendResponse(c, http.StatusOK, "Product added to cart successfully", nil)

}
