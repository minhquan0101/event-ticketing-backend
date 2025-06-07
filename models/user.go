package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Email       string             `bson:"email" json:"email"`
	Password    string             `bson:"password,omitempty" json:"-"`
	Role        string             `bson:"role" json:"role"` // "user" hoặc "admin"
	IsVerified  bool               `bson:"is_verified" json:"is_verified"` // đã xác minh Gmail chưa
	VerifyCode  string             `bson:"verify_code,omitempty" json:"-"` // mã xác minh
	VerifyExpiresAt  primitive.DateTime `bson:"verify_expires_at,omitempty" json:"verify_expires_at,omitempty"`
}
