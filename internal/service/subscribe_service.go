package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// SubscribeService 订阅服务接口
type SubscribeService interface {
	GetUserByToken(token string) (*model.User, error)
	GetUserNodes(user *model.User) ([]model.Node, error)
	GenerateSubscribe(user *model.User, format string) (string, error)
}

type subscribeService struct {
	userRepo repository.UserRepository
	nodeRepo repository.NodeRepository
	planRepo repository.PlanRepository
	db       *gorm.DB
}

// NewSubscribeService 创建订阅服务
func NewSubscribeService(userRepo repository.UserRepository, nodeRepo repository.NodeRepository, planRepo repository.PlanRepository) SubscribeService {
	return &subscribeService{
		userRepo: userRepo,
		nodeRepo: nodeRepo,
		planRepo: planRepo,
		db:       database.Get(),
	}
}

// GetUserByToken 根据订阅token获取用户
func (s *subscribeService) GetUserByToken(token string) (*model.User, error) {
	user, err := s.userRepo.GetBySubscribeToken(token)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid subscribe token")
	}
	return user, nil
}

// GetUserNodes 获取用户可用的节点列表
// 按套餐的NodeGroup过滤可见节点，支持多组（逗号分隔），默认返回所有可见节点
func (s *subscribeService) GetUserNodes(user *model.User) ([]model.Node, error) {
	// 检查用户是否过期
	if user.HasExpired() {
		return []model.Node{}, nil
	}

	// 获取用户最近一笔已支付订单的套餐
	var latestOrder model.Order
	err := s.db.Where("user_id = ? AND status = ?", user.ID, model.OrderStatusPaid).
		Order("paid_at DESC").
		Preload("Plan").
		First(&latestOrder).Error

	if err != nil {
		// 没有已支付订单，返回所有可见节点（向后兼容）
		return s.nodeRepo.ListVisible()
	}

	plan := latestOrder.Plan
	if plan.NodeGroup == "" {
		// 套餐未指定节点组，返回所有可见节点
		return s.nodeRepo.ListVisible()
	}

	// 解析NodeGroup（支持逗号分隔的多个组ID）
	groupIDs := parseGroupIDs(plan.NodeGroup)
	if len(groupIDs) == 0 {
		return s.nodeRepo.ListVisible()
	}

	return s.nodeRepo.ListVisibleByGroups(groupIDs)
}

// parseGroupIDs 解析逗号分隔的组ID字符串
func parseGroupIDs(s string) []uint {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var ids []uint
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if id, err := strconv.ParseUint(p, 10, 32); err == nil && id > 0 {
			ids = append(ids, uint(id))
		}
	}
	return ids
}

// GenerateSubscribe 生成订阅内容
func (s *subscribeService) GenerateSubscribe(user *model.User, format string) (string, error) {
	nodes, err := s.GetUserNodes(user)
	if err != nil {
		return "", fmt.Errorf("failed to get user nodes: %w", err)
	}

	if len(nodes) == 0 {
		return "", errors.New("no available nodes")
	}

	switch strings.ToLower(format) {
	case "clash", "":
		return s.generateClashSubscribe(user, nodes)
	case "clashmeta", "mihomo":
		return s.generateClashMetaSubscribe(user, nodes)
	case "v2ray", "v2":
		return s.generateV2RaySubscribe(user, nodes)
	case "shadowrocket", "shadow-rocket", "sr":
		return s.generateShadowrocketSubscribe(user, nodes)
	case "sing-box", "singbox":
		return s.generateSingBoxSubscribe(user, nodes)
	case "surge":
		return s.generateSurgeSubscribe(user, nodes)
	case "surfboard":
		return s.generateSurfboardSubscribe(user, nodes)
	case "loon":
		return s.generateLoonSubscribe(user, nodes)
	case "quantumultx", "quanx", "qx":
		return s.generateQuantumultXSubscribe(user, nodes)
	case "stash":
		return s.generateStashSubscribe(user, nodes)
	default:
		return "", errors.New("unsupported subscribe format")
	}
}

// generateClashSubscribe 生成 Clash 格式订阅
func (s *subscribeService) generateClashSubscribe(user *model.User, nodes []model.Node) (string, error) {
	var proxies []map[string]interface{}

	for _, node := range nodes {
		if !node.IsVisible() {
			continue
		}
		info := node.ParseServerInfoConfig()
		proxy := s.buildClashProxy(user, node, info)
		if proxy != nil {
			proxies = append(proxies, proxy)
		}
	}

	if len(proxies) == 0 {
		return "", errors.New("no available nodes")
	}

	// 收集代理名称
	var proxyNames []string
	for _, p := range proxies {
		if name, ok := p["name"].(string); ok {
			proxyNames = append(proxyNames, name)
		}
	}

	config := map[string]interface{}{
		"port":       7890,
		"socks-port": 7891,
		"mode":       "rule",
		"log-level":  "info",
		"proxies":    proxies,
		"proxy-groups": []map[string]interface{}{
			{
				"name":    "PROXY",
				"type":    "select",
				"proxies": proxyNames,
			},
		},
		"rules": []string{
			"GEOIP,CN,DIRECT",
			"MATCH,PROXY",
		},
	}

	yamlContent, err := toYAML(config)
	if err != nil {
		return "", fmt.Errorf("failed to generate yaml: %w", err)
	}
	return yamlContent, nil
}

// buildClashProxy 根据节点类型和ServerInfo构建Clash代理配置
func (s *subscribeService) buildClashProxy(user *model.User, node model.Node, info *model.ServerInfoConfig) map[string]interface{} {
	proxy := map[string]interface{}{
		"name":   node.Name,
		"server": node.Address,
		"port":   node.Port,
	}

	network := info.Network
	if network == "" {
		network = "tcp"
	}

	switch node.Type {
	case model.NodeTypeVMess:
		proxy["type"] = "vmess"
		proxy["uuid"] = user.SubscribeToken
		proxy["alterId"] = info.AlterID
		cipher := info.Security
		if cipher == "" {
			cipher = "auto"
		}
		proxy["cipher"] = cipher
		proxy["network"] = network
		if info.TLS {
			proxy["tls"] = true
			if info.TLSServerName != "" {
				proxy["servername"] = info.TLSServerName
			}
		}
		if info.AllowInsecure {
			proxy["skip-cert-verify"] = true
		}
		s.addClashTransport(proxy, info, network)

	case model.NodeTypeVLESS:
		proxy["type"] = "vless"
		proxy["uuid"] = user.SubscribeToken
		proxy["network"] = network
		if info.TLS {
			proxy["tls"] = true
			if info.TLSServerName != "" {
				proxy["servername"] = info.TLSServerName
			}
		}
		if info.AllowInsecure {
			proxy["skip-cert-verify"] = true
		}
		if info.ALPN != "" {
			proxy["alpn"] = strings.Split(info.ALPN, ",")
		}
		s.addClashTransport(proxy, info, network)

	case model.NodeTypeTrojan:
		proxy["type"] = "trojan"
		proxy["password"] = user.SubscribeToken
		sni := info.TLSServerName
		if sni == "" {
			sni = node.Address
		}
		proxy["sni"] = sni
		if info.AllowInsecure {
			proxy["skip-cert-verify"] = true
		}
		if info.ALPN != "" {
			proxy["alpn"] = strings.Split(info.ALPN, ",")
		}

	case model.NodeTypeShadowsocks:
		proxy["type"] = "ss"
		cipher := info.Cipher
		if cipher == "" {
			cipher = "aes-256-gcm"
		}
		proxy["cipher"] = cipher
		ssPassword := info.Password
		if ssPassword == "" {
			ssPassword = user.SubscribeToken
		}
		proxy["password"] = ssPassword

	case model.NodeTypeHysteria2:
		proxy["type"] = "hysteria2"
		proxy["password"] = user.SubscribeToken
		sni := info.TLSServerName
		if sni == "" {
			sni = node.Address
		}
		proxy["sni"] = sni
		if info.ALPN != "" {
			proxy["alpn"] = strings.Split(info.ALPN, ",")
		} else {
			proxy["alpn"] = []string{"h3"}
		}
		if info.AllowInsecure {
			proxy["skip-cert-verify"] = true
		}
		if info.UpMbps > 0 {
			proxy["up"] = fmt.Sprintf("%d mbps", info.UpMbps)
		}
		if info.DownMbps > 0 {
			proxy["down"] = fmt.Sprintf("%d mbps", info.DownMbps)
		}
		if info.ObfsType != "" {
			proxy["obfs"] = info.ObfsType
			proxy["obfs-password"] = info.Obfs
		}

	case model.NodeTypeTUIC:
		proxy["type"] = "tuic"
		proxy["uuid"] = user.SubscribeToken
		proxy["password"] = user.SubscribeToken
		sni := info.TLSServerName
		if sni == "" {
			sni = node.Address
		}
		proxy["sni"] = sni
		if info.ALPN != "" {
			proxy["alpn"] = strings.Split(info.ALPN, ",")
		}
		if info.CongestionControl != "" {
			proxy["congestion-controller"] = info.CongestionControl
		}
		if info.AllowInsecure {
			proxy["skip-cert-verify"] = true
		}

	default:
		return nil // 不支持的类型跳过
	}

	return proxy
}

// addClashTransport 添加Clash传输层配置
func (s *subscribeService) addClashTransport(proxy map[string]interface{}, info *model.ServerInfoConfig, network string) {
	switch network {
	case "ws":
		if info.WSPath != "" {
			proxy["ws-path"] = info.WSPath
		}
		if info.WSHost != "" {
			proxy["ws-headers"] = map[string]string{"Host": info.WSHost}
		}
	case "grpc":
		if info.GrpcServiceName != "" {
			proxy["grpc-opts"] = map[string]interface{}{
				"grpc-service-name": info.GrpcServiceName,
			}
		}
	case "h2":
		if info.H2Path != "" {
			proxy["h2-opts"] = map[string]interface{}{
				"path": info.H2Path,
				"host": []string{info.H2Host},
			}
		}
	}
}

// generateV2RaySubscribe 生成 V2Ray 格式订阅（base64 编码的分享链接）
func (s *subscribeService) generateV2RaySubscribe(user *model.User, nodes []model.Node) (string, error) {
	var links []string

	for _, node := range nodes {
		if !node.IsVisible() {
			continue
		}
		info := node.ParseServerInfoConfig()
		link := s.buildShareLink(user, node, info)
		if link != "" {
			links = append(links, link)
		}
	}

	if len(links) == 0 {
		return "", errors.New("no available nodes")
	}

	content := strings.Join(links, "\n")
	return base64.StdEncoding.EncodeToString([]byte(content)), nil
}

// buildShareLink 根据节点类型和ServerInfo构建分享链接
func (s *subscribeService) buildShareLink(user *model.User, node model.Node, info *model.ServerInfoConfig) string {
	network := info.Network
	if network == "" {
		network = "tcp"
	}

	switch node.Type {
	case model.NodeTypeVMess:
		return s.buildVMessLink(user, node, info, network)
	case model.NodeTypeVLESS:
		return s.buildVLESSLink(user, node, info, network)
	case model.NodeTypeTrojan:
		return s.buildTrojanLink(user, node, info)
	case model.NodeTypeShadowsocks:
		return s.buildSSLink(user, node, info)
	case model.NodeTypeHysteria2:
		return s.buildHysteria2Link(user, node, info)
	case model.NodeTypeTUIC:
		return s.buildTUICLink(user, node, info)
	default:
		return ""
	}
}

// buildVMessLink 构建VMess分享链接
func (s *subscribeService) buildVMessLink(user *model.User, node model.Node, info *model.ServerInfoConfig, network string) string {
	cipher := info.Security
	if cipher == "" {
		cipher = "auto"
	}
	tls := ""
	if info.TLS {
		tls = "tls"
	}
	host := ""
	path := ""
	if network == "ws" {
		host = info.WSHost
		path = info.WSPath
	}

	vmessConfig := map[string]interface{}{
		"v":    "2",
		"ps":   node.Name,
		"add":  node.Address,
		"port": strconv.Itoa(node.Port),
		"id":   user.SubscribeToken,
		"aid":  strconv.Itoa(info.AlterID),
		"scy":  cipher,
		"net":  network,
		"type": "none",
		"host": host,
		"path": path,
		"tls":  tls,
		"sni":  info.TLSServerName,
	}
	jsonData, _ := json.Marshal(vmessConfig)
	return "vmess://" + base64.StdEncoding.EncodeToString(jsonData)
}

// buildVLESSLink 构建VLESS分享链接
func (s *subscribeService) buildVLESSLink(user *model.User, node model.Node, info *model.ServerInfoConfig, network string) string {
	params := url.Values{}
	params.Set("type", network)
	if info.TLS {
		params.Set("security", "tls")
	} else {
		params.Set("security", "none")
	}
	if info.TLSServerName != "" {
		params.Set("sni", info.TLSServerName)
	}
	if info.AllowInsecure {
		params.Set("allowInsecure", "1")
	}
	if info.ALPN != "" {
		params.Set("alpn", info.ALPN)
	}
	switch network {
	case "ws":
		if info.WSPath != "" {
			params.Set("path", info.WSPath)
		}
		if info.WSHost != "" {
			params.Set("host", info.WSHost)
		}
	case "grpc":
		if info.GrpcServiceName != "" {
			params.Set("serviceName", info.GrpcServiceName)
		}
	case "h2":
		if info.H2Path != "" {
			params.Set("path", info.H2Path)
		}
		if info.H2Host != "" {
			params.Set("host", info.H2Host)
		}
	}

	return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
		user.SubscribeToken, node.Address, node.Port, params.Encode(), url.PathEscape(node.Name))
}

// buildTrojanLink 构建Trojan分享链接
func (s *subscribeService) buildTrojanLink(user *model.User, node model.Node, info *model.ServerInfoConfig) string {
	params := url.Values{}
	sni := info.TLSServerName
	if sni == "" {
		sni = node.Address
	}
	params.Set("sni", sni)
	if info.AllowInsecure {
		params.Set("allowInsecure", "1")
	}
	if info.ALPN != "" {
		params.Set("alpn", info.ALPN)
	}

	return fmt.Sprintf("trojan://%s@%s:%d?%s#%s",
		url.QueryEscape(user.SubscribeToken), node.Address, node.Port, params.Encode(), url.PathEscape(node.Name))
}

// buildSSLink 构建Shadowsocks分享链接
func (s *subscribeService) buildSSLink(user *model.User, node model.Node, info *model.ServerInfoConfig) string {
	cipher := info.Cipher
	if cipher == "" {
		cipher = "aes-256-gcm"
	}
	password := info.Password
	if password == "" {
		password = user.SubscribeToken
	}
	ssURL := fmt.Sprintf("%s:%s@%s:%d", cipher, password, node.Address, node.Port)
	return "ss://" + base64.StdEncoding.EncodeToString([]byte(ssURL)) + "#" + url.PathEscape(node.Name)
}

// buildHysteria2Link 构建Hysteria2分享链接
func (s *subscribeService) buildHysteria2Link(user *model.User, node model.Node, info *model.ServerInfoConfig) string {
	params := url.Values{}
	sni := info.TLSServerName
	if sni == "" {
		sni = node.Address
	}
	params.Set("sni", sni)
	if info.ALPN != "" {
		params.Set("alpn", info.ALPN)
	} else {
		params.Set("alpn", "h3")
	}
	if info.AllowInsecure {
		params.Set("insecure", "1")
	}
	if info.ObfsType != "" {
		params.Set("obfs", info.ObfsType)
		params.Set("obfs-password", info.Obfs)
	}
	if info.UpMbps > 0 {
		params.Set("upmbps", strconv.Itoa(info.UpMbps))
	}
	if info.DownMbps > 0 {
		params.Set("downmbps", strconv.Itoa(info.DownMbps))
	}

	return fmt.Sprintf("hysteria2://%s@%s:%d?%s#%s",
		url.QueryEscape(user.SubscribeToken), node.Address, node.Port, params.Encode(), url.PathEscape(node.Name))
}

// buildTUICLink 构建TUIC分享链接
func (s *subscribeService) buildTUICLink(user *model.User, node model.Node, info *model.ServerInfoConfig) string {
	params := url.Values{}
	sni := info.TLSServerName
	if sni == "" {
		sni = node.Address
	}
	params.Set("sni", sni)
	if info.ALPN != "" {
		params.Set("alpn", info.ALPN)
	}
	if info.CongestionControl != "" {
		params.Set("congestion_control", info.CongestionControl)
	}
	if info.AllowInsecure {
		params.Set("allowInsecure", "1")
	}

	return fmt.Sprintf("tuic://%s:%s@%s:%d?%s#%s",
		user.SubscribeToken, user.SubscribeToken, node.Address, node.Port, params.Encode(), url.PathEscape(node.Name))
}

// generateShadowrocketSubscribe 生成 Shadowrocket 格式订阅（base64 编码的分享链接，带流量信息）
func (s *subscribeService) generateShadowrocketSubscribe(user *model.User, nodes []model.Node) (string, error) {
	var links []string

	for _, node := range nodes {
		if !node.IsVisible() {
			continue
		}
		info := node.ParseServerInfoConfig()
		link := s.buildShareLink(user, node, info)
		if link != "" {
			links = append(links, link)
		}
	}

	if len(links) == 0 {
		return "", errors.New("no available nodes")
	}

	// Shadowrocket 可以在头部添加流量信息
	var header string
	if user.TrafficLimit > 0 && user.ExpiredAt != nil {
		totalGB := fmt.Sprintf("%.2f", float64(user.TrafficLimit)/(1024*1024*1024))
		header = fmt.Sprintf("upload=0; download=0; total=%s; expire=%d",
			totalGB, user.ExpiredAt.Unix())
	}

	var content string
	if header != "" {
		content = header + "\n" + strings.Join(links, "\n")
	} else {
		content = strings.Join(links, "\n")
	}

	return base64.StdEncoding.EncodeToString([]byte(content)), nil
}

// generateSingBoxSubscribe 生成 Sing-box 格式订阅
func (s *subscribeService) generateSingBoxSubscribe(user *model.User, nodes []model.Node) (string, error) {
	var outbounds []map[string]interface{}

	for _, node := range nodes {
		if !node.IsVisible() {
			continue
		}
		info := node.ParseServerInfoConfig()
		outbound := s.buildSingBoxOutbound(user, node, info)
		if outbound != nil {
			outbounds = append(outbounds, outbound)
		}
	}

	if len(outbounds) == 0 {
		return "", errors.New("no available nodes")
	}

	// 构建 Sing-box 配置
	config := map[string]interface{}{
		"log": map[string]interface{}{
			"level": "info",
		},
		"inbounds": []map[string]interface{}{
			{
				"type":        "mixed",
				"tag":         "mixed-in",
				"listen":      "127.0.0.1",
				"listen_port": 7890,
			},
		},
		"outbounds": append([]map[string]interface{}{
			{
				"type": "selector",
				"tag":  "proxy",
				"outbounds": func() []string {
					tags := make([]string, len(nodes))
					for i, n := range nodes {
						tags[i] = n.Name
					}
					return tags
				}(),
				"default_outbound": nodes[0].Name,
			},
			{
				"type": "direct",
				"tag":  "direct",
			},
			{
				"type": "block",
				"tag":  "block",
			},
		}, outbounds...),
		"route": map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"geosite":  []string{"cn"},
					"outbound": "direct",
				},
				{
					"geoip":    []string{"cn"},
					"outbound": "direct",
				},
			},
			"final": "proxy",
		},
		"experimental": map[string]interface{}{
			"cache_file": map[string]interface{}{
				"enabled": true,
			},
		},
	}

	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to generate sing-box config: %w", err)
	}

	return string(jsonData), nil
}

// buildSingBoxOutbound 根据节点类型和ServerInfo构建Sing-box outbound配置
func (s *subscribeService) buildSingBoxOutbound(user *model.User, node model.Node, info *model.ServerInfoConfig) map[string]interface{} {
	network := info.Network
	if network == "" {
		network = "tcp"
	}
	sni := info.TLSServerName
	if sni == "" {
		sni = node.Address
	}

	switch node.Type {
	case model.NodeTypeVMess:
		outbound := map[string]interface{}{
			"type":        "vmess",
			"tag":         node.Name,
			"server":      node.Address,
			"server_port": node.Port,
			"uuid":        user.SubscribeToken,
			"security":    "auto",
			"alter_id":    info.AlterID,
		}
		cipher := info.Security
		if cipher != "" {
			outbound["security"] = cipher
		}
		s.addSingBoxTransport(outbound, info, network)
		if info.TLS {
			outbound["tls"] = s.buildSingBoxTLS(info, sni)
		}
		return outbound

	case model.NodeTypeVLESS:
		outbound := map[string]interface{}{
			"type":        "vless",
			"tag":         node.Name,
			"server":      node.Address,
			"server_port": node.Port,
			"uuid":        user.SubscribeToken,
		}
		s.addSingBoxTransport(outbound, info, network)
		if info.TLS {
			outbound["tls"] = s.buildSingBoxTLS(info, sni)
		}
		return outbound

	case model.NodeTypeTrojan:
		outbound := map[string]interface{}{
			"type":        "trojan",
			"tag":         node.Name,
			"server":      node.Address,
			"server_port": node.Port,
			"password":    user.SubscribeToken,
		}
		outbound["tls"] = s.buildSingBoxTLS(info, sni)
		return outbound

	case model.NodeTypeShadowsocks:
		cipher := info.Cipher
		if cipher == "" {
			cipher = "aes-256-gcm"
		}
		password := info.Password
		if password == "" {
			password = user.SubscribeToken
		}
		return map[string]interface{}{
			"type":        "shadowsocks",
			"tag":         node.Name,
			"server":      node.Address,
			"server_port": node.Port,
			"method":      cipher,
			"password":    password,
		}

	case model.NodeTypeHysteria2:
		outbound := map[string]interface{}{
			"type":        "hysteria2",
			"tag":         node.Name,
			"server":      node.Address,
			"server_port": node.Port,
			"password":    user.SubscribeToken,
		}
		tlsConfig := s.buildSingBoxTLS(info, sni)
		if info.ALPN != "" {
			tlsConfig["alpn"] = []string{info.ALPN}
		} else {
			tlsConfig["alpn"] = []string{"h3"}
		}
		outbound["tls"] = tlsConfig
		if info.UpMbps > 0 {
			outbound["up_mbps"] = info.UpMbps
		}
		if info.DownMbps > 0 {
			outbound["down_mbps"] = info.DownMbps
		}
		if info.ObfsType != "" {
			outbound["obfs"] = map[string]interface{}{
				"type":     info.ObfsType,
				"password": info.Obfs,
			}
		}
		return outbound

	case model.NodeTypeTUIC:
		outbound := map[string]interface{}{
			"type":        "tuic",
			"tag":         node.Name,
			"server":      node.Address,
			"server_port": node.Port,
			"uuid":        user.SubscribeToken,
			"password":    user.SubscribeToken,
		}
		tlsConfig := s.buildSingBoxTLS(info, sni)
		if info.ALPN != "" {
			tlsConfig["alpn"] = []string{info.ALPN}
		}
		outbound["tls"] = tlsConfig
		if info.CongestionControl != "" {
			outbound["congestion_control"] = info.CongestionControl
		}
		return outbound

	default:
		return nil
	}
}

// buildSingBoxTLS 构建Sing-box TLS配置
func (s *subscribeService) buildSingBoxTLS(info *model.ServerInfoConfig, sni string) map[string]interface{} {
	tls := map[string]interface{}{
		"enabled":     true,
		"server_name": sni,
	}
	if info.AllowInsecure {
		tls["insecure"] = true
	}
	return tls
}

// addSingBoxTransport 添加Sing-box传输层配置
func (s *subscribeService) addSingBoxTransport(outbound map[string]interface{}, info *model.ServerInfoConfig, network string) {
	switch network {
	case "ws":
		transport := map[string]interface{}{
			"type": "ws",
		}
		if info.WSPath != "" {
			transport["path"] = info.WSPath
		}
		if info.WSHost != "" {
			transport["headers"] = map[string]string{"Host": info.WSHost}
		}
		outbound["transport"] = transport
	case "grpc":
		transport := map[string]interface{}{
			"type": "grpc",
		}
		if info.GrpcServiceName != "" {
			transport["service_name"] = info.GrpcServiceName
		}
		outbound["transport"] = transport
	case "h2":
		transport := map[string]interface{}{
			"type": "http",
		}
		if info.H2Path != "" {
			transport["path"] = info.H2Path
		}
		if info.H2Host != "" {
			transport["host"] = []string{info.H2Host}
		}
		outbound["transport"] = transport
	}
}

// generateClashMetaSubscribe 生成 ClashMeta (mihomo) 格式订阅
func (s *subscribeService) generateClashMetaSubscribe(user *model.User, nodes []model.Node) (string, error) {
	var proxies []map[string]interface{}

	for _, node := range nodes {
		if !node.IsVisible() {
			continue
		}
		info := node.ParseServerInfoConfig()
		proxy := s.buildClashProxy(user, node, info)
		if proxy != nil {
			// ClashMeta 支持更多协议特性
			proxies = append(proxies, proxy)
		}
	}

	if len(proxies) == 0 {
		return "", errors.New("no available nodes")
	}

	var proxyNames []string
	for _, p := range proxies {
		if name, ok := p["name"].(string); ok {
			proxyNames = append(proxyNames, name)
		}
	}

	config := map[string]interface{}{
		"port":          7890,
		"socks-port":    7891,
		"mode":          "rule",
		"log-level":     "info",
		"proxies":       proxies,
		"proxy-groups": []map[string]interface{}{
			{
				"name":    "PROXY",
				"type":    "select",
				"proxies": proxyNames,
			},
		},
		"rules": []string{
			"GEOIP,CN,DIRECT",
			"MATCH,PROXY",
		},
	}

	return toYAML(config)
}

// generateSurgeSubscribe 生成 Surge 格式订阅
func (s *subscribeService) generateSurgeSubscribe(user *model.User, nodes []model.Node) (string, error) {
	var builder strings.Builder
	builder.WriteString("[General]\n")
	builder.WriteString("loglevel = notify\n")
	builder.WriteString("skip-proxy = 127.0.0.1,192.168.0.0/16,10.0.0.0/8,172.16.0.0/12,100.64.0.0/10,localhost,*.local\n")
	builder.WriteString("\n[Proxy]\n")

	var proxyNames []string
	for _, node := range nodes {
		if !node.IsVisible() {
			continue
		}
		info := node.ParseServerInfoConfig()
		proxyLine := s.buildSurgeProxyLine(user, node, info)
		if proxyLine != "" {
			builder.WriteString(proxyLine + "\n")
			proxyNames = append(proxyNames, node.Name)
		}
	}

	builder.WriteString("\n[Proxy Group]\n")
	builder.WriteString("PROXY = select,")
	builder.WriteString(strings.Join(proxyNames, ","))
	builder.WriteString("\n")

	builder.WriteString("\n[Rule]\n")
	builder.WriteString("GEOIP,CN,DIRECT\n")
	builder.WriteString("FINAL,PROXY\n")

	return builder.String(), nil
}

// buildSurgeProxyLine 构建 Surge 代理行
func (s *subscribeService) buildSurgeProxyLine(user *model.User, node model.Node, info *model.ServerInfoConfig) string {
	network := info.Network
	if network == "" {
		network = "tcp"
	}

	switch node.Type {
	case model.NodeTypeVMess:
		return fmt.Sprintf("%s = vmess, %s, %d, uuid=%s, tls=%v",
			node.Name, node.Address, node.Port, user.SubscribeToken, info.TLS)
	case model.NodeTypeVLESS:
		return fmt.Sprintf("%s = vless, %s, %d, uuid=%s, tls=%v",
			node.Name, node.Address, node.Port, user.SubscribeToken, info.TLS)
	case model.NodeTypeTrojan:
		return fmt.Sprintf("%s = trojan, %s, %d, password=%s",
			node.Name, node.Address, node.Port, user.SubscribeToken)
	case model.NodeTypeShadowsocks:
		cipher := info.Cipher
		if cipher == "" {
			cipher = "aes-256-gcm"
		}
		password := info.Password
		if password == "" {
			password = user.SubscribeToken
		}
		return fmt.Sprintf("%s = ss, %s, %d, method=%s, password=%s",
			node.Name, node.Address, node.Port, cipher, password)
	default:
		return ""
	}
}

// generateSurfboardSubscribe 生成 Surfboard 格式订阅
func (s *subscribeService) generateSurfboardSubscribe(user *model.User, nodes []model.Node) (string, error) {
	var builder strings.Builder
	builder.WriteString("[General]\n")
	builder.WriteString("loglevel = notify\n")
	builder.WriteString("\n[Proxy]\n")

	var proxyNames []string
	for _, node := range nodes {
		if !node.IsVisible() {
			continue
		}
		info := node.ParseServerInfoConfig()
		proxyLine := s.buildSurgeProxyLine(user, node, info) // Surfboard 使用类似 Surge 的格式
		if proxyLine != "" {
			builder.WriteString(proxyLine + "\n")
			proxyNames = append(proxyNames, node.Name)
		}
	}

	builder.WriteString("\n[Proxy Group]\n")
	builder.WriteString("PROXY = select,")
	builder.WriteString(strings.Join(proxyNames, ","))
	builder.WriteString("\n")

	builder.WriteString("\n[Rule]\n")
	builder.WriteString("GEOIP,CN,DIRECT\n")
	builder.WriteString("FINAL,PROXY\n")

	return builder.String(), nil
}

// generateLoonSubscribe 生成 Loon 格式订阅
func (s *subscribeService) generateLoonSubscribe(user *model.User, nodes []model.Node) (string, error) {
	var builder strings.Builder
	builder.WriteString("[General]\n")
	builder.WriteString("skip-proxy = 127.0.0.1,192.168.0.0/16,10.0.0.0/8,172.16.0.0/12\n")
	builder.WriteString("\n[Proxy]\n")

	var proxyNames []string
	for _, node := range nodes {
		if !node.IsVisible() {
			continue
		}
		info := node.ParseServerInfoConfig()
		proxyLine := s.buildLoonProxyLine(user, node, info)
		if proxyLine != "" {
			builder.WriteString(proxyLine + "\n")
			proxyNames = append(proxyNames, node.Name)
		}
	}

	builder.WriteString("\n[Proxy Group]\n")
	builder.WriteString("PROXY = select,")
	builder.WriteString(strings.Join(proxyNames, ","))
	builder.WriteString("\n")

	builder.WriteString("\n[Rule]\n")
	builder.WriteString("GEOIP,CN,DIRECT\n")
	builder.WriteString("FINAL,PROXY\n")

	return builder.String(), nil
}

// buildLoonProxyLine 构建 Loon 代理行
func (s *subscribeService) buildLoonProxyLine(user *model.User, node model.Node, info *model.ServerInfoConfig) string {
	switch node.Type {
	case model.NodeTypeVMess:
		return fmt.Sprintf("%s = vmess, %s, %d, uuid=%s, tls=%v",
			node.Name, node.Address, node.Port, user.SubscribeToken, info.TLS)
	case model.NodeTypeTrojan:
		return fmt.Sprintf("%s = trojan, %s, %d, password=%s",
			node.Name, node.Address, node.Port, user.SubscribeToken)
	case model.NodeTypeShadowsocks:
		cipher := info.Cipher
		if cipher == "" {
			cipher = "aes-256-gcm"
		}
		password := info.Password
		if password == "" {
			password = user.SubscribeToken
		}
		return fmt.Sprintf("%s = ss, %s, %d, method=%s, password=%s",
			node.Name, node.Address, node.Port, cipher, password)
	default:
		return ""
	}
}

// generateQuantumultXSubscribe 生成 QuantumultX 格式订阅
func (s *subscribeService) generateQuantumultXSubscribe(user *model.User, nodes []model.Node) (string, error) {
	var builder strings.Builder
	builder.WriteString("[server_remote]\n")

	for _, node := range nodes {
		if !node.IsVisible() {
			continue
		}
		info := node.ParseServerInfoConfig()
		link := s.buildShareLink(user, node, info)
		if link != "" {
			builder.WriteString(fmt.Sprintf("%s, tag=%s\n", link, node.Name))
		}
	}

	return builder.String(), nil
}

// generateStashSubscribe 生成 Stash 格式订阅
func (s *subscribeService) generateStashSubscribe(user *model.User, nodes []model.Node) (string, error) {
	// Stash 使用与 ClashMeta 类似的格式
	return s.generateClashMetaSubscribe(user, nodes)
}

// toYAML 简单的 YAML 生成函数（避免引入额外依赖）
// 注意：这是一个简化版本，仅用于生成 Clash 配置
func toYAML(data map[string]interface{}) (string, error) {
	var builder strings.Builder
	writeYAMLMap(&builder, data, 0)
	return builder.String(), nil
}

func writeYAMLMap(builder *strings.Builder, data map[string]interface{}, indent int) {
	prefix := strings.Repeat("  ", indent)
	for key, value := range data {
		switch v := value.(type) {
		case string:
			if strings.Contains(v, "\n") || strings.Contains(v, ":") || strings.Contains(v, "#") {
				builder.WriteString(fmt.Sprintf("%s%s: %q\n", prefix, key, v))
			} else {
				builder.WriteString(fmt.Sprintf("%s%s: %s\n", prefix, key, v))
			}
		case int, int64, float64:
			builder.WriteString(fmt.Sprintf("%s%s: %v\n", prefix, key, v))
		case bool:
			builder.WriteString(fmt.Sprintf("%s%s: %t\n", prefix, key, v))
		case nil:
			builder.WriteString(fmt.Sprintf("%s%s: null\n", prefix, key))
		case []interface{}:
			builder.WriteString(fmt.Sprintf("%s%s:\n", prefix, key))
			for _, item := range v {
				writeYAMLArrayItem(builder, item, indent+1)
			}
		case []string:
			builder.WriteString(fmt.Sprintf("%s%s:\n", prefix, key))
			for _, item := range v {
				builder.WriteString(fmt.Sprintf("%s  - %s\n", prefix, item))
			}
		case map[string]interface{}:
			builder.WriteString(fmt.Sprintf("%s%s:\n", prefix, key))
			writeYAMLMap(builder, v, indent+1)
		default:
			builder.WriteString(fmt.Sprintf("%s%s: %v\n", prefix, key, v))
		}
	}
}

func writeYAMLArrayItem(builder *strings.Builder, item interface{}, indent int) {
	prefix := strings.Repeat("  ", indent)
	switch v := item.(type) {
	case map[string]interface{}:
		first := true
		for key, value := range v {
			if first {
				builder.WriteString(fmt.Sprintf("%s- %s: ", prefix, key))
				writeYAMLValueInline(builder, value)
				first = false
			} else {
				builder.WriteString(fmt.Sprintf("%s  %s: ", prefix, key))
				writeYAMLValueInline(builder, value)
			}
		}
	default:
		builder.WriteString(fmt.Sprintf("%s- %v\n", prefix, v))
	}
}

func writeYAMLValueInline(builder *strings.Builder, value interface{}) {
	switch v := value.(type) {
	case string:
		if strings.Contains(v, "\n") || strings.Contains(v, ":") || strings.Contains(v, "#") {
			builder.WriteString(fmt.Sprintf("%q\n", v))
		} else {
			builder.WriteString(fmt.Sprintf("%s\n", v))
		}
	case []string:
		builder.WriteString("\n")
	case []interface{}:
		builder.WriteString("\n")
	default:
		builder.WriteString(fmt.Sprintf("%v\n", v))
	}
}
