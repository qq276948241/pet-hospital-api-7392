package handlers

import (
	"pet-hospital/models"
	"pet-hospital/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PetRequest struct {
	Name        string  `json:"name" binding:"required"`
	Species     string  `json:"species" binding:"required"`
	Breed       string  `json:"breed"`
	Gender      string  `json:"gender"`
	Age         int     `json:"age"`
	Weight      float64 `json:"weight"`
	Color       string  `json:"color"`
	BirthDate   string  `json:"birth_date"`
	Neutered    bool    `json:"neutered"`
	Description string  `json:"description"`
}

func CreatePet(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req PetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	pet := models.Pet{
		OwnerID:     userID,
		Name:        req.Name,
		Species:     req.Species,
		Breed:       req.Breed,
		Gender:      req.Gender,
		Age:         req.Age,
		Weight:      req.Weight,
		Color:       req.Color,
		BirthDate:   req.BirthDate,
		Neutered:    req.Neutered,
		Description: req.Description,
	}

	if err := models.DB.Create(&pet).Error; err != nil {
		utils.InternalError(c, "创建宠物档案失败: "+err.Error())
		return
	}

	utils.Success(c, pet)
}

func GetPets(c *gin.Context) {
	userID := c.GetUint("user_id")

	var pets []models.Pet
	if err := models.DB.Where("owner_id = ?", userID).Find(&pets).Error; err != nil {
		utils.InternalError(c, "获取宠物列表失败: "+err.Error())
		return
	}

	utils.Success(c, pets)
}

func GetPetByID(c *gin.Context) {
	userID := c.GetUint("user_id")
	petID, _ := strconv.Atoi(c.Param("id"))

	var pet models.Pet
	if err := models.DB.Where("id = ? AND owner_id = ?", petID, userID).First(&pet).Error; err != nil {
		utils.NotFound(c, "宠物档案不存在")
		return
	}

	utils.Success(c, pet)
}

func UpdatePet(c *gin.Context) {
	userID := c.GetUint("user_id")
	petID, _ := strconv.Atoi(c.Param("id"))

	var pet models.Pet
	if err := models.DB.Where("id = ? AND owner_id = ?", petID, userID).First(&pet).Error; err != nil {
		utils.NotFound(c, "宠物档案不存在")
		return
	}

	var req PetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	pet.Name = req.Name
	pet.Species = req.Species
	pet.Breed = req.Breed
	pet.Gender = req.Gender
	pet.Age = req.Age
	pet.Weight = req.Weight
	pet.Color = req.Color
	pet.BirthDate = req.BirthDate
	pet.Neutered = req.Neutered
	pet.Description = req.Description

	if err := models.DB.Save(&pet).Error; err != nil {
		utils.InternalError(c, "更新宠物档案失败: "+err.Error())
		return
	}

	utils.Success(c, pet)
}

func DeletePet(c *gin.Context) {
	userID := c.GetUint("user_id")
	petID, _ := strconv.Atoi(c.Param("id"))

	var pet models.Pet
	if err := models.DB.Where("id = ? AND owner_id = ?", petID, userID).First(&pet).Error; err != nil {
		utils.NotFound(c, "宠物档案不存在")
		return
	}

	if err := models.DB.Delete(&pet).Error; err != nil {
		utils.InternalError(c, "删除宠物档案失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}
