package handler

import (
	"xboard-go/internal/middleware"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserServerHandler 用户端服务器处理器
type UserServerHandler struct {
	userServerService service.UserServerService
}

// NewUserServerHandler 创建用户端服务器处理器
func NewUserServerHandler(userServerService service.UserServerService) *UserServerHandler {
	return &UserServerHandler{
		userServerService: userServerService,
	}
}

// FetchServers 获取用户可用的服务器列表
// @Summary 获取服务器列表
// @Description 获取当前用户可用的服务器列表（根据套餐权限过滤）
// @Tags 用户-服务器
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=[]model.Node}
// @Router /api/v1/user/servers [get]
func (h *UserServerHandler) FetchServers(c *gin.Context) {
	userID := middleware.GetUserID(c)

	// 获取用户信息
	user, err := h.userServerService.GetUserByID(userID)
	if err != nil {
		response.InternalError(c, "获取用户信息失败")
		return
	}

	// 检查用户是否可用
	if !user.CanUseService() {
		response.Success(c, []interface{}{})
		return
	}

	// 获取用户可用的节点
	nodes, err := h.userServerService.GetUserNodes(user)
	if err != nil {
		response.InternalError(c, "获取服务器列表失败")
		return
	}

	// 转换为用户端响应格式
	var servers []gin.H
	for _, node := range nodes {
		if !node.IsVisible() {
			continue
		}

		server := gin.H{
			"id":          node.ID,
			"name":        node.Name,
			"type":        node.Type,
			"host":        node.Host,
			"port":        node.Port,
			"server_info": node.ServerInfo,
			"rate":        node.Rate,
			"tags":        node.Tags,
			"group_ids":   node.GroupIDs,
			"parent_id":   node.ParentID,
			"sort":        node.Sort,
			"show":        node.Show,
		}
		servers = append(servers, server)
	}

	if servers == nil {
		servers = []gin.H{}
	}

	response.Success(c, servers)
}
