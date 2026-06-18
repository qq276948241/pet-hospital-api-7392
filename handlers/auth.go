package handlers

import (
	"net/http"
	"pet-hospital/models"
	"pet-hospital/utils"

	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Email    string `json:"email"`
	RealName string `json:"real_name"`
	Address  string `json:"address"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	var existingUser models.User
	if err := models.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		utils.BadRequest(c, "用户名已存在")
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.InternalError(c, "密码加密失败")
		return
	}

	user := models.User{
		Username: req.Username,
		Password: hashedPassword,
		Phone:    req.Phone,
		Email:    req.Email,
		RealName: req.RealName,
		Address:  req.Address,
		Role:     "user",
	}

	if err := models.DB.Create(&user).Error; err != nil {
		utils.InternalError(c, "注册失败: "+err.Error())
		return
	}

	token, err := utils.GenerateToken(&user)
	if err != nil {
		utils.InternalError(c, "生成令牌失败")
		return
	}

	utils.Success(c, gin.H{
		"user":  user,
		"token": token,
	})
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	var user models.User
	if err := models.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		utils.BadRequest(c, "用户名或密码错误")
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		utils.BadRequest(c, "用户名或密码错误")
		return
	}

	token, err := utils.GenerateToken(&user)
	if err != nil {
		utils.InternalError(c, "生成令牌失败")
		return
	}

	utils.Success(c, gin.H{
		"user":  user,
		"token": token,
	})
}

func GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := models.DB.Preload("Pets").First(&user, userID).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	utils.Success(c, user)
}

func UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	var req struct {
		Phone    string `json:"phone"`
		Email    string `json:"email"`
		RealName string `json:"real_name"`
		Address  string `json:"address"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Address != "" {
		user.Address = req.Address
	}

	if err := models.DB.Save(&user).Error; err != nil {
		utils.InternalError(c, "更新失败: "+err.Error())
		return
	}

	utils.Success(c, user)
}

func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, utils.Response{
		Code:    0,
		Message: "退出登录成功",
	})
}
