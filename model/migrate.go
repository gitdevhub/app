package model

func Setup() {
	_ = db.GetDB().AutoMigrate(&User{})
	_ = db.GetDB().AutoMigrate(&Account{})


	_ = db.GetDB().AutoMigrate(&JwtAuth{})
	_ = db.GetDB().AutoMigrate(&Role{})
	_ = db.GetDB().AutoMigrate(&UserRole{})


	_ = db.GetDB().AutoMigrate(&PushNotification{})

	_ = db.GetDB().AutoMigrate(&Payment{})
	_ = db.GetDB().AutoMigrate(&PaymentType{})


	_ = db.GetDB().AutoMigrate(&Message{})
	_ = db.GetDB().AutoMigrate(&MessageStatus{})
	_ = db.GetDB().AutoMigrate(&Room{})
	_ = db.GetDB().AutoMigrate(&UserRoom{})



	_ = db.GetDB().AutoMigrate(&Currency{})

	_ = db.GetDB().AutoMigrate(&Comment{})
	_ = db.GetDB().AutoMigrate(&CommentLike{})
	_ = db.GetDB().AutoMigrate(&Follower{})
	_ = db.GetDB().AutoMigrate(&Withdraw{})
}
