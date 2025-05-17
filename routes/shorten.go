package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/anton-fuji/minilink/databases"
	"github.com/anton-fuji/minilink/helpers"
	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type request struct {
	URL         string `json:"url"`
	CustomShort string `json:"short"`
	ExpiryHours int    `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	Short           string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateResetAfter time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {
	// JSON パース
	body := new(request)
	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	// レートリミット
	quotaStr := os.Getenv("API_QUOTA")
	if quotaStr == "" {
		quotaStr = "15"
	}
	quota, _ := strconv.Atoi(quotaStr)
	resetMin := 30
	if v := os.Getenv("RATE_LIMIT_RESET"); v != "" {
		if x, err := strconv.Atoi(v); err == nil {
			resetMin = x
		}
	}
	resetDur := time.Duration(resetMin) * time.Minute

	rlDB := databases.CreateClient(1)
	defer rlDB.Close()

	countStr, err := rlDB.Get(databases.Ctx, c.IP()).Result()
	if err == redis.Nil {
		// 初回
		rlDB.Set(databases.Ctx, c.IP(), quota, resetDur)
	} else if n, _ := strconv.Atoi(countStr); n <= 0 {
		ttl, _ := rlDB.TTL(databases.Ctx, c.IP()).Result()
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error":            "Rate Limit Exceeded",
			"rate_limit_reset": ttl / time.Second,
		})
	}

	// URL検証
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}
	bodyURL := helpers.EnforceHTTP(body.URL)
	if !helpers.RemoveDomainError(bodyURL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Disallowed domain"})
	}

	key := body.CustomShort
	if key == "" {
		key = uuid.New().String()[:6]
	}

	// すでに存在しないか確認
	dataDB := databases.CreateClient(0)
	defer dataDB.Close()
	if ex, _ := dataDB.Exists(databases.Ctx, key).Result(); ex > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Short key already used"})
	}

	// 有効期限
	hours := body.ExpiryHours
	if hours <= 0 {
		hours = 24
	}
	expDur := time.Duration(hours) * time.Hour
	dataDB.Set(databases.Ctx, key, bodyURL, expDur)

	// カウントダウン
	newCount, _ := rlDB.Decr(databases.Ctx, c.IP()).Result()
	ttl, _ := rlDB.TTL(databases.Ctx, c.IP()).Result()

	// レスポンス
	resp := response{
		URL:             bodyURL,
		Short:           os.Getenv("DOMAIN") + "/" + key,
		Expiry:          expDur,
		XRateRemaining:  int(newCount),
		XRateResetAfter: ttl / time.Second,
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}
