package main

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"go-health/auth"
	"go-health/crud/journal"
	"go-health/db"
	"go-health/models"
	"go-health/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

var Db *gorm.DB

func main() {
	r := gin.Default()

	var err error

	Db, err = db.InitializeDB()

	if err != nil {
		panic("Failed to connect to the database.")
	}

	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	r.POST("/register", registerUser)
	r.POST("/login", loginUser)
	r.GET("/ws", auth.Authenticate, establishWSConnection)

	// journal endpoints
	r.POST("/journal", auth.Authenticate, func(c *gin.Context) {
		journal.AddJournalEntry(Db, c)
	})
	r.GET("/journal/entries", auth.Authenticate, func(c *gin.Context) {
		journal.GetJournalEntries(Db, c)
	})
	r.GET("/journal/:id", auth.Authenticate, func(c *gin.Context) {
		journal.GetJournalEntry(Db, c)
	})
	r.PUT("/journal/:id", auth.Authenticate, func(c *gin.Context) {
		journal.UpdateJournalEntry(Db, c)
	})
	r.DELETE("/journal/:id", auth.Authenticate, func(c *gin.Context) {
		journal.DeleteJournalEntry(Db, c)
	})

	go sendReminders()

	r.Run(":8080") // Default port
}

func registerUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Simple password hashing using sha256 (for demonstration purposes)
	h := sha256.New()
	h.Write([]byte(user.Password))
	user.Password = hex.EncodeToString(h.Sum(nil))

	if err := Db.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully registered"})
}

func loginUser(c *gin.Context) {
	var credentials models.User
	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := Db.Where("username = ?", credentials.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login failed"})
		return
	}

	// Compare the hashed password
	h := sha256.New()
	h.Write([]byte(credentials.Password))
	if user.Password != hex.EncodeToString(h.Sum(nil)) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login failed"})
		return
	}

	token, err := auth.GenerateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // In a real-world scenario, add a check here
	},
}

var wsManager = ws.NewConnectionManager() // Instantiate our manager

func establishWSConnection(c *gin.Context) {
	username, _ := c.Get("username")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to establish WebSocket connection"})
		return
	}

	// Add the connection to our manager
	wsManager.AddConnection(username.(string), conn)

	// This will listen for messages from the client to keep the connection alive
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				wsManager.RemoveConnection(username.(string))
				break
			}
		}
	}()
}

func sendReminders() {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			var reminders []models.Reminder
			Db.Where("Time <= ? AND Time > ?", now, now.Add(-1*time.Minute)).Find(&reminders)
			for _, reminder := range reminders {
				conn, ok := wsManager.GetConnection(string(rune(reminder.UserID)))
				if !ok {
					continue
				}

				err := conn.WriteMessage(websocket.TextMessage, []byte(reminder.Content))
				if err != nil {
					// remove the connection if it is no longer valid
					wsManager.RemoveConnection(string(rune(reminder.UserID)))
					continue
				}
			}
		}
	}
}
