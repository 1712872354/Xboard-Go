package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"xboard-go/config"

	"github.com/gin-gonic/gin"
)

// NodeAPIKeyAuth 是节点 REST API 的认证中间件。
// 认证方式与 gRPC AuthInterceptor 一致：
//   - token: 对比 config.Get().App.NodeAPIKey
//   - node_id: 从 query params 或 JSON body 中提取
//
// Xboard-Node GET 请求使用 query params 认证，POST 请求使用 JSON body 认证。
// 对于 POST 请求，body 会被完整读取并解析，原始 body 会重新放回供 handler 使用。
// 解析后的 map 存储在 gin.Context 的 "node_body" key 中，handler 可以直接使用。
func NodeAPIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		expectedKey := config.Get().App.NodeAPIKey
		if expectedKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "node_api_key not configured"})
			c.Abort()
			return
		}

		var token, nodeIDStr, machineIDStr string
		var bodyMap map[string]interface{}

		// 对于 POST/PUT 等有 body 的请求，读取并解析 body
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" && c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
				c.Abort()
				return
			}
			// 重新设置 body 供后续 handler 使用
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			if len(bodyBytes) > 0 {
				if err := json.Unmarshal(bodyBytes, &bodyMap); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
					c.Abort()
					return
				}
				if v, ok := bodyMap["token"].(string); ok {
					token = v
				}
				if v, ok := bodyMap["node_id"]; ok {
					switch val := v.(type) {
					case string:
						nodeIDStr = val
					case float64:
						nodeIDStr = strconv.FormatFloat(val, 'f', 0, 64)
					}
				}
				if v, ok := bodyMap["machine_id"]; ok {
					switch val := v.(type) {
					case string:
						machineIDStr = val
					case float64:
						machineIDStr = strconv.FormatFloat(val, 'f', 0, 64)
					}
				}
			}
		}

		// query params 优先（覆盖 body 中的值）
		if q := c.Query("token"); q != "" {
			token = q
		}
		if q := c.Query("node_id"); q != "" {
			nodeIDStr = q
		}
		if q := c.Query("machine_id"); q != "" {
			machineIDStr = q
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}
		if token != expectedKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// 支持 machine_id 或 node_id（至少需要一个）
		if nodeIDStr == "" && machineIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing node_id or machine_id"})
			c.Abort()
			return
		}

		// 存入 context
		if nodeIDStr != "" {
			nodeID, err := strconv.ParseUint(nodeIDStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node_id"})
				c.Abort()
				return
			}
			c.Set("node_id", uint32(nodeID))
		}
		if machineIDStr != "" {
			machineID, err := strconv.ParseUint(machineIDStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid machine_id"})
				c.Abort()
				return
			}
			c.Set("machine_id", uint32(machineID))
		}
		if bodyMap != nil {
			c.Set("node_body", bodyMap)
		}
		c.Next()
	}
}

// GetNodeID 从 gin.Context 中获取已认证的 node_id。
func GetNodeID(c *gin.Context) uint32 {
	v, _ := c.Get("node_id")
	id, _ := v.(uint32)
	return id
}

// GetMachineID 从 gin.Context 中获取已认证的 machine_id。
func GetMachineID(c *gin.Context) uint32 {
	v, _ := c.Get("machine_id")
	id, _ := v.(uint32)
	return id
}

// GetNodeBody 从 gin.Context 中获取已解析的请求 body。
func GetNodeBody(c *gin.Context) map[string]interface{} {
	v, _ := c.Get("node_body")
	m, _ := v.(map[string]interface{})
	return m
}
