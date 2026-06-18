package handlers

import (
	"pet-hospital/models"
	"pet-hospital/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MedicineRequest struct {
	Name        string  `json:"name" binding:"required"`
	Category    string  `json:"category"`
	Price       float64 `json:"price" binding:"required"`
	Stock       int     `json:"stock"`
	Unit        string  `json:"unit"`
	Description string  `json:"description"`
	Supplier    string  `json:"supplier"`
}

func GetMedicines(c *gin.Context) {
	category := c.Query("category")
	keyword := c.Query("keyword")
	lowStock := c.Query("low_stock")

	var medicines []models.Medicine
	query := models.DB

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if lowStock == "true" {
		query = query.Where("stock < ?", 20)
	}

	if err := query.Order("created_at DESC").Find(&medicines).Error; err != nil {
		utils.InternalError(c, "获取药品列表失败: "+err.Error())
		return
	}

	utils.Success(c, medicines)
}

func GetMedicineByID(c *gin.Context) {
	medicineID, _ := strconv.Atoi(c.Param("id"))

	var medicine models.Medicine
	if err := models.DB.First(&medicine, medicineID).Error; err != nil {
		utils.NotFound(c, "药品不存在")
		return
	}

	utils.Success(c, medicine)
}

func CreateMedicine(c *gin.Context) {
	var req MedicineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	medicine := models.Medicine{
		Name:        req.Name,
		Category:    req.Category,
		Price:       req.Price,
		Stock:       req.Stock,
		Unit:        req.Unit,
		Description: req.Description,
		Supplier:    req.Supplier,
	}

	if err := models.DB.Create(&medicine).Error; err != nil {
		utils.InternalError(c, "添加药品失败: "+err.Error())
		return
	}

	utils.Success(c, medicine)
}

func UpdateMedicine(c *gin.Context) {
	medicineID, _ := strconv.Atoi(c.Param("id"))

	var medicine models.Medicine
	if err := models.DB.First(&medicine, medicineID).Error; err != nil {
		utils.NotFound(c, "药品不存在")
		return
	}

	var req MedicineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	medicine.Name = req.Name
	medicine.Category = req.Category
	medicine.Price = req.Price
	medicine.Stock = req.Stock
	medicine.Unit = req.Unit
	medicine.Description = req.Description
	medicine.Supplier = req.Supplier

	if err := models.DB.Save(&medicine).Error; err != nil {
		utils.InternalError(c, "更新药品失败: "+err.Error())
		return
	}

	utils.Success(c, medicine)
}

func UpdateStock(c *gin.Context) {
	medicineID, _ := strconv.Atoi(c.Param("id"))

	var medicine models.Medicine
	if err := models.DB.First(&medicine, medicineID).Error; err != nil {
		utils.NotFound(c, "药品不存在")
		return
	}

	var req struct {
		Stock int `json:"stock" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	medicine.Stock = req.Stock

	if err := models.DB.Save(&medicine).Error; err != nil {
		utils.InternalError(c, "更新库存失败: "+err.Error())
		return
	}

	utils.Success(c, medicine)
}

func AdjustStock(c *gin.Context) {
	medicineID, _ := strconv.Atoi(c.Param("id"))

	var medicine models.Medicine
	if err := models.DB.First(&medicine, medicineID).Error; err != nil {
		utils.NotFound(c, "药品不存在")
		return
	}

	var req struct {
		Quantity int    `json:"quantity" binding:"required"`
		Type     string `json:"type" binding:"required"`
		Remark   string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if req.Type == "in" {
		medicine.Stock += req.Quantity
	} else if req.Type == "out" {
		if medicine.Stock < req.Quantity {
			utils.BadRequest(c, "库存不足")
			return
		}
		medicine.Stock -= req.Quantity
	} else {
		utils.BadRequest(c, "无效的操作类型")
		return
	}

	if err := models.DB.Save(&medicine).Error; err != nil {
		utils.InternalError(c, "调整库存失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "库存调整成功", medicine)
}

func DeleteMedicine(c *gin.Context) {
	medicineID, _ := strconv.Atoi(c.Param("id"))

	var medicine models.Medicine
	if err := models.DB.First(&medicine, medicineID).Error; err != nil {
		utils.NotFound(c, "药品不存在")
		return
	}

	if err := models.DB.Delete(&medicine).Error; err != nil {
		utils.InternalError(c, "删除药品失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}

func GetMedicineCategories(c *gin.Context) {
	var categories []string
	models.DB.Model(&models.Medicine{}).Distinct("category").Pluck("category", &categories)

	utils.Success(c, categories)
}

func GetLowStockAlert(c *gin.Context) {
	threshold := 20
	if t := c.Query("threshold"); t != "" {
		threshold, _ = strconv.Atoi(t)
	}

	var medicines []models.Medicine
	if err := models.DB.Where("stock < ?", threshold).Order("stock ASC").Find(&medicines).Error; err != nil {
		utils.InternalError(c, "获取库存预警失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"threshold": threshold,
		"count":     len(medicines),
		"list":      medicines,
	})
}
