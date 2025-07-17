package main

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("erro ao carregar o arquivo .env: %v", err)
	}
}

type LoginBody struct {
	CPF string `json:"cpf"`
}

type Body struct {
	SessionID string `json:"sessionId"`
}

type UserResponse struct {
	Name     string `json:"name"`
	LastName string `json:"lastName"`
}

type GuardianResponse struct {
	UserResponse
	Document string `json:"document"`
	Email    string `json:"email"`
}

type StudentResponse struct {
	UserResponse
	BirthDate string `json:"birthDate"`
	RA        string `json:"ra"`
}

type OfferResponse struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Class string  `json:"class"`
	Price float64 `json:"price"`
}

func main() {
	r := gin.Default()

	rGroupAPI := r.Group("/api")

	addLoginRequest(rGroupAPI)

	rGroupAPIV1 := r.Group("/api/v1", func(context *gin.Context) {
		apiKey := context.GetHeader("X-API-KEY")
		println("APIKEY=", apiKey)
		if apiKey == "" || apiKey != os.Getenv("API_KEY") {
			context.JSON(401, gin.H{"message": "Invalid API Key"})
			context.Abort()

			return
		}
	})

	addGetOffersRequests(rGroupAPIV1)
	addGetGuardianDataRequest(rGroupAPIV1)

	rGroupAPIV1.POST("/", func(c *gin.Context) {
		bodyReq := Body{}
		if err := c.ShouldBindJSON(&bodyReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

			c.Abort()

			return
		}

		time.AfterFunc(5*time.Second, func() {
			req := makeRequest(bodyReq.SessionID)
			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}

			defer res.Body.Close()
			println(res.StatusCode)
		})

	})

	port := os.Getenv("PORT")
	err := r.Run(fmt.Sprintf(":%s", port))
	if err != nil {
		panic(err)
	}
}

func addLoginRequest(r *gin.RouterGroup) {
	r.POST("/auth", func(c *gin.Context) {
		bodyReq := LoginBody{}
		if err := c.ShouldBindJSON(&bodyReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

			c.Abort()

			return
		}

		if len(strings.TrimSpace(bodyReq.CPF)) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "cpf is required"})
			c.Abort()

			return
		}

		src := rand.NewSource(time.Now().UnixNano())
		r := rand.New(src)
		n := r.Intn(9000) + 1000
		s := fmt.Sprintf("%d", n)

		c.JSON(http.StatusOK, gin.H{"token": s, "email": "jo****e@hotmail.com"})

		return
	})
}

func addGetGuardianDataRequest(r *gin.RouterGroup) {
	r.GET("/:doc", func(c *gin.Context) {
		doc := c.Param("doc")

		if doc == "404" {
			c.JSON(404, gin.H{"message": "not found"})
			c.Abort()

			return
		}

		student := StudentResponse{
			UserResponse: UserResponse{
				Name:     "James",
				LastName: "Smith",
			},
			BirthDate: "20/01/2015",
			RA:        doc,
		}

		guardian := GuardianResponse{
			UserResponse: UserResponse{
				Name:     "John",
				LastName: "Doe",
			},
			Email:    "jo****e@hotmail.com",
			Document: "123.456.789-12",
		}

		c.JSON(http.StatusOK, gin.H{"data": gin.H{"student": student, "guardian": guardian}})
	})
}

func addGetOffersRequests(rGroup *gin.RouterGroup) {
	offers := []map[string]interface{}{
		{
			"id":    1,
			"name":  "Oferta 1",
			"class": "6o ano Manhã",
			"price": 32000.00,
		},
		{
			"id":    2,
			"name":  "Oferta 2",
			"class": "6o ano Tarde",
			"price": 36000.00,
		},
		{
			"id":    3,
			"name":  "Oferta 2 10% Desc",
			"class": "6o ano Tarde",
			"price": 29000.00,
		},
		{
			"id":    4,
			"name":  "Oferta 3",
			"class": "6o ano Noite",
			"price": 25000.00,
		},
	}

	rGroup.GET("/offers/:doc", func(c *gin.Context) {
		doc := c.Params.ByName("doc")
		if doc == "404" {
			c.JSON(404, gin.H{"message": "not found"})
			c.Abort()

			return
		}

		c.JSON(200, gin.H{"data": gin.H{"offers": offers}})
	})
}

func makeRequest(sessionID string) *http.Request {
	chatURL := os.Getenv("CHAT_URL")
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", chatURL, sessionID), bytes.NewBuffer(nil))
	if err != nil {
		log.Fatalf("Error creating POST request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	return req
}
