package handlers

import (
	"pet-hospital/models"
	"pet-hospital/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BillItemRequest struct {
	ItemType  string  `json:"item_type" binding:"required"`
	ItemName  string  `json:"item_name" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required"`
	UnitPrice float64 `json:"unit_price" binding:"required"`
}

type BillRequest struct {
	AppointmentID uint              `json:"appointment_id"`
	UserID        uint              `json:"user_id" binding:"required"`
	PetID         uint              `json:"pet_id" binding:"required"`
	Remark        string            `json:"remark"`
	Items         []BillItemRequest `json:"items" binding:"required"`
}

func CreateBill(c *gin.Context) {
	var req BillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	totalAmount := 0.0
	for _, item := range req.Items {
		totalAmount += float64(item.Quantity) * item.UnitPrice
	}

	err := models.DB.Transaction(func(tx *gorm.DB) error {
		bill := models.Bill{
			AppointmentID: req.AppointmentID,
			UserID:        req.UserID,
			PetID:         req.PetID,
			TotalAmount:   totalAmount,
			PaidAmount:    0,
			Status:        "unpaid",
			Remark:        req.Remark,
		}

		if err := tx.Create(&bill).Error; err != nil {
			return err
		}

		for _, item := range req.Items {
			billItem := models.BillItem{
				BillID:    bill.ID,
				ItemType:  item.ItemType,
				ItemName:  item.ItemName,
				Quantity:  item.Quantity,
				UnitPrice: item.UnitPrice,
				Amount:    float64(item.Quantity) * item.UnitPrice,
			}

			if err := tx.Create(&billItem).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		utils.InternalError(c, "生成账单失败: "+err.Error())
		return
	}

	var bill models.Bill
	models.DB.Preload("Items").Where("user_id = ?", req.UserID).
		Order("created_at DESC").First(&bill)

	utils.Success(c, bill)
}

func GetBills(c *gin.Context) {
	userID := c.GetUint("user_id")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var bills []models.Bill
	var total int64

	query := models.DB.Preload("Items").Preload("Pet").Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Model(&models.Bill{}).Count(&total)
	offset := (page - 1) * pageSize

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&bills).Error; err != nil {
		utils.InternalError(c, "获取账单列表失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"list":       bills,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

func GetBillByID(c *gin.Context) {
	userID := c.GetUint("user_id")
	billID, _ := strconv.Atoi(c.Param("id"))

	var bill models.Bill
	if err := models.DB.Preload("Items").Preload("Pet").Preload("Appointment").
		Where("id = ? AND user_id = ?", billID, userID).First(&bill).Error; err != nil {
		utils.NotFound(c, "账单不存在")
		return
	}

	utils.Success(c, bill)
}

func PayBill(c *gin.Context) {
	userID := c.GetUint("user_id")
	billID, _ := strconv.Atoi(c.Param("id"))

	var bill models.Bill
	if err := models.DB.Where("id = ? AND user_id = ?", billID, userID).First(&bill).Error; err != nil {
		utils.NotFound(c, "账单不存在")
		return
	}

	if bill.Status == "paid" {
		utils.BadRequest(c, "该账单已支付")
		return
	}

	var req struct {
		PaymentMethod string `json:"payment_method" binding:"required"`
		PaidAmount    float64 `json:"paid_amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	paidAmount := bill.TotalAmount
	if req.PaidAmount > 0 {
		paidAmount = req.PaidAmount
	}

	now := time.Now()
	bill.PaidAmount = paidAmount
	bill.PaymentMethod = req.PaymentMethod
	bill.PaymentTime = &now
	bill.Status = "paid"

	if err := models.DB.Save(&bill).Error; err != nil {
		utils.InternalError(c, "支付失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "支付成功", bill)
}

func GetBillsByPet(c *gin.Context) {
	userID := c.GetUint("user_id")
	petID, _ := strconv.Atoi(c.Param("pet_id"))

	var pet models.Pet
	if err := models.DB.Where("id = ? AND owner_id = ?", petID, userID).First(&pet).Error; err != nil {
		utils.NotFound(c, "宠物不存在")
		return
	}

	var bills []models.Bill
	if err := models.DB.Preload("Items").Where("pet_id = ?", petID).
		Order("created_at DESC").Find(&bills).Error; err != nil {
		utils.InternalError(c, "获取账单列表失败: "+err.Error())
		return
	}

	utils.Success(c, bills)
}

func UpdateBillStatus(c *gin.Context) {
	billID, _ := strconv.Atoi(c.Param("id"))

	var bill models.Bill
	if err := models.DB.First(&bill, billID).Error; err != nil {
		utils.NotFound(c, "账单不存在")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	validStatuses := map[string]bool{
		"unpaid":   true,
		"paid":     true,
		"refunded": true,
		"cancelled": true,
	}

	if !validStatuses[req.Status] {
		utils.BadRequest(c, "无效的状态")
		return
	}

	bill.Status = req.Status

	if err := models.DB.Save(&bill).Error; err != nil {
		utils.InternalError(c, "更新状态失败: "+err.Error())
		return
	}

	utils.Success(c, bill)
}
