package payment

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// AlipayConfig 支付宝配置
type AlipayConfig struct {
	AppID      string // 应用ID
	PrivateKey string // 应用私钥
	PublicKey  string // 支付宝公钥
	NotifyURL  string // 异步通知地址
	ReturnURL  string // 同步跳转地址
	Gateway    string // 网关地址
}

// AlipayClient 支付宝客户端
type AlipayClient struct {
	config     *AlipayConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewAlipayClient 创建支付宝客户端
func NewAlipayClient(config *AlipayConfig) (*AlipayClient, error) {
	client := &AlipayClient{
		config: config,
	}

	// 解析私钥
	if err := client.parsePrivateKey(); err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// 解析公钥
	if err := client.parsePublicKey(); err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// 设置默认网关
	if client.config.Gateway == "" {
		client.config.Gateway = "https://openapi.alipay.com/gateway.do"
	}

	return client, nil
}

// parsePrivateKey 解析私钥
func (c *AlipayClient) parsePrivateKey() error {
	keyStr := c.config.PrivateKey
	if !strings.Contains(keyStr, "-----BEGIN") {
		keyStr = "-----BEGIN RSA PRIVATE KEY-----\n" + keyStr + "\n-----END RSA PRIVATE KEY-----"
	}

	block, _ := pem.Decode([]byte(keyStr))
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

// parsePublicKey 解析公钥
func (c *AlipayClient) parsePublicKey() error {
	keyStr := c.config.PublicKey
	if !strings.Contains(keyStr, "-----BEGIN") {
		keyStr = "-----BEGIN PUBLIC KEY-----\n" + keyStr + "\n-----END PUBLIC KEY-----"
	}

	block, _ := pem.Decode([]byte(keyStr))
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	c.publicKey = key.(*rsa.PublicKey)
	return nil
}

// AlipayTradePagePayRequest 支付宝网页支付请求
type AlipayTradePagePayRequest struct {
	OutTradeNo  string  `json:"out_trade_no"`  // 商户订单号
	TotalAmount float64 `json:"total_amount"`  // 订单总金额
	Subject     string  `json:"subject"`       // 订单标题
	ProductCode string  `json:"product_code"`  // 销售产品码
}

// TradePagePay 网页支付
func (c *AlipayClient) TradePagePay(req *AlipayTradePagePayRequest) (string, error) {
	params := map[string]string{
		"app_id":      c.config.AppID,
		"method":      "alipay.trade.page.pay",
		"format":      "JSON",
		"return_url":  c.config.ReturnURL,
		"notify_url":  c.config.NotifyURL,
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
	}

	// 构建业务参数
	bizContent := map[string]interface{}{
		"out_trade_no":  req.OutTradeNo,
		"total_amount":  fmt.Sprintf("%.2f", req.TotalAmount),
		"subject":       req.Subject,
		"product_code":  "FAST_INSTANT_TRADE_PAY",
	}

	bizContentJSON, err := json.Marshal(bizContent)
	if err != nil {
		return "", err
	}
	params["biz_content"] = string(bizContentJSON)

	// 签名
	sign, err := c.sign(params)
	if err != nil {
		return "", err
	}
	params["sign"] = sign

	// 构建跳转URL
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	return c.config.Gateway + "?" + values.Encode(), nil
}

// AlipayTradeQueryRequest 支付宝交易查询请求
type AlipayTradeQueryRequest struct {
	OutTradeNo string `json:"out_trade_no"` // 商户订单号
	TradeNo    string `json:"trade_no"`     // 支付宝交易号
}

// AlipayTradeQueryResponse 支付宝交易查询响应
type AlipayTradeQueryResponse struct {
	Code              string `json:"code"`
	Msg               string `json:"msg"`
	SubCode           string `json:"sub_code"`
	SubMsg            string `json:"sub_msg"`
	TradeNo           string `json:"trade_no"`
	OutTradeNo        string `json:"out_trade_no"`
	BuyerLogonID      string `json:"buyer_logon_id"`
	TradeStatus       string `json:"trade_status"`
	TotalAmount       string `json:"total_amount"`
	ReceiptAmount     string `json:"receipt_amount"`
	SendPayDate       string `json:"send_pay_date"`
	BuyerPayAmount    string `json:"buyer_pay_amount"`
	PointAmount       string `json:"point_amount"`
	InvoiceAmount     string `json:"invoice_amount"`
	StoreName         string `json:"store_name"`
	BuyerUserID       string `json:"buyer_user_id"`
}

// TradeQuery 交易查询
func (c *AlipayClient) TradeQuery(req *AlipayTradeQueryRequest) (*AlipayTradeQueryResponse, error) {
	params := map[string]string{
		"app_id":    c.config.AppID,
		"method":    "alipay.trade.query",
		"format":    "JSON",
		"charset":   "utf-8",
		"sign_type": "RSA2",
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"version":   "1.0",
	}

	// 构建业务参数
	bizContent := map[string]interface{}{}
	if req.OutTradeNo != "" {
		bizContent["out_trade_no"] = req.OutTradeNo
	}
	if req.TradeNo != "" {
		bizContent["trade_no"] = req.TradeNo
	}

	bizContentJSON, err := json.Marshal(bizContent)
	if err != nil {
		return nil, err
	}
	params["biz_content"] = string(bizContentJSON)

	// 签名
	sign, err := c.sign(params)
	if err != nil {
		return nil, err
	}
	params["sign"] = sign

	// 发送请求
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	resp, err := http.PostForm(c.config.Gateway, values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		AlipayTradeQueryResponse AlipayTradeQueryResponse `json:"alipay_trade_query_response"`
		Sign                     string                   `json:"sign"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result.AlipayTradeQueryResponse, nil
}

// AlipayNotifyRequest 支付宝异步通知请求
type AlipayNotifyRequest struct {
	NotifyTime     string `form:"notify_time"`
	NotifyType     string `form:"notify_type"`
	NotifyID       string `form:"notify_id"`
	AppID          string `form:"app_id"`
	Charset        string `form:"charset"`
	Version        string `form:"version"`
	SignType       string `form:"sign_type"`
	Sign           string `form:"sign"`
	TradeNo        string `form:"trade_no"`
	OutTradeNo     string `form:"out_trade_no"`
	OutBizNo       string `form:"out_biz_no"`
	BuyerID        string `form:"buyer_id"`
	BuyerLogonID   string `form:"buyer_logon_id"`
	SellerID       string `form:"seller_id"`
	SellerEmail    string `form:"seller_email"`
	TradeStatus    string `form:"trade_status"`
	TotalAmount    string `form:"total_amount"`
	ReceiptAmount  string `form:"receipt_amount"`
	BuyerPayAmount string `form:"buyer_pay_amount"`
	RefundFee      string `form:"refund_fee"`
	Subject        string `form:"subject"`
	Body           string `form:"body"`
	GmtCreate      string `form:"gmt_create"`
	GmtPayment     string `form:"gmt_payment"`
	GmtRefund      string `form:"gmt_refund"`
	GmtClose       string `form:"gmt_close"`
	FundBillList   string `form:"fund_bill_list"`
	PassbackParams string `form:"passback_params"`
}

// VerifyNotify 验证异步通知
func (c *AlipayClient) VerifyNotify(req *AlipayNotifyRequest) (bool, error) {
	// 构建待签名字符串
	params := map[string]string{
		"notify_time":     req.NotifyTime,
		"notify_type":     req.NotifyType,
		"notify_id":       req.NotifyID,
		"app_id":          req.AppID,
		"charset":         req.Charset,
		"version":         req.Version,
		"sign_type":       req.SignType,
		"trade_no":        req.TradeNo,
		"out_trade_no":    req.OutTradeNo,
		"out_biz_no":      req.OutBizNo,
		"buyer_id":        req.BuyerID,
		"buyer_logon_id":  req.BuyerLogonID,
		"seller_id":       req.SellerID,
		"seller_email":    req.SellerEmail,
		"trade_status":    req.TradeStatus,
		"total_amount":    req.TotalAmount,
		"receipt_amount":  req.ReceiptAmount,
		"buyer_pay_amount": req.BuyerPayAmount,
		"refund_fee":      req.RefundFee,
		"subject":         req.Subject,
		"body":            req.Body,
		"gmt_create":      req.GmtCreate,
		"gmt_payment":     req.GmtPayment,
		"gmt_refund":      req.GmtRefund,
		"gmt_close":       req.GmtClose,
		"fund_bill_list":  req.FundBillList,
		"passback_params": req.PassbackParams,
	}

	// 验证签名
	return c.verifySign(params, req.Sign)
}

// sign 签名
func (c *AlipayClient) sign(params map[string]string) (string, error) {
	// 排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建待签名字符串
	var signStr strings.Builder
	for i, k := range keys {
		if params[k] != "" {
			if i > 0 {
				signStr.WriteString("&")
			}
			signStr.WriteString(k)
			signStr.WriteString("=")
			signStr.WriteString(params[k])
		}
	}

	// SHA256 with RSA 签名
	hash := sha256.Sum256([]byte(signStr.String()))
	signature, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// verifySign 验证签名
func (c *AlipayClient) verifySign(params map[string]string, sign string) (bool, error) {
	// 排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建待验签字符串
	var signStr strings.Builder
	for i, k := range keys {
		if params[k] != "" {
			if i > 0 {
				signStr.WriteString("&")
			}
			signStr.WriteString(k)
			signStr.WriteString("=")
			signStr.WriteString(params[k])
		}
	}

	// 解码签名
	signBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false, err
	}

	// SHA256 with RSA 验签
	hash := sha256.Sum256([]byte(signStr.String()))
	err = rsa.VerifyPKCS1v15(c.publicKey, crypto.SHA256, hash[:], signBytes)
	return err == nil, nil
}

// TradeStatus 交易状态
const (
	TradeStatusWaitBuyerPay = "WAIT_BUYER_PAY" // 等待买家付款
	TradeStatusTradeClosed  = "TRADE_CLOSED"    // 交易关闭
	TradeStatusTradeSuccess = "TRADE_SUCCESS"   // 交易成功
	TradeStatusTradeFinish  = "TRADE_FINISH"    // 交易完成
)

// IsTradeSuccess 交易是否成功
func IsTradeSuccess(status string) bool {
	return status == TradeStatusTradeSuccess || status == TradeStatusTradeFinish
}
