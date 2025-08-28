package controller

import (
	"database/sql"
	"errors"
	"github.com/Arkine2054/l0/internal/kafka"
	"github.com/Arkine2054/l0/internal/logic"
	"github.com/gin-gonic/gin"
	"net/http"
)

type controller struct {
	logic    logic.Logic
	producer *kafka.Producer
}

type Controller interface {
	Index(ctx *gin.Context)
	GetOrder(ctx *gin.Context)
}

func NewController(logic logic.Logic, producer *kafka.Producer) Controller {
	return &controller{
		logic:    logic,
		producer: producer,
	}
}

func (c *controller) Index(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "index.html", gin.H{})
}

func (c *controller) GetOrder(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Введите ID заказа"})
		return
	}

	order, err := c.logic.GetOrder(ctx.Request.Context(), id)
	switch {
	case err == nil:
	case errors.Is(err, sql.ErrNoRows):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Заказ не найден"})
		return
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	ctx.JSON(http.StatusOK, order)
}
