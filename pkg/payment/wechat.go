package payment

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

// WechatConfig 微信支付配置
type WechatConfig struct {
	MchID      string // 商户号
	APIKey     string // API密钥
	CertPath   string // 证书路径
	KeyPath    string // 密钥路径
	NotifyURL  string // 异步通知地址
	Gateway    string // 网关地址
}

// WechatClient 微信支付客户端
type WechatClient struct {
	config     *WechatConfig
	privateKey *rsa.PrivateKey
}

// NewWechatClient 创建微信支付客户端
func NewWechatClient(config *WechatConfig) (*WechatClient, error) {
	client := &WechatClient{
		config: config,
	}

	// 设置默认网关
	if client.config.Gateway == "" {
		client.config.Gateway = "https://api.mch.weixin.qq.com"
	}

	// 解析私钥
	if config.KeyPath != "" {
		if err := client.parsePrivateKey(); err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	return client, nil
}

// parsePrivateKey 解析私钥
func (c *WechatClient) parsePrivateKey() error {
	keyData, err := readFile(c.config.KeyPath)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return err
		}
	}

	c.privateKey = key.(*rsa.PrivateKey)
	return nil
}

// readFile 读取文件
func readFile(path string) ([]byte, error) {
	data, err := io.ReadAll(nil) // 这里需要实际读取文件
	if err != nil {
		return nil, err
	}
	return data, nil
}

// WechatUnifiedOrderRequest 统一下单请求
type WechatUnifiedOrderRequest struct {
	OutTradeNo  string `xml:"out_trade_no"`  // 商户订单号
	TotalFee    int    `xml:"total_fee"`      // 订单金额（分）
	Body        string `xml:"body"`           // 商品描述
	TradeType   string `xml:"trade_type"`     // 交易类型
	OpenID      string `xml:"openid"`         // 用户标识（JSAPI支付时必填）
	NotifyURL   string `xml:"notify_url"`     // 异步通知地址
}

// WechatUnifiedOrderResponse 统一下单响应
type WechatUnifiedOrderResponse struct {
	ReturnCode string `xml:"return_code"`
	ReturnMsg  string `xml:"return_msg"`
	ResultCode string `xml:"result_code"`
	ErrCode    string `xml:"err_code"`
	ErrCodeDes string `xml:"err_code_des"`
	PrepayID   string `xml:"prepay_id"`
	CodeURL    string `xml:"code_url"`
}

// UnifiedOrder 统一下单
func (c *WechatClient) UnifiedOrder(req *WechatUnifiedOrderRequest) (*WechatUnifiedOrderResponse, error) {
	params := map[string]string{
		"appid":            c.config.MchID,
		"mch_id":           c.config.MchID,
		"nonce_str":        generateNonceStr(),
		"body":             req.Body,
		"out_trade_no":     req.OutTradeNo,
		"total_fee":        fmt.Sprintf("%d", req.TotalFee),
		"trade_type":       req.TradeType,
		"notify_url":       req.NotifyURL,
	}

	if req.OpenID != "" {
		params["openid"] = req.OpenID
	}

	// 签名
	params["sign"] = c.sign(params)

	// 构建XML
	xmlData := mapToXML(params)

	// 发送请求
	resp, err := http.Post(c.config.Gateway+"/pay/unifiedorder", "application/xml", strings.NewReader(xmlData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result WechatUnifiedOrderResponse
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.ReturnCode != "SUCCESS" {
		return nil, fmt.Errorf("wechat error: %s", result.ReturnMsg)
	}

	return &result, nil
}

// WechatOrderQueryRequest 订单查询请求
type WechatOrderQueryRequest struct {
	OutTradeNo string `xml:"out_trade_no"` // 商户订单号
	TransactionID string `xml:"transaction_id"` // 微信支付订单号
}

// WechatOrderQueryResponse 订单查询响应
type WechatOrderQueryResponse struct {
	ReturnCode    string `xml:"return_code"`
	ReturnMsg     string `xml:"return_msg"`
	ResultCode    string `xml:"result_code"`
	ErrCode       string `xml:"err_code"`
	ErrCodeDes    string `xml:"err_code_des"`
	TradeState    string `xml:"trade_state"`
	TradeStateDesc string `xml:"trade_state_desc"`
	TransactionID string `xml:"transaction_id"`
	OutTradeNo    string `xml:"out_trade_no"`
	TotalFee      int    `xml:"total_fee"`
	TimeEnd       string `xml:"time_end"`
}

// OrderQuery 订单查询
func (c *WechatClient) OrderQuery(req *WechatOrderQueryRequest) (*WechatOrderQueryResponse, error) {
	params := map[string]string{
		"appid":     c.config.MchID,
		"mch_id":    c.config.MchID,
		"nonce_str": generateNonceStr(),
	}

	if req.OutTradeNo != "" {
		params["out_trade_no"] = req.OutTradeNo
	}
	if req.TransactionID != "" {
		params["transaction_id"] = req.TransactionID
	}

	// 签名
	params["sign"] = c.sign(params)

	// 构建XML
	xmlData := mapToXML(params)

	// 发送请求
	resp, err := http.Post(c.config.Gateway+"/pay/orderquery", "application/xml", strings.NewReader(xmlData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result WechatOrderQueryResponse
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.ReturnCode != "SUCCESS" {
		return nil, fmt.Errorf("wechat error: %s", result.ReturnMsg)
	}

	return &result, nil
}

// WechatNotifyRequest 微信支付异步通知请求
type WechatNotifyRequest struct {
	ReturnCode    string `xml:"return_code"`
	ReturnMsg     string `xml:"return_msg"`
	ResultCode    string `xml:"result_code"`
	ErrCode       string `xml:"err_code"`
	ErrCodeDes    string `xml:"err_code_des"`
	AppID         string `xml:"appid"`
	MchID         string `xml:"mch_id"`
	NonceStr      string `xml:"nonce_str"`
	Sign          string `xml:"sign"`
	OpenID        string `xml:"openid"`
	TradeType     string `xml:"trade_type"`
	BankType      string `xml:"bank_type"`
	SettlementTotalFee int `xml:"settlement_total_fee"`
	CashFee       int    `xml:"cash_fee"`
	TransactionID string `xml:"transaction_id"`
	OutTradeNo    string `xml:"out_trade_no"`
	TotalFee      int    `xml:"total_fee"`
	TimeEnd       string `xml:"time_end"`
}

// VerifyNotify 验证异步通知
func (c *WechatClient) VerifyNotify(req *WechatNotifyRequest) (bool, error) {
	if req.ReturnCode != "SUCCESS" {
		return false, fmt.Errorf("notify failed: %s", req.ReturnMsg)
	}

	params := map[string]string{
		"return_code":    req.ReturnCode,
		"result_code":    req.ResultCode,
		"appid":          req.AppID,
		"mch_id":         req.MchID,
		"nonce_str":      req.NonceStr,
		"openid":         req.OpenID,
		"trade_type":     req.TradeType,
		"bank_type":      req.BankType,
		"transaction_id": req.TransactionID,
		"out_trade_no":   req.OutTradeNo,
		"time_end":       req.TimeEnd,
	}

	// 验证签名
	expectedSign := c.sign(params)
	return expectedSign == req.Sign, nil
}

// sign 签名
func (c *WechatClient) sign(params map[string]string) string {
	// 排序参数
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if v != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 构建待签名字符串
	var signStr strings.Builder
	for i, k := range keys {
		if i > 0 {
			signStr.WriteString("&")
		}
		signStr.WriteString(k)
		signStr.WriteString("=")
		signStr.WriteString(params[k])
	}
	signStr.WriteString("&key=")
	signStr.WriteString(c.config.APIKey)

	// MD5签名
	hash := sha256.Sum256([]byte(signStr.String()))
	return strings.ToUpper(fmt.Sprintf("%x", hash))
}

// generateNonceStr 生成随机字符串
func generateNonceStr() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// mapToXML 将map转换为XML
func mapToXML(params map[string]string) string {
	var buf bytes.Buffer
	buf.WriteString("<xml>")
	for k, v := range params {
		buf.WriteString(fmt.Sprintf("<%s>%s</%s>", k, v, k))
	}
	buf.WriteString("</xml>")
	return buf.String()
}

// TradeState 交易状态
const (
	WechatTradeStateSuccess = "SUCCESS"    // 支付成功
	WechatTradeStateRefund  = "REFUND"     // 转入退款
	WechatTradeStateNotPay  = "NOTPAY"     // 未支付
	WechatTradeStateClosed  = "CLOSED"     // 已关闭
	WechatTradeStateRevoked = "REVOKED"    // 已撤销
	WechatTradeStatePaying  = "USERPAYING" // 用户支付中
	WechatTradeStatePayError = "PAYERROR"  // 支付失败
)

// IsTradeSuccess 交易是否成功
func IsWechatTradeSuccess(state string) bool {
	return state == WechatTradeStateSuccess
}
