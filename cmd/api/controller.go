package main

import (
	"fmt"
	"gin-subscription/internal/database"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// createSubscription creates new subscription
//
//	@Summary		creates new subscription
//	@Description	creates new subscription
//	@Tags			Subscription
//	@Accept			json
//	@Produce		json
//	@Param			subscription	body		database.Subscription	true	"Name of subscription provider"
//	@Success		201				{object}	database.Subscription
//	@Router			/api/v1/subscription [post]
func (app *application) createSubscription(c *gin.Context) {
	var subscription database.Subscription

	if err := c.ShouldBindJSON(&subscription); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := app.models.Subscriptions.Insert(&subscription)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription"})
		return
	}

	c.JSON(http.StatusCreated, subscription)
}

// getSubscription returns single subscription
//
//	@Summary		returns single subscription
//	@Description	returns single subscription
//	@Tags			Subscription
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Subscription id"
//	@Success		200	{object}	database.Subscription
//	@Router			/api/v1/subscription/{id} [get]
func (app *application) getSubscription(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	sub, err := app.models.Subscriptions.Get(id)
	fmt.Println(err)
	if sub == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retreive event"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// updateSubscription updates an existing subscription
//
//	@Summary		updates existing subscription
//	@Description	updates existing subscription
//	@Tags			Subscription
//	@Accept			json
//	@Produce		json
//	@Param			id				path		int						true	"Subscription id"
//	@Param			subscription	body		database.Subscription	true	"Subscription"
//	@Success		200				{object}	database.Subscription
//	@Router			/api/v1/subscription/{id} [put]
func (app *application) updateSubscription(c *gin.Context) {
	slog.Info("Method updateSubscription in controller", "id", c.Param("id"))

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	existingSub, err := app.models.Subscriptions.Get(id)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retreive subscription"})
		return
	}

	if existingSub == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	updatedSub := &database.Subscription{}

	if err := c.ShouldBindJSON(updatedSub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedSub.Id = id

	if err := app.models.Subscriptions.Update(updatedSub); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, updatedSub)
}

// deleteSubscription deletes an existing subscription
//
//	@Summary		deletes existing subscription
//	@Description	deletes existing subscription
//	@Tags			Subscription
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"Subscription id"
//	@Success		204
//	@Router			/api/v1/subscription/{id} [delete]
func (app *application) deleteSubscription(c *gin.Context) {
	slog.Info("Method deleteSubscription in controller", "id", c.Param("id"))

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
	}

	existingSub, err := app.models.Subscriptions.Get(id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retreive subscription"})
		return
	}

	if existingSub == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	if err := app.models.Subscriptions.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscription"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// listSubscription returns list of all subscriptions
//
//	@Summary		returns list of all subscriptions
//	@Description	returns list of all subscriptions
//	@Tags			Subscription
//	@Accept			json
//	@Produce		json
//	@Param			user_id			query	int		false	"filter for concrete user"
//	@Param			service_name	query	string	false	"filter for concrete service"
//	@Success		200
//	@Router			/api/v1/subscription [get]
func (app *application) listSubscription(c *gin.Context) {
	slog.Info("Method listSubscription in controller", "query_filter", c.Request.URL.Query())

	filter := make(map[string]string)
	if u := c.Query("user_id"); u != "" {
		_, err := strconv.Atoi(u)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter type"})
			return
		}
		filter["user_id"] = u
	}
	if s := c.Query("service_name"); s != "" {
		filter["service_name"] = s
	}

	events, err := app.models.Subscriptions.GetList(filter)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retreive subscriptions"})
		return
	}

	c.JSON(http.StatusOK, events)
}

// getPeriodPrice returns price of chosen subscription for period
//
//	@Summary		returns price of choosen subscription for period
//	@Description	requests period of time in path, format "mm-yyyy:{mm-yyyy}", where right side might be ommited and autoreplaced with time.Now()
//	@Description	query params 'user_id' and 'service_name' used as filter for request
//	@Tags			Subscription
//	@Accept			json
//	@Produce		json
//	@Param			period			path	string	true	"period"	example(07-2025:08-2025)
//	@Param			user_id			query	int		false	"filter for concrete user"
//	@Param			service_name	query	string	false	"filter for concrete service"
//	@Success		200
//	@Router			/api/v1/subscription/period-price/{period} [get]
func (app *application) getPeriodPrice(c *gin.Context) {
	slog.Info("Method getPeriodPrice in controller", "period", c.Param("period"), "query_filter", c.Request.URL.Query())

	periodSlice := strings.Split(c.Param("period"), ":")
	var periodTime []time.Time
	for _, d := range periodSlice {
		time, err := time.Parse("01-2006", d)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription period format"})
			return
		}
		periodTime = append(periodTime, time)
	}

	start := periodTime[0]
	end := time.Now()
	if len(periodTime) == 2 {
		end = periodTime[1]
	}

	filter := make(map[string]string)
	if u := c.Query("user_id"); u != "" {
		_, err := strconv.Atoi(u)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter type"})
			return
		}
		filter["user_id"] = u
	}
	if s := c.Query("service_name"); s != "" {
		filter["service_name"] = s
	}

	total, prices, err := app.models.Subscriptions.GetPrice(start, end, filter)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retreive price"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total price": total, "prices": prices})
}
