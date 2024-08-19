package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/robfig/cron/v3"
)

type Request struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
}

func main() {

	go func() {
		c := cron.New()
		c.AddFunc("@every 15m", func() {
			cmd := exec.Command("sudo", "systemctl", "restart", "ipsec")
			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
			}
		})
		c.Start()
		select {}
	}()

	app := fiber.New()
	app.Use(cors.New())
	app.Post("/api", func(c *fiber.Ctx) error {
		var req Request
		err := c.BodyParser(&req)
		if err != nil {
			return err
		}

		var body []byte
		if req.Body != nil {
			body, _ = json.Marshal(req.Body)
		}

		httpReq, err := http.NewRequest("POST", req.URL, bytes.NewBuffer(body))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create request"})
		}

		if req.Headers != nil {
			for key, value := range req.Headers {
				httpReq.Header.Set(key, value)
			}
		}

		client := &http.Client{}
		resp, err := client.Do(httpReq)
		if err != nil {
			fmt.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Request failed"})
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read response"})
		}

		return c.Status(resp.StatusCode).Send(respBody)

	})
	app.Listen(":3021")
}
