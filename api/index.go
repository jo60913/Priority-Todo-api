package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/jo60913/Todo-api/model"
	. "github.com/tbxark/g4vercel"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var (
	firebaseSdkAdmin = os.Getenv("FIREBASE_ADMIN_SDK")
	firebasefcmkey   = os.Getenv("TODO_API_FIREBASE_FCM_KEY")
	fcmHeader        = os.Getenv("FCM_HEADER")
	attribute        = "attribute"
	passwordAtt      = "password"
	fcmValue         = "FcmValue"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	server := New()
	sa := option.WithCredentialsJSON([]byte(firebaseSdkAdmin))
	app, newAppErr := firebase.NewApp(context.Background(), nil, sa)
	if newAppErr != nil {
		log.Println("firebase.NewApp錯誤", newAppErr)
		return
	}

	client, err := app.Firestore(context.Background())
	if err != nil {
		log.Println("firestore登入錯誤", err.Error())
		return
	}
	defer client.Close()
	// 新增會員
	server.POST("/insert/user", func(ctx *Context) {
		var userAdd model.UserAdd
		err := json.NewDecoder(ctx.Req.Body).Decode(&userAdd)
		log.Println("/insert/user ", "UserToken : "+userAdd.UserAccount)
		if err != nil {
			log.Println("/insert/user 傳入參數錯誤", err.Error())
			ctx.JSON(http.StatusOK, H{
				"ErrorMsg":  "參數錯誤",
				"ErrorFlag": "1",
			})
			return
		}

		_, readErr := client.Collection(userAdd.UserAccount).Doc(attribute).Get(context.Background())
		if readErr != nil { //沒有notification時新增
			log.Println("/insert/user 找notification時錯誤", readErr.Error())
			_, addErr := client.Collection(userAdd.UserAccount).Doc(attribute).Create(context.Background(), map[string]interface{}{
				"password": userAdd.UserPassword,
				"FcmValue": true,
				"userName": userAdd.UserName,
			})
			if addErr != nil {
				log.Println("/insert/user 新增時錯誤", addErr.Error())
				ctx.JSON(http.StatusOK, H{
					"ErrorMsg":  "新增時錯誤",
					"ErrorFlag": "2",
				})
				return
			}
			log.Println("/insert/user 新增成功")
			ctx.JSON(http.StatusOK, H{
				"ErrorMsg":  "",
				"ErrorFlag": "0",
			})
			return
		}

		ctx.JSON(http.StatusOK, H{
			"ErrorMsg":  "該帳號已存在，請取其他帳號",
			"ErrorFlag": "1",
		})
	})
	// 判斷會員
	server.POST("/login", func(ctx *Context) {
		var userInfo model.UserInfo
		err := json.NewDecoder(ctx.Req.Body).Decode(&userInfo)

		log.Println("/login ", "account : "+userInfo.UserAccount)
		if err != nil {
			log.Println("/login", "json轉換錯誤 "+err.Error())
			ctx.JSON(http.StatusOK, H{
				"ErrorMsg":  "欄位錯誤",
				"ErrorFlag": "3",
			})
			return
		}
		_, readErr := client.Collection(userInfo.UserAccount).Doc(attribute).Get(context.Background())
		if readErr != nil { //沒有notificati
			log.Println("/login 尚未新增", "新增FCM欄位"+readErr.Error())

			ctx.JSON(http.StatusOK, H{
				"ErrorMsg":  "查無帳號",
				"ErrorFlag": "2",
			})
			return

		}

		attributeDoc := client.Collection(userInfo.UserAccount).Doc(attribute)
		getvalue, getDocError := attributeDoc.Get(context.Background())

		if getDocError != nil {
			log.Println("login notification欄位找不到", getDocError.Error())
			ctx.JSON(http.StatusOK, H{
				"ErrorMsg":  "查無帳號",
				"ErrorFlag": "2",
			})
			return
		}

		password, getFcmValueError := getvalue.DataAt(passwordAtt)

		if getFcmValueError != nil {
			log.Println("login ", "找不到帳號")
			ctx.JSON(http.StatusOK, H{
				"ErrorMsg":  "找不到帳號",
				"ErrorFlag": "2",
			})
			return
		}

		if userInfo.UserPassword == password {
			log.Println("login ", "密碼正確")
			ctx.JSON(http.StatusOK, H{
				"ErrorMsg":  "",
				"ErrorFlag": "0",
			})
			return
		}

		log.Println("login ", "密碼錯誤")
		ctx.JSON(http.StatusOK, H{
			"ErrorMsg":  "密碼錯誤",
			"ErrorFlag": "1",
		})
	})
	// 修改會員
	server.POST("/update/userInfo", func(ctx *Context) {
		var firstlogin model.FirstLogin
		err := json.NewDecoder(ctx.Req.Body).Decode(&firstlogin)

		log.Println("update/firstlogin ", "UserToken : "+firstlogin.UserToken)
		if err != nil {
			log.Println("update/firstlogin", "json轉換錯誤 "+err.Error())
			ctx.JSON(http.StatusOK, H{
				"ErrorMsg":  "欄位錯誤",
				"ErrorFlag": "3",
			})
			return
		}
		_, readErr := client.Collection(firstlogin.UserID).Doc("notification").Get(context.Background())
		if readErr != nil { //沒有notification
			log.Println("update/firstlogin 尚未新增", "新增FCM欄位"+readErr.Error())
			_, addErr := client.Collection(firstlogin.UserID).Doc("notification").Create(context.Background(), map[string]interface{}{
				"FCM":      true,
				"FCMToken": firstlogin.UserToken,
			})
			if addErr != nil {
				ctx.JSON(http.StatusOK, H{
					"ErrorMsg":  "新增notification FCM時錯誤",
					"ErrorFlag": "2",
				})
				return
			}
		}

		_, updateErr := client.Collection(firstlogin.UserID).Doc("notification").Update(context.Background(), []firestore.Update{
			{Path: "FCMToken", Value: firstlogin.UserToken},
		})

		if updateErr != nil {
			log.Println("update/firstlogin notification更新錯誤", updateErr.Error())
			ctx.JSON(http.StatusOK, H{
				"ErrorMsg":  "更新錯誤",
				"ErrorFlag": "2",
			})
			return
		}

		notificationDoc := client.Collection(firstlogin.UserID).Doc("notification")
		getvalue, getDocError := notificationDoc.Get(context.Background())

		if getDocError != nil {
			log.Println("update/firstlogin notification欄位找不到", getDocError.Error())
			ctx.JSON(http.StatusOK, H{
				"ErrorMsg":  "找不到文件",
				"ErrorFlag": "2",
			})
			return
		}

		fcmValue, getFcmValueError := getvalue.DataAt("FCM")

		if getFcmValueError != nil {
			log.Println("update/firstlogin ", "找不到FCM屬性")
			ctx.JSON(http.StatusOK, H{
				"ErrorMsg":  "找不到FCM屬性",
				"ErrorFlag": "2",
			})
			return
		}

		log.Println("update/firstlogin ", "成功")
		ctx.JSON(http.StatusOK, H{
			"ErrorMsg":  "",
			"ErrorFlag": "0",
			"Data":      fcmValue,
		})
	})

	server.POST("/notification/fcm", func(ctx *Context) {
		authHeader := ctx.Req.Header.Get("FCMHeader")
		if authHeader != fcmHeader {
			ctx.JSON(http.StatusUnauthorized, H{
				"ErrorMsg":  "請設置header",
				"ErrorFlag": "2",
			})
			return
		}
		cctx := context.Background()
		collections := client.Collections(cctx)
		for {
			collectionRef, err := collections.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Fatalf("Failed to get collection: %v", err)
			}

			log.Printf("Collection: %s", collectionRef.ID)
			collection := client.Collection(collectionRef.ID)

			log.Printf("取得notification資料")
			notificationDoc := collection.Doc("notification")
			getvalue, getDocError := notificationDoc.Get(context.Background())
			if getDocError != nil {
				log.Printf("取得 notfication時 錯誤")
				continue
			}

			var nyData model.FcmInfo
			log.Printf("轉換notification的值")
			fcmError := getvalue.DataTo(&nyData)
			if fcmError != nil {
				log.Printf("轉換notification錯誤", fcmError)
			}
			if nyData.FcmValue {
				log.Printf("開始傳送fcm token 為", nyData.FCMToken)
				taskInfo := getTaskCount(collection)
				if hasIncompleteTesk(taskInfo) {
					//有代辦事項時傳送
					hasIncompleteTodos(nyData.FCMToken, taskInfo.InCompleteCount, taskInfo.TotalCount)
				} else {
					// 沒有訊息時傳送
					getNoToDoListMessage(nyData.FCMToken)
				}
			}

		}

	})

	server.Handle(w, r)
}

func getNoToDoListMessage(FCMToken string) {
	data := map[string]interface{}{
		"to": FCMToken,
		"notification": map[string]string{
			"body":  "點擊通知新增事項",
			"title": "美好的一天開始 8點新增事項",
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("getNoToDoListMessage轉換JSON錯誤:", err)
		return
	}
	sendNotficationToUser(jsonData, FCMToken)
}

func hasIncompleteTodos(FCMToken string, inCompleteCount int, totalCount int) {
	fmt.Println("未完成數 :" + strconv.Itoa(inCompleteCount) + " 總數" + strconv.Itoa(totalCount))
	var successRate float32 = 0
	if float32(totalCount)-float32(inCompleteCount) > 0 {
		completedCount := totalCount - inCompleteCount
		successRate = float32(completedCount) / float32(totalCount)
	}
	data := map[string]interface{}{
		"to": FCMToken,
		"notification": map[string]string{
			"body":  fmt.Sprintf("點擊查看未完成任務 完成率為%.1f%%", successRate*100),
			"title": "加油 目前還有" + strconv.Itoa(inCompleteCount) + "個未完成任務",
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("hasIncompleteTodos轉換JSON錯誤:", err)
		return
	}
	sendNotficationToUser(jsonData, FCMToken)
}

func hasIncompleteTesk(taskinfo model.TaskInfo) bool {
	return taskinfo.InCompleteCount > 0
}

func getTaskCount(collection *firestore.CollectionRef) model.TaskInfo {
	docsRefs, docsErr := collection.Documents(context.Background()).GetAll()
	inCompleteCount := 0
	totalCount := 0
	if docsErr != nil {
		return model.TaskInfo{
			InCompleteCount: inCompleteCount,
			TotalCount:      totalCount,
		}
	}

	for _, element := range docsRefs {
		if element.Ref.ID == "notification" {
			continue
		}

		fmt.Println("element : " + element.Ref.ID)
		todoEntry := element.Ref.Collection("todo-entries")
		todoEntryDocs, todoentryErr := todoEntry.Documents(context.Background()).GetAll()
		if todoentryErr != nil {
			continue
		}

		for _, todoEntryItemElement := range todoEntryDocs {
			todoEntryData := todoEntryItemElement.Data()
			isDone := todoEntryData["isDone"].(bool)
			totalCount += 1
			if !isDone {
				inCompleteCount += 1
			}
		}
	}
	fmt.Println("未完成 ：", inCompleteCount)
	fmt.Println("總數", totalCount)
	return model.TaskInfo{
		InCompleteCount: inCompleteCount,
		TotalCount:      totalCount,
	}
}

func sendNotficationToUser(message []byte, FCMToken string) {
	log.Printf("執行sendNotifcation方法 ", FCMToken)

	url := "https://fcm.googleapis.com/fcm/send"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(message))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", firebasefcmkey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("FCM推播錯誤", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("FCM推播結果", resp.Status)
}
