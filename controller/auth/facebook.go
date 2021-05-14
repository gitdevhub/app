package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	fb "github.com/huandu/facebook/v2"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"html"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type FacebookForm struct {
	UserId string `form:"user_id" json:"user_id" binding:"required,min=1,max=100"`
	Token  string `form:"token" json:"token" binding:"required,min=1,max=2048"`
}

type FacebookResult struct {
	ID        string                `json:"id"`
	FirstName string                `json:"first_name"`
	LastName  string                `json:"last_name"`
	Name      string                `json:"name"`
	Birthday  string                `json:"birthday"`
	Email     string                `json:"email"`
	Gender    string                `json:"gender"`
	Picture   FacebookPictureResult `json:"picture"`
}

type FacebookPictureResult struct {
	Data FacebookPictureDataResult `json:"data"`
}

type FacebookPictureDataResult struct {
	Url          string `json:"url"`
	IsSilhouette bool   `json:"is_silhouette"`
}

type FacebookInspectResult struct {
	AppId   string `json:"app_id"`
	UserId  string `json:"user_id"`
	IsValid bool   `json:"is_valid"`
}

func FacebookLogin(c *gin.Context) {
	var form FacebookForm

	if err := c.ShouldBindJSON(&form); err != nil {
		logger.Error("the error form validation: ", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid json provided"})
		return
	}

	uniqueId := html.EscapeString(strings.TrimSpace(c.Request.Header.Get("X-Request-ID")))
	if uniqueId == "" {
		logger.Error("the error form validation: empty unique id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token not valid"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userIdForm := html.EscapeString(strings.TrimSpace(form.UserId))
	tokenForm := html.EscapeString(strings.TrimSpace(form.Token))

	// validate token
	var result FacebookResult
	var inspectResult FacebookInspectResult

	//fb.Version = "v10.0"
	globalApp := fb.New(config.Auth.FacebookId, config.Auth.FacebookSecret)

	// enable "appsecret_proof" for all sessions created by this app.
	globalApp.EnableAppsecretProof = true

	session := globalApp.Session(tokenForm)
	session.SetDebug(fb.DEBUG_OFF)
	session.Version = "v10.0"

	inspect, err := session.Inspect()
	if err != nil {
		logger.Errorf("token validation failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide valid login details"})
		return
	}

	err = inspect.Decode(&inspectResult)
	if err != nil {
		logger.Errorf("token validation failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide valid login details"})
		return
	}

	if !inspectResult.IsValid {
		logger.Error("token validation failed: IsValid")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide valid login details"})
		return
	}

	if config.Auth.FacebookId != inspectResult.AppId {
		logger.Error("token validation failed: FacebookId")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide valid login details"})
		return
	}

	res, err := session.WithContext(ctx).Get("/"+userIdForm, fb.Params{
		"fields": "id, name, first_name, last_name, email, gender, birthday, location, picture.width(300).height(300), link, is_verified",
		//"access_token": tokenForm,
	})
	if err != nil {
		logger.Errorf("token validation failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide valid login details"})
		return
	}

	err = res.Decode(&result)
	if err != nil {
		logger.Errorf("token validation failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide valid login details"})
		return
	}

	logger.Infof("data : %v", res)
	logger.Infof("data : %v", result)

	if result.ID == "" {
		logger.Errorf("token validation failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide valid login details"})
		return
	}

	email := result.Email
	var imageUrl string

	if email == "" {
		email = result.ID + "@facebook.com"
	}

	email = html.EscapeString(strings.ToLower(strings.TrimSpace(email)))

	var userId int64

	// check user is signup
	user, err := model.GetUserByEmail(email)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error. Try again"})
			return
		}

		var filePath string

		// check default image
		if !result.Picture.Data.IsSilhouette {
			fileName := fmt.Sprintf("facebook-%s.png", result.ID)
			exPath := util.GetExecutablePath()
			filePath = filepath.Join(exPath, config.App.UploadDir, "images/", fileName)

			// download image
			_, err = util.DownloadImage(result.Picture.Data.Url, filePath)
			if err != nil {
				logger.Errorf("image error: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Image error"})
				return
			}

			imageUrl = fmt.Sprintf("%s/images/%s", config.Server.GetFullHostName(), fileName)
		}

		var gender null.Int

		if result.Gender == "male" {
			gender = null.IntFrom(model.UserGenderMale)
		} else if result.Gender == "female" {
			gender = null.IntFrom(model.UserGenderFemale)
		}

		var birthday null.Time

		if result.Birthday != "" {
			d, err := time.Parse("02/01/2006", result.Birthday)
			if err == nil {
				birthday = null.TimeFrom(d)
			}
		}

		user := model.User{
			Name:      html.EscapeString(result.Name),
			Email:     email,
			Password:  util.RandStringRunes(20),
			Image:     null.StringFrom(imageUrl),
			Gender:    gender,
			Birthday:  birthday,
			IsSocial:  null.IntFrom(model.UserIsSocial),
			CreatedAt: time.Now(),
		}

		if _, err := model.CreateUser(&user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error. Try again"})
			return
		}

		userId = user.ID
	} else {
		if err := jwt.DeleteAllRedisJWTAuth(user.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error. Try again"})
			return
		}

		if err := model.DeleteAllJWTAuth(user.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error. Try again"})
			return
		}

		userId = user.ID
	}

	tokens, err := auth.Authenticate(userId, uniqueId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error. Try again"})
		return
	}

	c.JSON(http.StatusOK, tokens)
}
