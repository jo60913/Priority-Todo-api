package model

type FcmInfo struct {
	FcmValue bool   `firestore:"FCM"`
	FCMToken string `firestore:"FCMToken"` // in millions
}
