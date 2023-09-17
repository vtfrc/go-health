package journal

import (
	"go-health/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AddJournalEntry(db *gorm.DB, c *gin.Context) {
	username, _ := c.Get("username")
	var entry models.JournalEntry

	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := db.Where("username = ?", username.(string)).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	entry.UserID = user.ID

	if err := db.Create(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add journal entry"})
		return
	}

	c.JSON(http.StatusOK, entry)
}

func GetJournalEntries(db *gorm.DB, c *gin.Context) {
	username, _ := c.Get("username")

	var entries []models.JournalEntry
	if err := db.Where("UserID = ?", username.(uint)).Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve journal entries"})
		return
	}

	c.JSON(http.StatusOK, entries)
}

func GetJournalEntry(db *gorm.DB, c *gin.Context) {
	var entry models.JournalEntry
	id := c.Param("id")

	if err := db.First(&entry, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
		return
	}

	c.JSON(http.StatusOK, entry)
}

func UpdateJournalEntry(db *gorm.DB, c *gin.Context) {
	var entry models.JournalEntry
	id := c.Param("id")

	if err := db.First(&entry, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
		return
	}

	if err := c.BindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Save(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update entry"})
		return
	}

	c.JSON(http.StatusOK, entry)
}

func DeleteJournalEntry(db *gorm.DB, c *gin.Context) {
	username, _ := c.Get("username")
	id := c.Param("id")

	// Ensure the journal entry belongs to the authenticated user
	var entry models.JournalEntry
	if err := db.Where("ID = ? AND UserID = ?", id, username.(uint)).First(&entry).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Journal entry not found"})
		return
	}

	if err := db.Delete(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete journal entry"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Journal entry deleted successfully"})
}
